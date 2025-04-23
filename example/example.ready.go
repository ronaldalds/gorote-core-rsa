package example

func (config *AppConfig) PreReady() error {
	// Executar as Migrations
	if err := config.ExampleDB.AutoMigrate(); err != nil {
		return err
	}
	return nil
}

func (s *Service) PosReady() error {
	return nil
}
