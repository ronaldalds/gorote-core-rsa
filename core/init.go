package core

import (
	"crypto/rsa"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
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
	App         *fiber.App
	AppName     string
	AppTimeZone string
	CoreGorm    *gorm.DB
	PrivateKey  *rsa.PrivateKey
	Jwt         *AppJwt
	Super       *AppSuper
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
		log.Fatal(err.Error())
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
		log.Fatal(err.Error())
	}
	return service
}
