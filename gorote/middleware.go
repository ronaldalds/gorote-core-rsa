package gorote

import (
	"crypto/rsa"
	"fmt"
	"reflect"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/golang-jwt/jwt/v5"
)

type HandlerJWTProtected func(jwt.Claims) *fiber.Error

func Check() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		return ctx.Status(fiber.StatusOK).JSON(map[string]string{"status": "OK"})
	}
}

func IsWsMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(ctx) {
			return fiber.NewError(fiber.StatusUpgradeRequired, "upgrade required")
		}
		return ctx.Next()
	}
}

func JWTProtected(claims jwt.Claims, jwtSecret string, handles ...HandlerJWTProtected) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if err := ValidateOrGetJWT(claims, GetAccessToken(ctx), jwtSecret); err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}
		for _, handle := range handles {
			if err := handle(claims); err != nil {
				return err
			}
		}
		ctx.Locals("claimsData", claims)
		return ctx.Next()
	}
}

func JWTProtectedRSA(claims jwt.Claims, publicKey *rsa.PublicKey, handles ...HandlerJWTProtected) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		if err := ValidateOrGetJWTRSA(claims, GetAccessToken(ctx), publicKey); err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}
		for _, handle := range handles {
			if err := handle(claims); err != nil {
				return err
			}
		}
		ctx.Locals("claimsData", claims)
		return ctx.Next()
	}
}

func ValidationMiddleware(requestStruct any) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		v := reflect.ValueOf(requestStruct)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return fiber.NewError(fiber.StatusInternalServerError, "validation target must be a struct")
		}
		t := v.Type()
		var foundTag bool
		var parseErr error
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if _, ok := field.Tag.Lookup("query"); ok {
				foundTag = true
				if parseErr = ctx.QueryParser(requestStruct); parseErr != nil {
					return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid query parameters: %s", parseErr.Error()))
				}
			}
			if _, ok := field.Tag.Lookup("json"); ok {
				foundTag = true
				if parseErr = ctx.BodyParser(requestStruct); parseErr != nil {
					return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid body: %s", parseErr.Error()))
				}
			}
			if _, ok := field.Tag.Lookup("param"); ok {
				foundTag = true
				if parseErr = ctx.ParamsParser(requestStruct); parseErr != nil {
					return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid URL parameters: %s", parseErr.Error()))
				}
			}
		}
		if !foundTag {
			return fiber.NewError(fiber.StatusBadRequest, "no valid tags found in struct (query, json or params)")
		}
		if err := validateStruct(requestStruct); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		ctx.Locals("validatedData", requestStruct)
		return ctx.Next()
	}
}

func Cached(ttl time.Duration) func(ctx *fiber.Ctx) error {
	config := cache.Config{
		Expiration: ttl,
	}
	return cache.New(config)
}

func Limited(max int) func(c *fiber.Ctx) error {
	config := limiter.Config{
		Max: max,
		LimitReached: func(c *fiber.Ctx) error {
			return fiber.ErrTooManyRequests
		},
	}
	return limiter.New(config)
}
