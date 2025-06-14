package config

import "fmt"

type ConfigReader interface {
	SetDefaults()
	ReadConfigFile(configDir, configName string) error
	Unmarshal(cfg interface{}) error
}

type EnvLoader interface {
	LoadEnvFile(filePath string) error
	GetRequiredEnvVars(environment string) map[string]string
}

type ConfigPostProcessor interface {
	ProcessHTTPConfig(cfg *Config) error
	ProcessDatabaseConfig(cfg *Config) error
}

type ConfigService struct {
	reader        ConfigReader
	envLoader     EnvLoader
	postProcessor ConfigPostProcessor
}

func NewConfigService(
	reader ConfigReader,
	envLoader EnvLoader,
	postProcessor ConfigPostProcessor,
) *ConfigService {
	return &ConfigService{
		reader:        reader,
		envLoader:     envLoader,
		postProcessor: postProcessor,
	}
}

func (s *ConfigService) LoadConfig(configDir, environment string) (*Config, error) {
	s.reader.SetDefaults()

	// Load environment file
	if err := s.loadEnvironmentFile(environment); err != nil {
		return nil, fmt.Errorf("failed to load environment: %w", err)
	}

	// Read config file
	configName := s.getConfigFileName(environment)
	if err := s.reader.ReadConfigFile(configDir, configName); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Create and unmarshal config
	cfg := &Config{Environment: environment}
	if err := s.reader.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Load and set environment variables
	s.setEnvironmentVariables(cfg, environment)

	// Post-process derived values
	if err := s.postProcessor.ProcessHTTPConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to process HTTP config: %w", err)
	}

	if err := s.postProcessor.ProcessDatabaseConfig(cfg); err != nil {
		return nil, fmt.Errorf("failed to process database config: %w", err)
	}

	return cfg, nil
}

func (s *ConfigService) loadEnvironmentFile(environment string) error {
	var envFile string

	switch environment {
	case TestEnvironment:
		envFile = "../.env.test"
	case ProdEnvironment:
		envFile = "" // No env file for production
	default:
		envFile = "./.env.dev"
	}

	return s.envLoader.LoadEnvFile(envFile)
}

func (s *ConfigService) getConfigFileName(environment string) string {
	if environment == TestEnvironment {
		return "test"
	}
	return "main"
}

func (s *ConfigService) setEnvironmentVariables(cfg *Config, environment string) {
	envVars := s.envLoader.GetRequiredEnvVars(environment)

	// Map environment variables to config fields
	cfg.DB.Host = envVars["DB_HOST"]
	cfg.DB.Port = envVars["DB_PORT"]
	cfg.DB.User = envVars["DB_USER"]
	cfg.DB.Password = envVars["DB_PASSWORD"]
	cfg.DB.DBName = envVars["DB_NAME"]
	cfg.DB.SSLMode = envVars["DB_SSLMODE"]

	if environment != TestEnvironment {
		cfg.Logger.LoggerEnv = envVars["LOGG_ENV"]
		cfg.HTTP.Host = envVars["HTTP_HOST"]
		cfg.ThirdParty.WeatherAPIKey = envVars["WEATHER_API_KEY"]
		cfg.SMTP.Pass = envVars["SMTP_PASSWORD"]
	}
}
