package config

import "time"

const (
	ProdEnvironment = "prod"
	DevEnvironment  = "dev"
	TestEnvironment = "test"
	ConfigsDir      = "configs"
)

type Config struct {
	Environment string
	HTTP        HTTPConfig     `mapstructure:"http_server"`
	Logger      LoggerConfig   `mapstructure:"logger"`
	DB          DatabaseConfig `mapstructure:"db"`
	Redis       RedisConfig
	ThirdParty  ThirdPartyConfig
	SMTP        SMTPConfig  `mapstructure:"smtp"`
	Email       EmailConfig `mapstructure:"email"`
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
	WeatherAPIKey        string
	VisualCrossingAPIKey string
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

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	From     string `mapstructure:"from"`
	FromName string `mapstructure:"from_name"`
	Pass     string
}

type EmailConfig struct {
	Templates EmailTemplates
	Subjects  EmailSubjects
}

type EmailTemplates struct {
	Confirmation          string `mapstructure:"confirmation_email"`
	WeatherForecastDaily  string `mapstructure:"weather_forecast_daily"`
	WeatherForecastHourly string `mapstructure:"weather_forecast_hourly"`
}

type EmailSubjects struct {
	Confirmation    string `mapstructure:"confirmation_email"`
	WeatherForecast string `mapstructure:"weather_forecast"`
}
