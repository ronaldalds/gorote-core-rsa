package core

import (
	"crypto/rsa"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AppJwt struct {
	JwtExpireAccess  time.Duration
	JwtExpireRefresh time.Duration
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
	*GormStore
	AppName     string
	AppTimeZone string
	PrivateKey  *rsa.PrivateKey
	Jwt         *AppJwt
	Super       *AppSuper
	Domain      string
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
	TimeUCT *time.Location
}

func New(config *AppConfig) *Router {
	if err := config.PreReady(); err != nil {
		log.Fatal("err on pre ready: ", err.Error())
	}
	return &Router{
		AppConfig:  config,
		Controller: NewController(config),
	}
}

func NewController(config *AppConfig) *Controller {
	return &Controller{
		AppConfig: config,
		Service:   NewService(config),
	}
}

func NewService(config *AppConfig) *Service {
	location, err := time.LoadLocation(config.AppTimeZone)
	if err != nil {
		log.Fatal("invalid timezone: ", err.Error())
	}
	service := &Service{
		AppConfig: config,
		TimeUCT:   location,
	}
	if err := service.PosReady(); err != nil {
		log.Fatal("err on pos ready: ", err.Error())
	}
	return service
}
