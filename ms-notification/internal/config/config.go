package config

func NewDefaultConfigService() *ConfigService {
	reader := &ViperConfigReader{}
	envLoader := &GodotenvLoader{}
	postProcessor := &DefaultConfigPostProcessor{}

	return NewConfigService(reader, envLoader, postProcessor)
}

func Init(configDir, environment string) (*Config, error) {
	service := NewDefaultConfigService()
	return service.LoadConfig(configDir, environment)
}
