package core

func (config *AppConfig) PreReady() error {
	// Exe. Migrations
	if err := config.DB.AutoMigrate(
		&User{},
		&Role{},
		&Permission{},
	); err != nil {
		return err
	}
	return nil
}

func (s *Service) PosReady() error {
	if err := s.Seed(s.Super); err != nil {
		return err
	}
	return nil
}
