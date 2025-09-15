package core

import (
	"crypto/rsa"
	"time"

	"gorm.io/gorm"
)

type jwtConfig struct {
	JwtExpireAccess  time.Duration
	JwtExpireRefresh time.Duration
}

type super struct {
	SuperEmail string
	SuperPass  string
}

type Config struct {
	*gorm.DB
	PrivateKey       *rsa.PrivateKey
	JwtExpireAccess  time.Duration
	JwtExpireRefresh time.Duration
	SuperEmail       string
	SuperPass        string
	Domain           string
}

func (c *Config) db() *gorm.DB {
	return c.DB
}

func (c *Config) domain() string {
	return c.Domain
}

func (c *Config) jwt() *jwtConfig {
	return &jwtConfig{
		JwtExpireAccess:  c.JwtExpireAccess,
		JwtExpireRefresh: c.JwtExpireRefresh,
	}
}

func (c *Config) super() *super {
	if c.SuperEmail == "" || c.SuperPass == "" {
		return nil
	}
	return &super{
		SuperEmail: c.SuperEmail,
		SuperPass:  c.SuperPass,
	}
}

func (c *Config) privateKeyRSA() *rsa.PrivateKey {
	return c.PrivateKey
}

type configLoad interface {
	db() *gorm.DB
	privateKeyRSA() *rsa.PrivateKey
	super() *super
	jwt() *jwtConfig
	domain() string
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
	if err := migrate(config); err != nil {
		return nil, err
	}

	if config.super() != nil {
		if err := saveUserAdmin(config); err != nil {
			return nil, err
		}
	}
	if err := savePermissions(config); err != nil {
		return nil, err
	}

	service := appService{config}

	controller := appController{
		service: &service,
	}

	router := appRouter{
		publicKey:  &config.privateKeyRSA().PublicKey,
		controller: &controller,
	}

	return &router, nil
}
