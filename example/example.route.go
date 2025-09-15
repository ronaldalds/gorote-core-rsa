package example

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/ronaldalds/gorote-core-rsa/core"
	"github.com/ronaldalds/gorote-core-rsa/gorote"
)

func (r *appRouter) RegisterRouter(router fiber.Router) {
	r.check(router.Group("/"))
	r.ws(router.Group("/ws"))
}

func (r *appRouter) check(router fiber.Router) {
	router.Get("/", gorote.Check())
}

func (r *appRouter) ws(router fiber.Router) {
	router.Get(
		"/:id",
		gorote.IsWsMiddleware(),
		gorote.ValidationMiddleware(&WsConn{}),
		gorote.JWTProtectedRSA(&core.JwtClaims{}, r.publicKey),
		websocket.New(r.controller.websocketHandler),
	)
}
