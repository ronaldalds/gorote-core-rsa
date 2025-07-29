package core

import (
	"crypto/rsa"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

type AppJwt struct {
	JwtExpireAccess  time.Duration
	JwtExpireRefresh time.Duration
	JwtSecret        string
}

type AppSuper struct {
	SuperName  string
	SuperUser  string
	SuperEmail string
	SuperPass  string
	SuperPhone string
}

type AppConfig struct {
	*fiber.App
	*gorm.DB
	AppName     string
	AppTimeZone string
	PrivateKey  *rsa.PrivateKey
	Jwt         AppJwt
	Super       *AppSuper
	Domain      string
	Meter       *metric.Meter
	Trace       *trace.Tracer
	Logger      *slog.Logger
}

type Router struct {
	AppConfig
	Controller Controller
}

type Controller struct {
	AppConfig
	Service Service
}

type Service struct {
	AppConfig
}

func New(config AppConfig, ready ...func(AppConfig) error) (*Router, error) {
	if err := migrate(config); err != nil {
		return nil, fmt.Errorf("err on migrate: %v", err.Error())
	}
	if config.Super != nil {
		if err := saveUserAdmin(config); err != nil {
			return nil, fmt.Errorf("err on save admin: %v", err.Error())
		}
	}
	if err := savePermissions(config); err != nil {
		return nil, fmt.Errorf("err on save permissions: %v", err.Error())
	}
	router := &Router{
		AppConfig: config,
		Controller: Controller{
			AppConfig: config,
			Service:   Service{config},
		},
	}
	for _, r := range ready {
		if err := r(config); err != nil {
			return nil, fmt.Errorf("err on ready: %v", err.Error())
		}
	}
	return router, nil
}
