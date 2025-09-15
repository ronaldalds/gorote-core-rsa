package core

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID      `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return
}

type Tenant struct {
	BaseModel
	Name        string `gorm:"uniqueIndex;size:100" validate:"required,min=3,max=100" json:"name"`
	Description string `json:"description"`
	Users       []User `gorm:"many2many:users_tenants" json:"users"`
	Active      bool   `gorm:"default:true" json:"active"`
}

type Permission struct {
	BaseModel
	Code        string `gorm:"uniqueIndex;size:50" validate:"required,regexp=^[a-zA-Z0-9_]+$" json:"code"`
	Description string `json:"description"`
	Active      bool   `gorm:"default:true" json:"active"`
	Roles       []Role `gorm:"many2many:roles_permissions" json:"roles"`
}

type Role struct {
	BaseModel
	Name        string       `gorm:"uniqueIndex;size:100" validate:"required,min=3,max=100,regexp=^[a-zA-Z0-9._]+$" json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `gorm:"many2many:roles_permissions" json:"permissions"`
	Active      bool         `gorm:"default:true" json:"active"`
}

type User struct {
	BaseModel
	FirstName   string   `gorm:"size:50" validate:"omitempty,min=1,max=50" json:"first_name"`
	LastName    string   `gorm:"size:50" validate:"omitempty,max=50" json:"last_name"`
	Email       string   `gorm:"uniqueIndex" validate:"required,email" json:"email"`
	Password    string   `validate:"required" json:"-"`
	IsSuperUser bool     `gorm:"default:false" json:"is_super_user"`
	Phone1      *string  `gorm:"type:varchar(20)" validate:"omitempty,e164" json:"phone1"`
	Phone2      *string  `gorm:"type:varchar(20)" validate:"omitempty,e164" json:"phone2"`
	Roles       []Role   `gorm:"many2many:users_roles" json:"roles"`
	Tenants     []Tenant `gorm:"many2many:users_tenants" json:"tenants"`
	Active      bool     `gorm:"default:true" json:"active"`
}
