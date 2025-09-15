package gorote

import (
	"crypto/rsa"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password")
	}
	return string(hashedPassword), nil
}

func CheckPasswordHash(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func ValidatePassword(password string) error {
	hasUpper := false
	hasSymbol := false
	for _, r := range password {
		if unicode.IsUpper(r) {
			hasUpper = true
		}
		if unicode.IsSymbol(r) || unicode.IsPunct(r) {
			hasSymbol = true
		}
	}
	if !hasUpper {
		return fmt.Errorf("uppercase-password must contain at least one uppercase letter")
	}
	if !hasSymbol {
		return fmt.Errorf("symbol-password must contain at least one symbol")
	}
	return nil
}

func Pagination[T any](page, limit uint, data *[]T) error {
	count := uint(len(*data))
	start := (page - 1) * limit
	end := page * limit
	if start >= count {
		return fmt.Errorf("page not exist")
	}
	if end > count {
		end = count
	}
	*data = (*data)[start:end]
	return nil
}

func GetAccessToken(ctx *fiber.Ctx) string {
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		authHeader = ctx.Cookies("access_token")
	}
	return authHeader
}

func RemoveInvisibleChars(input string) string {
	var result []rune
	for _, r := range input {
		if unicode.IsPrint(r) && r != '\u200b' {
			result = append(result, r)
		}
	}
	return string(result)
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

func MustEnvAsTime(key string, defaultValue ...int) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		if len(defaultValue) > 0 {
			return time.Duration(defaultValue[0]) * time.Second
		}
		panic(fmt.Sprintf("variable %s is required", key))
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		panic(fmt.Sprintf("failed to convert %s to integer: %v", key, err))
	}
	return time.Duration(value) * time.Second
}

func MustEnvAsInt(key string, defaultValue ...int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		panic(fmt.Sprintf("variable %s is required", key))
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		panic(fmt.Sprintf("failed to convert %s to integer: %v", key, err))
	}
	return value
}

func MustEnvAsString(key string, defaultValue ...string) string {
	value := os.Getenv(key)
	if value == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		panic(fmt.Sprintf("variable %s is required", key))
	}
	return value
}

func MustEnvPublicKeyRSA(key string) *rsa.PublicKey {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("variable %s is required", key))
	}
	return MustReadPublicKeyFromString(value)
}

func MustEnvPrivateKeyRSA(key string) *rsa.PrivateKey {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("variable %s is required", key))
	}
	return MustReadPrivateKeyFromString(value)
}
