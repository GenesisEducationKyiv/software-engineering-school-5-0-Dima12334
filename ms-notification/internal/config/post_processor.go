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

func (p *DefaultConfigPostProcessor) processEmailConfig(cfg *Config) {
	cfg.Email.Templates.Confirmation = commonCfg.GetOriginalPath(cfg.Email.Templates.Confirmation)
	cfg.Email.Templates.WeatherForecastDaily = commonCfg.GetOriginalPath(
		cfg.Email.Templates.WeatherForecastDaily,
	)
	cfg.Email.Templates.WeatherForecastHourly = commonCfg.GetOriginalPath(
		cfg.Email.Templates.WeatherForecastHourly,
	)
}

func (p *DefaultConfigPostProcessor) ProcessConfig(cfg *Config) {
	p.processHTTPConfig(cfg)
	p.processEmailConfig(cfg)
}
