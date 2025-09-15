package core

type login struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type refrashToken struct {
	RefreshToken string `json:"refresh_token"`
}

type token struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type createRole struct {
	Name        string   `json:"name" validate:"required,min=3,max=100"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type createUser struct {
	schemaUser
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type recieveUser struct {
	ID string `param:"id" validate:"required"`
}

type schemaUser struct {
	ID          string   `param:"id"`
	FirstName   string   `json:"first_name" validate:"omitempty,min=1,max=50"`
	LastName    string   `json:"last_name" validate:"omitempty,max=50"`
	Active      bool     `json:"active" validate:"omitempty"`
	IsSuperUser bool     `json:"is_super_user" validate:"omitempty"`
	Roles       []string `json:"roles" validate:"omitempty"`
	Tenants     []string `json:"tenants" validate:"omitempty"`
	Phone1      string   `json:"phone1" validate:"omitempty,e164"`
	Phone2      string   `json:"phone2" validate:"omitempty,e164"`
}

type paginateReq struct {
	Page  uint `query:"page" validate:"required,min=1"`
	Limit uint `query:"limit" validate:"required"`
}

type paginateRes struct {
	Page  uint `json:"page" validate:"required,min=1"`
	Limit uint `json:"limit" validate:"required"`
	Total uint `json:"total" validate:"required"`
}

type listUser struct {
	paginateRes
	Data []User `json:"data"`
}

type listRole struct {
	paginateRes
	Data []Role `json:"data"`
}

type listPermission struct {
	paginateRes
	Data []Permission `json:"data"`
}
