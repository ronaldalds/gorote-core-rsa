package core

type HealthHandler struct {
	Sql map[string]string `json:"sql"`
}

type Paginate struct {
	Page  uint `query:"page" validate:"required,min=1"`
	Limit uint `query:"limit" validate:"required"`
}

type Login struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type Token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type CreateRole struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Description string `json:"description"`
	Permissions []uint `json:"permissions"`
}

type CreateUser struct {
	UserSchema
	Password string `json:"password" validate:"required,min=6"`
}

// DTO model
type UserParam struct {
	ID uint `param:"id"`
}
type UserSchema struct {
	ID          uint   `json:"id"`
	FirstName   string `json:"firstName" validate:"required,min=1,max=50"`
	LastName    string `json:"lastName" validate:"omitempty,max=50"`
	Username    string `json:"username" validate:"required,min=3,max=50"`
	Email       string `json:"email" validate:"required,email"`
	Active      bool   `json:"active"`
	IsSuperUser bool   `json:"isSuperUser"`
	Roles       []uint `json:"roles"`
	Phone1      string `json:"phone1" validate:"required,e164"`
	Phone2      string `json:"phone2" validate:"omitempty,e164"`
}

type RoleSchema struct {
	ID          uint               `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Permissions []PermissionSchema `json:"permissions"`
}

type PermissionSchema struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DTO list

type ListUser struct {
	Page  uint         `json:"page" validate:"required,min=1"`
	Limit uint         `json:"limit" validate:"required"`
	Data  []UserSchema `json:"data"`
	Total uint         `json:"total" validate:"required"`
}

type ListRole struct {
	Page  uint         `json:"page"`
	Limit uint         `json:"limit"`
	Data  []RoleSchema `json:"data"`
	Total uint         `json:"total"`
}

type ListPermission struct {
	Page  uint               `json:"page" validate:"required,min=1"`
	Limit uint               `json:"limit" validate:"required"`
	Data  []PermissionSchema `json:"data"`
	Total uint               `json:"total" validate:"required"`
}
