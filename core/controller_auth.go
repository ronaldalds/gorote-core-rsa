package core

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ronaldalds/gorote-core-rsa/gorote"
)

func (c *Controller) LoginHandler(ctx *fiber.Ctx) error {
	req := ctx.Locals("validatedData").(*Login)
	user, err := c.Service.Login(req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	var permissions []string
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			permissions = append(permissions, permission.Code)
		}
	}
	var tenants []uint
	for _, tenant := range user.Tenants {
		tenants = append(tenants, tenant.ID)
	}
	accessToken, err := gorote.GenerateTokenRSA(JwtClaims{
		Sub:         user.ID,
		Email:       user.Email,
		IsSuperUser: user.IsSuperUser,
		Permissions: permissions,
		Tenants:     tenants,
		Type:        "access_token",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    c.AppName,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(c.Jwt.JwtExpireAccess)),
		},
	}, c.PrivateKey)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := gorote.SetTokenCookie(ctx, c.Domain, "access_token", accessToken, c.Jwt.JwtExpireAccess); err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	refreshToken, err := gorote.GenerateTokenRSA(JwtClaims{
		Sub:         user.ID,
		Email:       user.Email,
		IsSuperUser: user.IsSuperUser,
		Permissions: permissions,
		Tenants:     tenants,
		Type:        "refresh_token",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    c.AppName,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(c.Jwt.JwtExpireRefresh)),
		},
	}, c.PrivateKey)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := gorote.SetTokenCookie(ctx, c.Domain, "refresh_token", refreshToken, c.Jwt.JwtExpireRefresh); err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
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
	var claims JwtClaims
	if err := gorote.ValidateOrGetJWTRSA(&claims, refreshToken, &c.PrivateKey.PublicKey); err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, err.Error())
	}
	user, err := c.Service.GetUserByID(claims.Sub)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if !user.Active {
		return fiber.NewError(fiber.StatusBadRequest, "failed to refrash token: user is inactive")
	}
	var permissions []string
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			permissions = append(permissions, permission.Code)
		}
	}
	var tenants []uint
	for _, tenant := range user.Tenants {
		tenants = append(tenants, tenant.ID)
	}
	accessToken, err := gorote.GenerateTokenRSA(JwtClaims{
		Sub:         user.ID,
		Email:       user.Email,
		IsSuperUser: user.IsSuperUser,
		Permissions: permissions,
		Tenants:     tenants,
		Type:        "access_token",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    c.AppName,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(c.Jwt.JwtExpireAccess)),
		},
	}, c.PrivateKey)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := gorote.SetTokenCookie(ctx, c.Domain, "access_token", accessToken, c.Jwt.JwtExpireAccess); err != nil {
		return fiber.NewError(fiber.StatusBadGateway, err.Error())
	}
	res := &Token{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
	}
	return ctx.Status(fiber.StatusOK).JSON(res)
}
