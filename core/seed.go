package core

import (
	"fmt"

	"gorm.io/gorm"
)

func (s *Service) saveUserAdmin() error {
	hashPassword, err := HashPassword(s.Super.SuperPass)
	if err != nil {
		return fmt.Errorf("failed to hash password: %s", err.Error())
	}
	if err := s.DB.
		FirstOrCreate(&User{
			FirstName:   s.Super.SuperName,
			LastName:    "Admin",
			Email:       s.Super.SuperEmail,
			Password:    hashPassword,
			Active:      true,
			IsSuperUser: true,
			Phone1:      s.Super.SuperPhone,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (s *Service) savePermissions(permissions ...PermissionCode) error {
	for _, permission := range permissions {
		var p Permission
		if err := s.DB.Where("code = ?", string(permission)).First(&p).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				p = Permission{Code: string(permission), Name: string(permission)}
				if err := s.DB.Create(&p).Error; err != nil {
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
