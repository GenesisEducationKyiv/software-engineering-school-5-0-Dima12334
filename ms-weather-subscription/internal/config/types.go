package config

import "time"

const (
	ProdEnvironment = "prod"
	DevEnvironment  = "dev"
	TestEnvironment = "test"
	ConfigsDir      = "ms-weather-subscription/configs"
)

type Config struct {
	Environment string
	HTTP        HTTPConfig     `mapstructure:"http_server"`
	Logger      LoggerConfig   `mapstructure:"logger"`
	DB          DatabaseConfig `mapstructure:"db"`
	Redis       RedisConfig
	ThirdParty  ThirdPartyConfig
}

type HTTPConfig struct {
	Host    string `mapstructure:"host"`
	Port    string `mapstructure:"port"`
	Scheme  string
	Domain  string
	BaseURL string

	ReadTimeout       time.Duration `mapstructure:"readTimeout"`
	ReadHeaderTimeout time.Duration `mapstructure:"readHeaderTimeout"`
	WriteTimeout      time.Duration `mapstructure:"writeTimeout"`
}

type ThirdPartyConfig struct {
	WeatherAPIKey          string
	VisualCrossingAPIKey   string
	NotificationServiceURL string
}

type LoggerConfig struct {
	LoggerEnv string
	FilePath  string `mapstructure:"file_path"`
}

type DatabaseConfig struct {
	Host           string
	Port           string
	User           string
	Password       string
	DBName         string
	SSLMode        string
	DSN            string
	MigrationsPath string `mapstructure:"migrationsPath"`
}

type RedisConfig struct {
	Address  string
	CacheDB  int
	Password string
}
