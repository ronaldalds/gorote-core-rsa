package core

import (
	"fmt"
	"slices"

	"github.com/gofiber/fiber/v2"
)

func (c *Controller) LoginHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*Login)

	user, err := c.Service.Login(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	accessToken, err := SetToken(ctx, user, &ConfigToken{
		TokenType:   "access_token",
		AppName:     c.AppName,
		AppTimeZone: c.AppTimeZone,
		Domain:      c.Domain,
		PrivateKey:  c.PrivateKey,
		Ttl:         c.Jwt.JwtExpireAccess,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	refreshToken, err := SetToken(ctx, user, &ConfigToken{
		TokenType:   "refresh_token",
		AppName:     c.AppName,
		AppTimeZone: c.AppTimeZone,
		Domain:      c.Domain,
		PrivateKey:  c.PrivateKey,
		Ttl:         c.Jwt.JwtExpireRefresh,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	res := &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *Controller) RefrashTokenHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*RefrashToken)
	refreshToken := req.RefreshToken
	if refreshToken == "" {
		refreshToken = ctx.Cookies("refresh_token")
	}

	claims, err := GetJwtHeaderPayloadRSA(refreshToken, &c.PrivateKey.PublicKey)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	user, err := c.Service.GetUserByID(claims.Sub)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if !user.Active {
		return fiber.NewError(fiber.StatusBadRequest, "failed to refrash token: user is inactive")
	}

	accessToken, err := SetToken(ctx, user, &ConfigToken{
		TokenType:   "access_token",
		AppName:     c.AppName,
		AppTimeZone: c.AppTimeZone,
		Domain:      c.Domain,
		PrivateKey:  c.PrivateKey,
		Ttl:         c.Jwt.JwtExpireAccess,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	res := &Token{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *Controller) ListPermissiontHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*Paginate)

	permissions, err := c.Service.ListPermissions()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	countPermissions := uint(len(permissions))

	if err := Pagination(req.Page, req.Limit, &permissions); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	data := []PermissionSchema{}
	for _, permission := range permissions {
		schema := &PermissionSchema{
			ID:          permission.ID,
			Code:        permission.Code,
			Name:        permission.Name,
			Description: permission.Description,
		}
		data = append(data, *schema)
	}

	res := &ListPermission{
		Page:  req.Page,
		Limit: req.Limit,
		Data:  data,
		Total: countPermissions,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *Controller) ListRolesHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*Paginate)

	roles, err := c.Service.ListRoles()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	countRoles := uint(len(roles))

	if err := Pagination(req.Page, req.Limit, &roles); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	data := []RoleSchema{}
	for _, role := range roles {
		schema := &RoleSchema{
			ID:          role.ID,
			Name:        role.Name,
			Description: role.Description,
		}
		for _, permission := range role.Permissions {
			schema.Permissions = append(schema.Permissions, PermissionSchema{
				ID:          permission.ID,
				Code:        permission.Code,
				Name:        permission.Name,
				Description: permission.Description,
			})
		}
		data = append(data, *schema)
	}

	res := &ListRole{
		Page:  req.Page,
		Limit: req.Limit,
		Data:  data,
		Total: countRoles,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *Controller) CreateRoleHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*CreateRole)

	role, err := c.Service.CreateRole(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var permissionCodes []PermissionSchema
	for _, permission := range role.Permissions {
		schema := &PermissionSchema{
			ID:          permission.ID,
			Code:        permission.Code,
			Name:        permission.Name,
			Description: permission.Description,
		}
		permissionCodes = append(permissionCodes, *schema)
	}

	res := &RoleSchema{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: permissionCodes,
	}
	return ctx.Status(fiber.StatusCreated).JSON(res)
}

func (c *Controller) ListUsersHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*Paginate)

	users, err := c.Service.ListUsers()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	countUsers := uint(len(users))

	if err := Pagination(req.Page, req.Limit, &users); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	var data []UserSchema
	for _, user := range users {
		schema := UserSchema{
			ID:          user.ID,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			Email:       user.Email,
			Active:      user.Active,
			IsSuperUser: user.IsSuperUser,
			Phone1:      user.Phone1,
			Phone2:      *user.Phone2,
			Roles:       ExtractNameRolesByUser(user),
		}
		data = append(data, schema)
	}

	res := &ListUser{
		Page:  req.Page,
		Limit: req.Limit,
		Data:  data,
		Total: countUsers,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *Controller) CreateUserHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*CreateUser)

	if err := ValidatePassword(req.Password); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("crypting password failed: %s", err.Error()))
	}
	req.Password = hashedPassword

	user, err := c.Service.CreateUser(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	res := &UserSchema{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Active:      user.Active,
		IsSuperUser: user.IsSuperUser,
		Phone1:      user.Phone1,
		Phone2:      *user.Phone2,
		Roles:       ExtractNameRolesByUser(*user),
	}
	return ctx.Status(fiber.StatusCreated).JSON(res)
}

func (c *Controller) UpdateUserHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*UserSchema)

	claims, err := GetJwtHeaderPayloadRSA(GetAccessToken(ctx), &c.PrivateKey.PublicKey)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	userAdmin := slices.Contains(claims.Permissions, string(PermissionSuperUser)) || slices.Contains(claims.Permissions, string(PermissionUpdateUser))

	var user *User
	if claims.IsSuperUser || userAdmin {
		user, err = c.Service.AdminUpdateUser(req)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	} else if claims.Sub == req.ID {
		user, err = c.Service.UpdateUserPartial(req)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
	} else {
		return fiber.NewError(fiber.StatusBadRequest, "you are not authorized to update this user")
	}

	res := &UserSchema{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Active:      user.Active,
		IsSuperUser: user.IsSuperUser,
		Phone1:      user.Phone1,
		Phone2:      *user.Phone2,
		Roles:       ExtractNameRolesByUser(*user),
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}
