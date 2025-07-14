package config

import (
	"fmt"
	"strconv"
)

type ConfigReader interface {
	SetDefaults()
	ReadConfigFile(configDirPath, configName string) error
	Unmarshal(cfg interface{}) error
}

type EnvLoader interface {
	LoadEnvFile(filePath string) error
	GetRequiredEnvVars(environment string) map[string]string
}

type ConfigPostProcessor interface {
	ProcessConfig(cfg *Config)
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
	configDirPath := GetOriginalPath(configDir)
	if err := s.reader.ReadConfigFile(configDirPath, "main"); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Create and unmarshal config
	cfg := &Config{Environment: environment}
	if err := s.reader.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Load and set environment variables
	if err := s.setEnvironmentVariables(cfg, environment); err != nil {
		return nil, fmt.Errorf("failed to set environment variables: %w", err)
	}

	// Post-process derived values
	s.postProcessor.ProcessConfig(cfg)

	return cfg, nil
}

func (s *ConfigService) loadEnvironmentFile(environment string) error {
	var envFile string

	switch environment {
	case TestEnvironment:
		envFile = ".env.test"
	case ProdEnvironment:
		envFile = "" // No env file for production
	default:
		envFile = ".env.dev"
	}

	return s.envLoader.LoadEnvFile(envFile)
}

func (s *ConfigService) setEnvironmentVariables(cfg *Config, environment string) error {
	envVars := s.envLoader.GetRequiredEnvVars(environment)

	// Map environment variables to config fields
	cfg.DB.Host = envVars["DB_HOST"]
	cfg.DB.Port = envVars["DB_PORT"]
	cfg.DB.User = envVars["DB_USER"]
	cfg.DB.Password = envVars["DB_PASSWORD"]
	cfg.DB.DBName = envVars["DB_NAME"]
	cfg.DB.SSLMode = envVars["DB_SSLMODE"]

	cfg.Redis.Address = envVars["REDIS_ADDRESS"]
	cacheDB, err := strconv.Atoi(envVars["REDIS_CACHE_DB"])
	if err != nil {
		return fmt.Errorf("REDIS_CACHE_DB must be integer: %v", err)
	}
	cfg.Redis.CacheDB = cacheDB
	cfg.Redis.Password = envVars["REDIS_PASSWORD"]

	if environment != TestEnvironment {
		cfg.Logger.LoggerEnv = envVars["LOGG_ENV"]
		cfg.HTTP.Host = envVars["HTTP_HOST"]
		cfg.ThirdParty.WeatherAPIKey = envVars["WEATHER_API_KEY"]
		cfg.ThirdParty.VisualCrossingAPIKey = envVars["VISUAL_CROSSING_API_KEY"]
		cfg.ThirdParty.NotificationServiceURL = envVars["NOTIFICATION_SERVICE_URL"]
	}

	return nil
}
