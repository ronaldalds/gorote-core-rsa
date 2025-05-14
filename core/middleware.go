package core

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"reflect"
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

func JWTProtectedRSA(publicKey *rsa.PublicKey, permissions ...PermissionCode) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		token, err := GetJwtHeaderPayloadRSA(ctx.Get("Authorization"), publicKey)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		// Check permissions
		if token.Claims.IsSuperUser {
			return ctx.Next()
		}
		if len(permissions) == 0 {
			return ctx.Next()
		}

		// Check if any required permission exists in user's permissions
		for _, requiredPermission := range permissions {
			if slices.Contains(token.Claims.Permissions, string(requiredPermission)) {
				log.Println("Permission validated, proceeding to next handler")
				return ctx.Next()
			}
		}

		// If no errors, log success and continue to the next handler
		log.Println("JWT validated and session matched, proceeding to next handler")
		return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
}

func ValidationMiddleware(requestStruct any) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		v := reflect.ValueOf(requestStruct)
		if v.Kind() == reflect.Ptr {
			v = v.Elem() // Dereferencia o ponteiro para obter o valor subjacente
		}

		// Verifica se o valor subjacente é uma struct
		if v.Kind() != reflect.Struct {
			return fiber.NewError(fiber.StatusInternalServerError, "validation target must be a struct")
		}

		t := v.Type()
		var foundTag bool
		var parseErr error

		// Verifica todas as tags para determinar o tipo de input
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Verifica tags de query
			if _, ok := field.Tag.Lookup("query"); ok {
				foundTag = true
				if parseErr = ctx.QueryParser(requestStruct); parseErr != nil {
					return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid query parameters: %s", parseErr.Error()))
				}
			}

			// Verifica tags de json
			if _, ok := field.Tag.Lookup("json"); ok {
				foundTag = true
				if parseErr = ctx.BodyParser(requestStruct); parseErr != nil {
					return fiber.NewError(fiber.StatusBadRequest, fmt.Sprintf("invalid body: %s", parseErr.Error()))
				}
			}

			// Verifica tags de params (URL parameters)
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

		// Valide os dados usando o validator
		if err := validateStruct(requestStruct); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		// Armazene os dados validados no contexto
		ctx.Locals("validatedData", requestStruct)

		// Prossiga para o próximo middleware ou handler
		return ctx.Next()
	}
}

func getJSONFieldName(s any, field string) string {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Procura o field na struct
	if f, ok := t.FieldByName(field); ok {
		jsonTag := f.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			// pega só o nome antes da vírgula (caso tenha omitempty, etc.)
			return splitJSONTag(jsonTag)
		}
	}
	return field // fallback pro nome real
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

func Telemetry(funcSend func(LogTelemetry) error, Confidential bool) fiber.Handler {
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

				if err := funcSend(logData); err != nil {
					return fmt.Errorf("error on send log: %v", err.Error())
				}
			}
			return err
		}
		logData.Status = ctx.Response().StatusCode()
		logData.Latency = time.Since(start).Milliseconds()
		logData.ResponseBody = string(ctx.Response().Body())

		if err := funcSend(logData); err != nil {
			log.Println("error on send log:", err)
		}
		return nil
	}
}
