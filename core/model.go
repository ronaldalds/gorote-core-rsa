package core

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (base *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if base.ID == uuid.Nil {
		base.ID = uuid.New()
	}
	return
}

type Tenant struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;size:100" validate:"required,min=3,max=100"`
	Description string
	Users       []User `gorm:"many2many:users_tenants"`
	Active      bool   `gorm:"default:true"`
}

type Permission struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;size:100" validate:"required,min=3,max=100"`
	Code        string `gorm:"uniqueIndex;size:50" validate:"required,regexp=^[a-zA-Z0-9_]+$"`
	Description string
	Active      bool `gorm:"default:true"`
}

type Role struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;size:100" validate:"required,min=3,max=100,regexp=^[a-zA-Z0-9._]+$"`
	Description string
	Permissions []Permission `gorm:"many2many:roles_permissions"`
	Active      bool         `gorm:"default:true"`
}

type User struct {
	gorm.Model
	FirstName   string   `gorm:"size:50" validate:"required,min=1,max=50"`
	LastName    string   `gorm:"size:50" validate:"omitempty,max=50"`
	Email       string   `gorm:"uniqueIndex" validate:"required,email"`
	Password    string   `validate:"required"`
	IsSuperUser bool     `gorm:"default:false"`
	Phone1      string   `gorm:"type:varchar(20)" validate:"required,e164"`
	Phone2      *string  `gorm:"type:varchar(20)" validate:"omitempty,e164"`
	Roles       []Role   `gorm:"many2many:users_roles"`
	Tenants     []Tenant `gorm:"many2many:users_tenants"`
	Active      bool     `gorm:"default:true"`
}
