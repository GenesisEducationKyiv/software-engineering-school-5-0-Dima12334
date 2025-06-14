package config

import "fmt"

type DefaultConfigPostProcessor struct{}

func NewDefaultConfigPostProcessor() *DefaultConfigPostProcessor {
	return &DefaultConfigPostProcessor{}
}

func (p *DefaultConfigPostProcessor) ProcessHTTPConfig(cfg *Config) error {
	if cfg.Environment == ProdEnvironment {
		cfg.HTTP.Scheme = "https"
		cfg.HTTP.Domain = cfg.HTTP.Host
	} else {
		cfg.HTTP.Scheme = "http"
		cfg.HTTP.Domain = fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port)
	}
	cfg.HTTP.BaseURL = fmt.Sprintf("%s://%s", cfg.HTTP.Scheme, cfg.HTTP.Domain)
	return nil
}

func (p *DefaultConfigPostProcessor) ProcessDatabaseConfig(cfg *Config) error {
	cfg.DB.DSN = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.DBName,
		cfg.DB.SSLMode,
	)
	return nil
}
