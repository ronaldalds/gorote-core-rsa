package example

import (
	"crypto/rsa"

	"github.com/gofiber/contrib/websocket"
	"gorm.io/gorm"
)

type Config struct {
	*gorm.DB
	PublicKey *rsa.PublicKey
}

func (c *Config) db() *gorm.DB {
	return c.DB
}

func (c *Config) publicKeyRSA() *rsa.PublicKey {
	return c.PublicKey
}

type configLoad interface {
	db() *gorm.DB
	publicKeyRSA() *rsa.PublicKey
}

type controller interface {
	websocketHandler(ctx *websocket.Conn)
}

type servicer interface {
	getConnection(id uint) (*websocket.Conn, bool)
	broadcast(message any)
	sendTo(id uint, message any) error
}

type appRouter struct {
	publicKey  *rsa.PublicKey
	controller controller
}

type appController struct {
	service servicer
}

type appService struct {
	configLoad
}

func New(config configLoad) (*appRouter, error) {
	if err := config.db().AutoMigrate(); err != nil {
		return nil, err
	}

	service := appService{config}

	controller := appController{
		service: &service,
	}

	router := appRouter{
		publicKey:  config.publicKeyRSA(),
		controller: &controller,
	}

	return &router, nil
}
