package example

import (
	"crypto/rsa"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AppConfig struct {
	*fiber.App
	*gorm.DB
	AppName   string
	TimeZone  string
	PublicKey *rsa.PublicKey
}

type Router struct {
	*AppConfig
	Controller *Controller
}

type Controller struct {
	*AppConfig
	Service *Service
}

type Service struct {
	*AppConfig
}

func NewMicroService(config *AppConfig, ready ...func(*AppConfig) error) (*Router, error) {
	router := &Router{
		AppConfig: config,
		Controller: &Controller{
			AppConfig: config,
			Service: &Service{
				AppConfig: config,
			},
		},
	}
	for _, r := range ready {
		if err := r(config); err != nil {
			return nil, fmt.Errorf("err on ready: %v", err.Error())
		}
	}
	return router, nil
}
