package config

func NewDefaultConfigService() *ConfigService {
	reader := NewViperConfigReader()
	envLoader := NewGodotenvLoader()
	postProcessor := NewDefaultConfigPostProcessor()

	return NewConfigService(reader, envLoader, postProcessor)
}

func Init(configDir, environ string) (*Config, error) {
	service := NewDefaultConfigService()
	return service.LoadConfig(configDir, environ)
}
