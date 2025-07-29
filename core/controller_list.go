package core

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ronaldalds/gorote-core-rsa/gorote"
)

func (c *Controller) ListPermissiontHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*Paginate)
	permissions, err := c.Service.ListPermissions()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	countPermissions := uint(len(permissions))
	if err := gorote.Pagination(req.Page, req.Limit, &permissions); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	res := &ListPermission{
		Page:  req.Page,
		Limit: req.Limit,
		Data:  permissions,
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
	if err := gorote.Pagination(req.Page, req.Limit, &roles); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	res := &ListRole{
		Page:  req.Page,
		Limit: req.Limit,
		Data:  roles,
		Total: countRoles,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}

func (c *Controller) ListUsersHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*Paginate)
	users, err := c.Service.ListUsers()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	countUsers := uint(len(users))
	if err := gorote.Pagination(req.Page, req.Limit, &users); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	res := &ListUser{
		Page:  req.Page,
		Limit: req.Limit,
		Data:  users,
		Total: countUsers,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}
