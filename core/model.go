package core

import (
	"gorm.io/gorm"
)

type Permission struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;size:100;not null" validate:"required,min=3,max=100"`
	Code        string `gorm:"uniqueIndex;size:50;not null" validate:"required"`
	Description string `gorm:"size:255"` // Descrição opcional da permission
	Active      bool   `gorm:"default:true"`
}

type Role struct {
	gorm.Model
	Name        string       `gorm:"uniqueIndex;size:100;not null" validate:"required,min=3,max=100"`
	Description string       `gorm:"size:255"`
	Permissions []Permission `gorm:"many2many:roles_permissions"`
	Active      bool         `gorm:"default:true"`
}

// User representa o modelo de usuário no sistema.
type User struct {
	gorm.Model
	FirstName   string `gorm:"size:50;not null" validate:"required,min=1,max=50"`
	LastName    string `gorm:"size:50" validate:"omitempty,max=50"`
	Username    string `gorm:"uniqueIndex;size:50;not null" validate:"required,min=3,max=50"`
	Email       string `gorm:"uniqueIndex;not null" validate:"required,email"`
	Password    string `gorm:"not null" validate:"required"`
	Active      bool   `gorm:"default:true"`
	IsSuperUser bool   `gorm:"default:false"`
	Roles       []Role `gorm:"many2many:users_roles"`
	Phone1      string `gorm:"type:varchar(20);not null" validate:"required,e164"`
	Phone2      string `gorm:"type:varchar(20);nullable" validate:"omitempty,e164"`
}
