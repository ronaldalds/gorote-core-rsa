package core

import (
	"fmt"

	"github.com/ronaldalds/gorote-core-rsa/gorote"
	"gorm.io/gorm"
)

func (s *Service) Health() (*gorote.Health, error) {
	return gorote.HealthGorm(s.DB)
}

func (s *Service) Login(req *Login) (*User, error) {
	var user User
	result := s.DB.
		Preload("Roles.Permissions").
		Preload("Tenants").
		Where("email = ?", req.Username).
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

func (s *Service) ListPermissions() ([]Permission, error) {
	var permissions []Permission
	if err := s.DB.
		Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to query database")
	}
	return permissions, nil
}

func (s *Service) GetPermissionsByIDs(ids []uint) ([]Permission, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no permission IDs provided")
	}

	var permissions []Permission
	if err := s.DB.Where("id IN ?", ids).Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch permissions: %s", err.Error())
	}

	if len(permissions) != len(ids) {
		return nil, fmt.Errorf("permissions not found for IDs")
	}

	return permissions, nil
}

func (s *Service) ListRoles() ([]Role, error) {
	var roles []Role
	result := s.DB.
		Preload("Permissions").
		Find(&roles)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query database: %w", result.Error)
	}
	return roles, nil
}

func (s *Service) CreateRole(req *CreateRole) (*Role, error) {
	var role Role
	permissions, err := s.GetPermissionsByIDs(req.Permissions)
	if err != nil {
		return nil, fmt.Errorf("permission with ids '%v' does not exist", req.Permissions)
	}

	role.Name = req.Name
	role.Description = req.Description
	role.Permissions = permissions
	role.Active = true

	if err := s.DB.Create(&role).Error; err != nil {
		return nil, fmt.Errorf("failed to create role: %s", err.Error())
	}
	return &role, nil
}

func (s *Service) GetRolesByIDs(ids []uint) ([]Role, error) {
	var roles []Role
	if len(ids) == 0 {
		return roles, nil
	}

	if err := s.DB.Where("id IN ?", ids).Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch roles: %s", err.Error())
	}

	if len(roles) != len(ids) {
		return nil, fmt.Errorf("roles not found for IDs")
	}

	return roles, nil
}

func (s *Service) GetUserByID(id uint) (*User, error) {
	var user User
	result := s.DB.
		Where("id = ?", id).
		Preload("Roles.Permissions").
		Preload("Tenants").
		First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no record found for id: %d", id)
		}
		return nil, fmt.Errorf("failed to query database: %w", result.Error)
	}
	return &user, nil
}

func (s *Service) CreateUser(req *CreateUser) (*User, error) {
	roles, err := s.GetRolesByIDs(req.Roles)
	if err != nil {
		return nil, fmt.Errorf("group with ids '%v' does not exist", req.Roles)
	}

	user := User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
		Active:    req.Active,
		Phone1:    req.Phone1,
		Phone2:    &req.Phone2,
	}

	if err := s.DB.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user")
	}

	if err := s.DB.Model(&user).Association("Roles").Replace(roles); err != nil {
		return nil, fmt.Errorf("failed to set roles for user")
	}

	return &user, nil
}

func (s *Service) UpdateUserPartial(req *UserSchema) (*User, error) {
	user, err := s.GetUserByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("user modified with id '%v' does not exists", req.ID)
	}
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Active = req.Active
	user.Phone1 = req.Phone1
	user.Phone2 = &req.Phone2

	if err := s.DB.Model(user).Updates(user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %s", err.Error())
	}
	return user, nil
}

func (s *Service) AdminUpdateUser(req *UserSchema) (*User, error) {
	user, err := s.GetUserByID(req.ID)
	if err != nil {
		return nil, fmt.Errorf("user modified with id '%v' does not exists", req.ID)
	}
	if len(req.Roles) > 0 {
		roles, err := s.GetRolesByIDs(req.Roles)
		if err != nil {
			return nil, fmt.Errorf("role with ids '%v' does not exist", req.Roles)
		}

		if err := s.DB.Model(user).Association("Roles").Replace(roles); err != nil {
			return nil, fmt.Errorf("failed to set roles for user: %v", err)
		}
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email
	user.Active = req.Active
	user.Phone1 = req.Phone1
	user.Phone2 = &req.Phone2

	if err := s.DB.Model(user).Updates(user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %s", err.Error())
	}
	return user, nil
}

func (s *Service) ListUsers() ([]User, error) {
	var users []User
	result := s.DB.
		Preload("Roles.Permissions").
		Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query database list: %w", result.Error)
	}
	return users, nil
}
