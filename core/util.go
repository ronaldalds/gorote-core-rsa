package core

import (
	"fmt"
	"os"
	"strconv"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

func ExtractNameRolesByUser(user User) []uint {
	var data []uint
	for _, role := range user.Roles {
		data = append(data, role.ID)
	}
	return data
}

func ExtractCodePermissionsByUser(user *User) []string {
	var codePermissions []string
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			codePermissions = append(codePermissions, permission.Code)
		}
	}
	return codePermissions
}

func ContainsAll(listX, listY []Role) bool {
	itemMap := make(map[uint]bool)
	for _, item := range listX {
		itemMap[item.ID] = true
	}

	for _, item := range listY {
		if !itemMap[item.ID] {
			return false
		}
	}

	return true
}

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

func RemoveInvisibleChars(input string) string {
	var result []rune
	for _, r := range input {
		if unicode.IsPrint(r) && r != '\u200b' {
			result = append(result, r)
		}
	}
	return string(result)
}

func GetEnvAsTime(key string, required bool, defaultValue ...int) time.Duration {
	valueStr := os.Getenv(key)

	if valueStr == "" {
		if required {
			panic(fmt.Sprintf("variable %s is required", key))
		}
		if len(defaultValue) > 0 {
			return time.Duration(defaultValue[0]) * time.Minute
		}
		return 0
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		panic(fmt.Sprintf("failed to convert %s to integer: %v", key, err))
	}
	return time.Duration(value) * time.Minute
}

func GetEnvAsInt(key string, required bool, defaultValue ...int) int {
	valueStr := os.Getenv(key)

	if valueStr == "" {
		if required {
			panic(fmt.Sprintf("variable %s is required", key))
		}
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return 0
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		panic(fmt.Sprintf("failed to convert %s to integer: %v", key, err))
	}
	return value
}

func GetEnv(key string, required bool, defaultValue ...string) string {
	value := os.Getenv(key)

	if value == "" {
		if required {
			panic(fmt.Sprintf("variable %s is required", key))
		}
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		return ""
	}
	return value
}
