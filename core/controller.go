package core

import (
	"fmt"
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/ronaldalds/gorote-core-rsa/gorote"
)

func (c *Controller) HealthHandler(ctx *fiber.Ctx) error {
	res, err := c.Service.Health()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *Controller) CreateRoleHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*CreateRole)
	role, err := c.Service.CreateRole(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return ctx.Status(fiber.StatusCreated).JSON(role)
}

func (c *Controller) CreateUserHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*CreateUser)
	if err := gorote.ValidatePassword(req.Password); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	hashedPassword, err := gorote.HashPassword(req.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("crypting password failed: %s", err.Error()))
	}
	req.Password = hashedPassword
	user, err := c.Service.CreateUser(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	var roles []uint
	for _, role := range user.Roles {
		roles = append(roles, role.ID)
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
		Roles:       roles,
	}
	return ctx.Status(fiber.StatusCreated).JSON(res)
}

func (c *Controller) UpdateUserHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*UserSchema)
	claims := ctx.Locals("claimsData").(*JwtClaims)
	userAdmin := slices.Contains(claims.Permissions, string(PermissionSuperUser)) || slices.Contains(claims.Permissions, string(PermissionUpdateUser))
	var user *User
	var err error
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
	var roles []uint
	for _, role := range user.Roles {
		roles = append(roles, role.ID)
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
		Roles:       roles,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}
