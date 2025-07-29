package core

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ronaldalds/gorote-core-rsa/gorote"
)

func (r *Router) RegisterRouter(router fiber.Router) {
	r.Check(router.Group("/check"))
	r.Health(router.Group("/health"))
	r.Auth(router.Group("/auth", gorote.Limited(60)))
	r.RefrashToken(router.Group("/refrash"))
	r.User(router.Group("/users"))
	r.Role(router.Group("/roles"))
	r.Permission(router.Group("/permissions"))
}

func (r *Router) Check(router fiber.Router) {
	router.Get("/", gorote.Check())
}

func (r *Router) Health(router fiber.Router) {
	router.Get("/", r.Controller.HealthHandler)
}

func (r *Router) Auth(router fiber.Router) {
	router.Post("/login",
		gorote.ValidationMiddleware(&Login{}),
		r.Controller.LoginHandler,
	)
}

func (r *Router) RefrashToken(router fiber.Router) {
	router.Post("/",
		gorote.ValidationMiddleware(&RefrashToken{}),
		r.Controller.RefrashTokenHandler,
	)
}

func (r *Router) User(router fiber.Router) {
	router.Get("/",
		gorote.ValidationMiddleware(&Paginate{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, &r.PrivateKey.PublicKey, ProtectedRoute(PermissionViewUser)),
		r.Controller.ListUsersHandler,
	)
	router.Post("/",
		gorote.ValidationMiddleware(&CreateUser{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, &r.PrivateKey.PublicKey, ProtectedRoute(PermissionCreateUser)),
		r.Controller.CreateUserHandler,
	)
	router.Put("/:id",
		gorote.ValidationMiddleware(&UserParam{}), gorote.ValidationMiddleware(&UserSchema{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, &r.PrivateKey.PublicKey, ProtectedRoute()),
		r.Controller.UpdateUserHandler,
	)
}

func (r *Router) Role(router fiber.Router) {
	router.Get("/",
		gorote.ValidationMiddleware(&Paginate{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, &r.PrivateKey.PublicKey, ProtectedRoute()),
		r.Controller.ListRolesHandler,
	)
	router.Post("/",
		gorote.ValidationMiddleware(&CreateRole{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, &r.PrivateKey.PublicKey, ProtectedRoute(PermissionCreateRole)),
		r.Controller.CreateRoleHandler,
	)
}

func (r *Router) Permission(router fiber.Router) {
	router.Get("/",
		gorote.ValidationMiddleware(&Paginate{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, &r.PrivateKey.PublicKey, ProtectedRoute(PermissionViewPermission)),
		r.Controller.ListPermissiontHandler,
	)
}
