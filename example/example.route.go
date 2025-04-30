package example

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/ronaldalds/gorote-core-rsa/core"
)

func (r *Router) RegisterRouter(router fiber.Router) {
	r.Check(router.Group("/"))
	r.Health(router.Group("/health"))
	r.Ws(router.Group("/ws"))
}

func (r *Router) Check(router fiber.Router) {
	router.Get(
		"/",
		core.Check(),
	)
}

func (r *Router) Health(router fiber.Router) {
	router.Get(
		"/",
		r.HealthGorm(),
	)
}

func (r *Router) Ws(router fiber.Router) {
	router.Get(
		"/:id",
		core.IsWsMiddleware(),
		core.ValidationMiddleware(&WsConn{}),
		core.JWTProtectedRSA(r.PublicKey),
		websocket.New(r.Controller.websocketHandler),
	)
}
