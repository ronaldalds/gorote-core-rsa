package core

import "github.com/gofiber/fiber/v2"

func (r *Router) RegisterRouter(router fiber.Router) {
	r.Check(router.Group("/check"))
	r.Health(router.Group("/health"))
	r.Auth(router.Group("/auth", Limited(10)))
	r.RefrashToken(router.Group("/refrash"))
	r.User(router.Group("/users"))
	r.Role(router.Group("/roles"))
	r.Permission(router.Group("/permissions"))
}

func (r *Router) Check(router fiber.Router) {
	router.Get(
		"/",
		Check(),
	)
}

func (r *Router) Health(router fiber.Router) {
	router.Get(
		"/",
		r.HealthGorm(),
	)
}

func (r *Router) Auth(router fiber.Router) {
	router.Post(
		"/login",
		ValidationMiddleware(&Login{}),
		r.Controller.LoginHandler,
	)
}

func (r *Router) RefrashToken(router fiber.Router) {
	router.Post(
		"/",
		ValidationMiddleware(&RefrashToken{}),
		r.Controller.RefrashTokenHandler,
	)
}

func (r *Router) User(router fiber.Router) {
	router.Get(
		"/",
		ValidationMiddleware(&Paginate{}),
		JWTProtectedRSA(&r.PrivateKey.PublicKey, PermissionViewUser),
		r.Controller.ListUsersHandler,
	)
	router.Post(
		"/",
		ValidationMiddleware(&CreateUser{}),
		JWTProtectedRSA(&r.PrivateKey.PublicKey, PermissionCreateUser),
		r.Controller.CreateUserHandler,
	)
	router.Put(
		"/:id",
		ValidationMiddleware(&UserParam{}),
		ValidationMiddleware(&UserSchema{}),
		JWTProtectedRSA(&r.PrivateKey.PublicKey),
		r.Controller.UpdateUserHandler,
	)
}

func (r *Router) Role(router fiber.Router) {
	router.Get(
		"/",
		ValidationMiddleware(&Paginate{}),
		JWTProtectedRSA(&r.PrivateKey.PublicKey),
		r.Controller.ListRolesHandler,
	)
	router.Post(
		"/",
		ValidationMiddleware(&CreateRole{}),
		JWTProtectedRSA(&r.PrivateKey.PublicKey, PermissionCreateRole),
		r.Controller.CreateRoleHandler,
	)
}

func (r *Router) Permission(router fiber.Router) {
	router.Get(
		"/",
		ValidationMiddleware(&Paginate{}),
		JWTProtectedRSA(&r.PrivateKey.PublicKey, PermissionViewPermission),
		r.Controller.ListPermissiontHandler,
	)
}
