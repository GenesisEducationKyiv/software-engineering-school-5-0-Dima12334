package config

import (
	"github.com/spf13/viper"
)

type ViperConfigReader struct{}

func (r *ViperConfigReader) ReadConfigFile(configDirPath, configName string) error {
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDirPath)
	return viper.ReadInConfig()
}

func (r *ViperConfigReader) Unmarshal(cfg interface{}) error {
	return viper.Unmarshal(cfg)
}
