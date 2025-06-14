package config

import (
	"github.com/spf13/viper"
)

const (
	defaultHTTPPort       = "8080"
	defaultMigrationsPath = "file://migrations"
)

type ViperConfigReader struct{}

func (r *ViperConfigReader) SetDefaults() {
	viper.SetDefault("http_server.port", defaultHTTPPort)
	viper.SetDefault("db.migrationsPath", defaultMigrationsPath)
}

func (r *ViperConfigReader) ReadConfigFile(configDir, configName string) error {
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	return viper.ReadInConfig()
}

func (r *ViperConfigReader) Unmarshal(cfg interface{}) error {
	return viper.Unmarshal(cfg)
}
