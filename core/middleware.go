package core

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"log"
	"reflect"
	"slices"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

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

func GetAccessToken(ctx *fiber.Ctx) string {
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		authHeader = ctx.Cookies("access_token")
	}
	return authHeader
}

func JWTProtectedRSA(publicKey *rsa.PublicKey, permissions ...PermissionCode) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		claims, err := GetJwtHeaderPayloadRSA(GetAccessToken(ctx), publicKey)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		if claims.Type == "refresh_token" {
			return fiber.NewError(fiber.StatusUnauthorized, "token is refresh token")
		}

		if claims.IsSuperUser {
			return ctx.Next()
		}
		if len(permissions) == 0 {
			return ctx.Next()
		}

		for _, requiredPermission := range permissions {
			if slices.Contains(claims.Permissions, string(requiredPermission)) {
				return ctx.Next()
			}
		}

		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
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

func getJSONFieldName(s any, field string) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if f, ok := t.FieldByName(field); ok {
		jsonTag := f.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			return splitJSONTag(jsonTag)
		}
	}
	return field
}

func splitJSONTag(tag string) string {
	for i, c := range tag {
		if c == ',' {
			return tag[:i]
		}
	}
	return tag
}

func validateStruct(data any) error {
	validate := validator.New()
	err := validate.Struct(data)
	if err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, err := range validationErrors {
				jsonField := getJSONFieldName(data, err.StructField())
				return fmt.Errorf("invalid validation: (field: '%s' is %s type: %s)", jsonField, err.ActualTag(), err.Type())
			}
		}
		return fmt.Errorf("invalid data: %s", err.Error())
	}
	return nil
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

func Telemetry(funcSend func(*LogTelemetry) error, Confidential bool) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		start := time.Now()
		var body map[string]any
		var logData LogTelemetry
		ctx.BodyParser(&body)

		logData.Timestamp = time.Now().Format(time.RFC3339)
		logData.Method = ctx.Method()
		logData.Path = ctx.Path()
		logData.Headers = ctx.GetReqHeaders()
		logData.IP = ctx.IP()
		logData.RequestBody = body

		if Confidential {
			logData.RequestBody = map[string]any{"confidential": "**********************"}
		}

		if err := ctx.Next(); err != nil {
			var e *fiber.Error
			if errors.As(err, &e) {
				logData.Status = e.Code
				logData.Latency = time.Since(start).Milliseconds()
				logData.ResponseBody = e.Message

				if err := funcSend(&logData); err != nil {
					return fmt.Errorf("error on send log: %v", err.Error())
				}
			}
			return err
		}
		logData.Status = ctx.Response().StatusCode()
		logData.Latency = time.Since(start).Milliseconds()
		logData.ResponseBody = string(ctx.Response().Body())

		if err := funcSend(&logData); err != nil {
			log.Println("error on send log:", err)
		}
		return nil
	}
}
