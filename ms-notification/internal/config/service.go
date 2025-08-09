package config

import (
	commonCfg "common/config"
	"fmt"
)

type ConfigReader interface {
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
	// Load environment file
	if err := s.loadEnvironmentFile(environment); err != nil {
		return nil, fmt.Errorf("failed to load environment: %w", err)
	}

	// Read config file
	configDirPath := commonCfg.GetOriginalPath(configDir)
	if err := s.reader.ReadConfigFile(configDirPath, "main"); err != nil {
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
	s.postProcessor.ProcessConfig(cfg)

	return cfg, nil
}

func (s *ConfigService) loadEnvironmentFile(environment string) error {
	var envFile string

	switch environment {
	case TestEnvironment:
		return nil // No test envs
	case ProdEnvironment:
		envFile = "" // No production env file, using os envs
	default:
		envFile = "ms-notification/.env.dev"
	}

	return s.envLoader.LoadEnvFile(envFile)
}

func (s *ConfigService) setEnvironmentVariables(cfg *Config, environment string) {
	envVars := s.envLoader.GetRequiredEnvVars(environment)

	if environment != TestEnvironment {
		cfg.Logger.LoggerEnv = envVars["LOGG_ENV"]
		cfg.SMTP.Pass = envVars["SMTP_PASSWORD"]
		cfg.RabbitMQ.URL = envVars["RABBITMQ_URL"]
	}
}
