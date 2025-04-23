package core

import (
	"fmt"

	"gorm.io/gorm"
)

func (a *AppConfig) SaveUserAdmin() error {
	hashPassword, err := HashPassword(a.Super.SuperPass)
	if err != nil {
		return fmt.Errorf("failed to hash password: %s", err.Error())
	}
	if err := a.CoreGorm.
		FirstOrCreate(&User{
			FirstName:   a.Super.SuperName,
			LastName:    "Admin",
			Username:    a.Super.SuperUser,
			Email:       a.Super.SuperEmail,
			Password:    hashPassword,
			Active:      true,
			IsSuperUser: true,
			Phone1:      a.Super.SuperPhone,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (a *AppConfig) SavePermissions(permissions ...PermissionCode) error {
	for _, permission := range permissions {
		fmt.Println(permission)
		var p Permission
		if err := a.CoreGorm.Where("code = ?", string(permission)).First(&p).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				p = Permission{Code: string(permission), Name: string(permission)}
				if err := a.CoreGorm.Create(&p).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// Permission already exists, skip to the next one
			continue
		}
	}
	return nil
}
