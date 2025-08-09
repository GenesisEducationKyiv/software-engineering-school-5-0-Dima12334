package config

import (
	commonCfg "common/config"
)

type DefaultConfigPostProcessor struct{}

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
	p.processEmailConfig(cfg)
}
