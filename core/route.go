package core

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/ronaldalds/gorote-core-rsa/gorote"
)

func (r *appRouter) RegisterRouter(router fiber.Router) {
	r.Check(router.Group("/check"))
	r.Health(router.Group("/health"))
	r.Auth(router.Group("/auth", gorote.Limited(60)))
	r.User(router.Group("/users"))
	r.Role(router.Group("/roles"))
	r.Permission(router.Group("/permissions"))
	r.Swagger(router.Group("/swagger"))
}

func (r *appRouter) Check(router fiber.Router) {
	router.Get("/", gorote.Check())
}

func (r *appRouter) Health(router fiber.Router) {
	router.Get("/", r.controller.healthHandler)
}

func (r *appRouter) Auth(router fiber.Router) {
	router.Post("/login",
		gorote.ValidationMiddleware(&login{}),
		r.controller.loginHandler,
	)
	router.Post("/refresh",
		gorote.ValidationMiddleware(&refreshToken{}),
		r.controller.refreshTokenHandler,
	)
}

func (r *appRouter) User(router fiber.Router) {
	router.Get("/",
		gorote.ValidationMiddleware(&paginateReq{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, r.publicKey, ProtectedRoute(PermissionViewUser)),
		r.controller.listUsersHandler,
	)
	router.Get("/:id",
		gorote.ValidationMiddleware(&recieveUser{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, r.publicKey, ProtectedRoute(PermissionViewUser)),
		r.controller.recieveUserHandler,
	)
	router.Post("/",
		gorote.ValidationMiddleware(&createUser{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, r.publicKey, ProtectedRoute(PermissionCreateUser)),
		r.controller.createUserHandler,
	)
	router.Put("/:id",
		gorote.ValidationMiddleware(&schemaUser{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, r.publicKey, ProtectedRoute()),
		r.controller.updateUserHandler,
	)
}

func (r *appRouter) Role(router fiber.Router) {
	router.Get("/",
		gorote.ValidationMiddleware(&paginateReq{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, r.publicKey, ProtectedRoute()),
		r.controller.listRolesHandler,
	)
	router.Post("/",
		gorote.ValidationMiddleware(&createRole{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, r.publicKey, ProtectedRoute(PermissionCreateRole)),
		r.controller.createRoleHandler,
	)
}

func (r *appRouter) Permission(router fiber.Router) {
	router.Get("/",
		gorote.ValidationMiddleware(&paginateReq{}),
		gorote.JWTProtectedRSA(&JwtClaims{}, r.publicKey, ProtectedRoute(PermissionViewPermission)),
		r.controller.listPermissiontHandler,
	)
}

func (r *appRouter) Swagger(router fiber.Router) {
	router.Get("/*", swagger.HandlerDefault)
}
