package core

func (config *AppConfig) PreReady() error {
	// Exe. Migrations
	if err := config.CoreGorm.AutoMigrate(
		&User{},
		&Role{},
		&Permission{},
	); err != nil {
		return err
	}
	// Exe. Seeds
	if config.Super != nil {
		if err := config.SaveUserAdmin(); err != nil {
			return err
		}
	}

	if err := config.SavePermissions(
		PermissionSuperUser,
		PermissionCreateUser,
		PermissionViewUser,
		PermissionUpdateUser,
		PermissionEditePermissionsUser,
		PermissionCreateRole,
		PermissionViewRole,
		PermissionUpdateRole,
	); err != nil {
		return err
	}
	return nil
}

func (s *Service) PosReady() error {
	return nil
}
