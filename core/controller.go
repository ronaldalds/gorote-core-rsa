package core

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

func (con *Controller) HealthHandler(ctx *fiber.Ctx) error {
	sql, err := con.Service.HealthGorm()
	if err != nil {
		log.Println(err.Error())
	}

	health := &HealthHandler{
		Sql: sql,
	}
	return ctx.Status(fiber.StatusOK).JSON(health)
}

func (con *Controller) LoginHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*Login)

	// find username or email in database
	user, err := con.Service.Login(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	permissions := ExtractCodePermissionsByUser(user)

	// generate tokens
	accessToken, err := GenerateTokenRSA(&GenToken{
		Id:          user.ID,
		AppName:     con.AppName,
		Permissions: permissions,
		IsSuperUser: user.IsSuperUser,
		TimeZone:    con.AppTimeZone,
		PrivateKey:  con.PrivateKey,
		Ttl:         con.Jwt.JwtExpireAccess,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	refreshToken, err := GenerateTokenRSA(&GenToken{
		Id:          user.ID,
		AppName:     con.AppName,
		Permissions: permissions,
		IsSuperUser: user.IsSuperUser,
		TimeZone:    con.AppTimeZone,
		PrivateKey:  con.PrivateKey,
		Ttl:         con.Jwt.JwtExpireRefresh,
	})
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// send response
	res := &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (con *Controller) RefrashTokenHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*RefrashToken)

	userID, err := GetJwtHeaderPayloadRSA(ctx.Get("Authorization"), &con.PrivateKey.PublicKey)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}

	user, err := con.Service.GetUserByID(userID.Claims.Sub)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if !user.Active {
		return fiber.NewError(fiber.StatusUnauthorized, "failed to refrash token: user is inactive")
	}

	permissions := ExtractCodePermissionsByUser(user)

	// generate accesstokens
	accessToken, err := GenerateTokenRSA(&GenToken{
		Id:          user.ID,
		AppName:     con.AppName,
		Permissions: permissions,
		IsSuperUser: user.IsSuperUser,
		TimeZone:    con.AppTimeZone,
		PrivateKey:  con.PrivateKey,
		Ttl:         con.Jwt.JwtExpireAccess,
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

func (con *Controller) ListPermissiontHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*Paginate)

	var permissions []Permission
	if err := con.Service.ListPermission(&permissions); err != nil {
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

func (con *Controller) ListRoleHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*Paginate)

	var roles []Role
	if err := con.Service.ListRole(&roles); err != nil {
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

func (con *Controller) CreateRoleHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*CreateRole)

	var role Role
	if err := con.Service.CreateRole(&role, req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Extrair os codes das permissões
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

	// Preparar a resposta
	res := &RoleSchema{
		ID:          role.ID,
		Name:        role.Name,
		Permissions: permissionCodes, // Adicionar apenas os codes
	}
	return ctx.Status(fiber.StatusCreated).JSON(res)
}

func (con *Controller) ListUserHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*Paginate)

	users, err := con.Service.ListUser()
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
			Username:    user.Username,
			Email:       user.Email,
			Active:      user.Active,
			IsSuperUser: user.IsSuperUser,
			Phone1:      user.Phone1,
			Phone2:      user.Phone2,
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

func (con *Controller) CreateUserHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*CreateUser)

	if err := ValidatePassword(req.Password); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(err)
	}
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("crypting password failed: %s", err.Error()))
	}
	req.Password = hashedPassword

	creator, err := GetJwtHeaderPayloadRSA(ctx.Get("Authorization"), &con.PrivateKey.PublicKey)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := con.Service.CreateUser(creator.Claims.Sub, req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	res := &UserSchema{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Username:    user.Username,
		Email:       user.Email,
		Active:      user.Active,
		IsSuperUser: user.IsSuperUser,
		Phone1:      user.Phone1,
		Phone2:      user.Phone2,
		Roles:       ExtractNameRolesByUser(*user),
	}
	return ctx.Status(fiber.StatusCreated).JSON(res)
}

func (con *Controller) UpdateUserHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*UserSchema)

	fmt.Println(req.ID)

	editor, err := GetJwtHeaderPayloadRSA(ctx.Get("Authorization"), &con.PrivateKey.PublicKey)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	user, err := con.Service.UpdateUser(editor.Claims.Sub, uint(req.ID), req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	res := &UserSchema{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Username:    user.Username,
		Email:       user.Email,
		Active:      user.Active,
		IsSuperUser: user.IsSuperUser,
		Phone1:      user.Phone1,
		Phone2:      user.Phone2,
		Roles:       ExtractNameRolesByUser(*user),
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}
