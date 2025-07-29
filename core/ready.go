package core

import (
	"fmt"

	"github.com/ronaldalds/gorote-core-rsa/gorote"
	"gorm.io/gorm"
)

func migrate(config AppConfig) error {
	if err := config.DB.AutoMigrate(
		&User{},
		&Role{},
		&Permission{},
		&Tenant{},
	); err != nil {
		return err
	}
	return nil
}

func saveUserAdmin(config AppConfig) error {
	hashPassword, err := gorote.HashPassword(config.Super.SuperPass)
	if err != nil {
		return fmt.Errorf("failed to hash password: %s", err.Error())
	}
	if err := config.DB.
		FirstOrCreate(&User{
			FirstName:   config.Super.SuperName,
			LastName:    "Admin",
			Email:       config.Super.SuperEmail,
			Password:    hashPassword,
			Active:      true,
			IsSuperUser: true,
			Phone1:      config.Super.SuperPhone,
		}).Error; err != nil {
		return err
	}
	return nil
}

func savePermissions(config AppConfig) error {
	permissions := []PermissionCode{
		PermissionSuperUser, PermissionCreateUser,
		PermissionViewUser, PermissionUpdateUser,
		PermissionCreatePermission, PermissionViewPermission,
		PermissionUpdatePermission, PermissionCreateRole,
		PermissionViewRole, PermissionUpdateRole,
	}
	for _, permission := range permissions {
		var p Permission
		if err := config.DB.Where("code = ?", string(permission)).First(&p).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				p = Permission{Code: string(permission), Name: string(permission)}
				if err := config.DB.Create(&p).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			continue
		}
	}
	return nil
}
