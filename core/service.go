package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ronaldalds/gorote-core-rsa/gorote"
	"gorm.io/gorm"
)

type servicer interface {
	health() (*gorote.Health, error)
	setCookie(*fiber.Ctx, string, string) error
	generateJwt(*User, string) (string, error)
	login(*login) (*User, error)
	users(...string) ([]User, error)
	roles(...string) ([]Role, error)
	permissions(...string) ([]Permission, error)
	createRole(*createRole) (*Role, error)
	createUser(*createUser) (*User, error)
	updateUser(*schemaUser, bool, bool) (*User, error)
	claims(jwt.Claims, string) error
}

func (s *appService) health() (*gorote.Health, error) {
	return gorote.HealthGorm(s.db())
}

func (s *appService) setCookie(ctx *fiber.Ctx, typeToken, token string) error {
	var expire time.Duration
	switch typeToken {
	case "access_token":
		expire = s.jwt().JwtExpireAccess
	case "refresh_token":
		expire = s.jwt().JwtExpireRefresh
	default:
		return fmt.Errorf("invalid token type")
	}
	domains := s.domain()
	if domains != "" {
		for domain := range strings.SplitSeq(domains, ",") {
			domain = strings.TrimSpace(domain)
			if domain == "" {
				continue
			}
			ctx.Cookie(&fiber.Cookie{
				Name:     typeToken,
				Value:    token,
				HTTPOnly: true,
				Secure:   true,
				SameSite: "None",
				Domain:   domain,
				Path:     "/",
				MaxAge:   int(expire.Seconds()),
			})
		}
	}
	return nil
}

func (s *appService) generateJwt(user *User, typeToken string) (string, error) {
	var permissions []string
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			permissions = append(permissions, permission.Code)
		}
	}
	var tenants []string
	for _, tenant := range user.Tenants {
		tenants = append(tenants, tenant.ID.String())
	}

	var expire time.Duration
	switch typeToken {
	case "access_token":
		expire = s.jwt().JwtExpireAccess
	case "refresh_token":
		expire = s.jwt().JwtExpireRefresh
	default:
		return "", fmt.Errorf("invalid token type")
	}

	token, err := gorote.GenerateJwtWithRSA(JwtClaims{
		IsSuperUser: user.IsSuperUser,
		Permissions: permissions,
		Tenants:     tenants,
		Type:        typeToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        user.ID.String(),
			Subject:   user.Email,
			Issuer:    "gorote-core",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
		},
	}, s.privateKeyRSA())
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *appService) login(req *login) (*User, error) {
	var user User
	result := s.db().
		Preload("Roles.Permissions").
		Preload("Tenants").
		Where("email = ?", req.Email).
		First(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to login: username or password is incorrect")
	}
	if !gorote.CheckPasswordHash(req.Password, user.Password) {
		return nil, fmt.Errorf("failed to login: username or password is incorrect")
	}
	if !user.Active {
		return nil, fmt.Errorf("failed to login: user is inactive")
	}
	return &user, nil
}

func (s *appService) users(ids ...string) ([]User, error) {
	var data []User
	if len(ids) == 0 {
		if err := s.db().
			Preload("Roles.Permissions").
			Preload("Tenants").
			Find(&data).Error; err != nil {
			return nil, fmt.Errorf("failed to query database list")
		}
		return data, nil
	}

	if err := s.db().
		Preload("Roles.Permissions").
		Preload("Tenants").
		Where("id IN ?", ids).
		Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to query database")
	}
	return data, nil
}

func (s *appService) tenants(ids ...string) ([]Tenant, error) {
	var data []Tenant
	if len(ids) == 0 {
		if err := s.db().
			Find(&data).Error; err != nil {
			return nil, fmt.Errorf("failed to query database")
		}
		return data, nil
	}
	if err := s.db().
		Where("id IN ?", ids).
		Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch tenants")
	}
	return data, nil
}

func (s *appService) roles(ids ...string) ([]Role, error) {
	var data []Role
	if len(ids) == 0 {
		if err := s.db().
			Preload("Permissions").
			Find(&data).Error; err != nil {
			return nil, fmt.Errorf("failed to query database")
		}
		return data, nil
	}
	if err := s.db().
		Preload("Permissions").
		Where("id IN ?", ids).
		Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch roles")
	}
	return data, nil
}

func (s *appService) permissions(ids ...string) ([]Permission, error) {
	var permissions []Permission
	if len(ids) == 0 {
		if err := s.db().
			Preload("Roles").
			Find(&permissions).Error; err != nil {
			return nil, fmt.Errorf("failed to query database")
		}
		return permissions, nil
	}
	if err := s.db().
		Preload("Roles").
		Where("id IN ?", ids).
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch permissions")
	}

	return permissions, nil
}

func (s *appService) createRole(req *createRole) (*Role, error) {
	var role Role
	permissions, err := s.permissions(req.Permissions...)
	if err != nil {
		return nil, fmt.Errorf("permission with ids does not exist")
	}

	role.Name = req.Name
	role.Description = req.Description
	role.Permissions = permissions
	role.Active = true

	if err := s.db().Create(&role).Error; err != nil {
		return nil, fmt.Errorf("failed to create role")
	}
	return &role, nil
}

func (s *appService) createUser(req *createUser) (*User, error) {
	roles, err := s.roles(req.Roles...)
	if err != nil {
		return nil, fmt.Errorf("group with ids does not exist")
	}

	user := User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
		Active:    req.Active,
		Phone1:    &req.Phone1,
		Phone2:    &req.Phone2,
	}

	if err := s.db().Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user")
	}

	if err := s.db().Model(&user).Association("Roles").Replace(roles); err != nil {
		return nil, fmt.Errorf("failed to set roles for user")
	}

	return &user, nil
}

func (s *appService) updateUser(req *schemaUser, editorSuper, editorPermission bool) (*User, error) {
	var user User

	err := s.db().Transaction(func(tx *gorm.DB) error {
		users, err := s.users(req.ID)
		if err != nil {
			return err
		}
		if len(users) == 0 {
			return fmt.Errorf("no users found")
		}
		user = users[0]

		user.FirstName = req.FirstName
		user.LastName = req.LastName
		user.Active = req.Active
		if editorSuper {
			user.IsSuperUser = req.IsSuperUser
		}
		user.Phone1 = &req.Phone1
		user.Phone2 = &req.Phone2

		if editorPermission || editorSuper {
			roles, err := s.roles(req.Roles...)
			if err != nil {
				return err
			}
			tenants, err := s.tenants(req.Tenants...)
			if err != nil {
				return err
			}
			user.Roles = roles
			user.Tenants = tenants
		}

		if err := tx.Model(&user).Updates(user).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		if editorPermission || editorSuper {
			if err := tx.Model(&user).Association("Roles").Replace(user.Roles); err != nil {
				return fmt.Errorf("failed to update roles: %w", err)
			}
			if err := tx.Model(&user).Association("Tenants").Replace(user.Tenants); err != nil {
				return fmt.Errorf("failed to update tenants: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *appService) claims(claims jwt.Claims, token string) error {
	if err := gorote.ValidateOrGetJWTRSA(claims, token, &s.privateKeyRSA().PublicKey); err != nil {
		return err
	}
	return nil
}
