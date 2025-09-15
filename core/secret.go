package core

import (
	"slices"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type JwtClaims struct {
	IsSuperUser bool     `json:"isSuperUser"`
	Permissions []string `json:"permissions"`
	Tenants     []string `json:"tenants"`
	Type        string   `json:"type"`
	jwt.RegisteredClaims
}

func ProtectedRoute(p ...PermissionCode) func(jwt.Claims) *fiber.Error {
	return func(c jwt.Claims) *fiber.Error {
		claims, ok := c.(*JwtClaims)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid claims type")
		}
		if claims.Type == "refresh_token" {
			return fiber.NewError(fiber.StatusUnauthorized, "token is refresh token")
		}
		if claims.IsSuperUser {
			return nil
		}
		if len(p) == 0 {
			return nil
		}
		for _, permission := range p {
			if slices.Contains(claims.Permissions, string(permission)) {
				return nil
			}
		}
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
}
