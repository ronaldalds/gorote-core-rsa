package core

func (config *AppConfig) PreReady() error {
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

func (s *Service) PosReady() error {
	if s.Super != nil {
		if err := s.saveUserAdmin(); err != nil {
			return err
		}
	}

	if err := s.savePermissions(
		PermissionSuperUser,
		PermissionCreateUser,
		PermissionViewUser,
		PermissionUpdateUser,
		PermissionCreatePermission,
		PermissionViewPermission,
		PermissionUpdatePermission,
		PermissionCreateRole,
		PermissionViewRole,
		PermissionUpdateRole,
	); err != nil {
		return err
	}
	return nil
}
