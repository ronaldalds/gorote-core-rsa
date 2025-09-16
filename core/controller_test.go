package core

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAuthintegration(t *testing.T) {
	app := fiber.New(fiber.Config{AppName: "test"})
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("err on open db: %v", err.Error())
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("err on read private key: %v", err.Error())
	}

	auth := Config{
		DB:               db,
		PrivateKey:       privateKey,
		JwtExpireAccess:  time.Hour,
		JwtExpireRefresh: time.Hour * 24,
		SuperEmail:       "admin@admin.com",
		SuperPass:        "Senha@123",
		Domain:           ".ralds.com.br,.ralds.br",
	}
	router, err := New(&auth)
	if err != nil {
		t.Fatalf("err on new auth: %v", err.Error())
	}

	router.RegisterRouter(app.Group("/test"))

	var Token token

	t.Run("login superadmin", func(t *testing.T) {
		body := `{"email": "admin@admin.com", "password": "Senha@123"}`
		req := httptest.NewRequest("POST", "/test/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("err on test: %v", err.Error())
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("esperava status 200, recebeu %d", resp.StatusCode)
		}
		if err := json.NewDecoder(resp.Body).Decode(&Token); err != nil {
			t.Fatalf("err on decode: %v", err.Error())
		}
	})

	t.Run("receive user", func(t *testing.T) {
		tk, _, err := jwt.NewParser().ParseUnverified(Token.AccessToken, &JwtClaims{})
		if err != nil {
			t.Error("err on validate token")
		}
		claims, ok := tk.Claims.(*JwtClaims)
		if !ok {
			t.Error("err on validate token")
		}
		req := httptest.NewRequest("GET", fmt.Sprintf("/test/users/%s", claims.ID), strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", Token.AccessToken)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("err on test: %v", err.Error())
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("esperava status 200, recebeu %d", resp.StatusCode)
		}
	})

	t.Run("list permissions", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test/permissions?page=1&limit=10", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", Token.AccessToken)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("err on test: %v", err.Error())
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("esperava status 200, recebeu %d", resp.StatusCode)
		}
	})
}
