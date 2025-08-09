package config

import (
	commonCfg "common/config"

	"fmt"
)

type DefaultConfigPostProcessor struct{}

func (p *DefaultConfigPostProcessor) processHTTPConfig(cfg *Config) {
	if cfg.Environment == ProdEnvironment {
		cfg.HTTP.Scheme = "https"
		cfg.HTTP.Domain = cfg.HTTP.Host
	} else {
		cfg.HTTP.Scheme = "http"
		cfg.HTTP.Domain = fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port)
	}
	cfg.HTTP.BaseURL = fmt.Sprintf("%s://%s", cfg.HTTP.Scheme, cfg.HTTP.Domain)
}

func (p *DefaultConfigPostProcessor) processDatabaseConfig(cfg *Config) {
	cfg.DB.DSN = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DBName, cfg.DB.SSLMode,
	)

	migrationsPath := "file://" + commonCfg.GetOriginalPath("ms-weather-subscription/migrations")
	cfg.DB.MigrationsPath = migrationsPath
}

func (p *DefaultConfigPostProcessor) ProcessConfig(cfg *Config) {
	p.processHTTPConfig(cfg)
	p.processDatabaseConfig(cfg)
}
