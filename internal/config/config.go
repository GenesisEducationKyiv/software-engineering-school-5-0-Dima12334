package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const (
	ProdEnvironment = "prod"
	DevEnvironment  = "dev"
	TestEnvironment = "test"

	ConfigsDir = "configs"

	defaultHTTPPort       = "8080"
	defaultMigrationsPath = "file://migrations"
)

type (
	Config struct {
		Environment string
		HTTP        HTTPConfig `mapstructure:"http_server"`
		Logger      LoggerConfig
		DB          DatabaseConfig `mapstructure:"db"`
		ThirdParty  ThirdPartyConfig
		SMTP        SMTPConfig  `mapstructure:"smtp"`
		Email       EmailConfig `mapstructure:"email"`
	}

	HTTPConfig struct {
		Host    string `mapstructure:"host"`
		Port    string `mapstructure:"port"`
		Scheme  string
		Domain  string
		BaseURL string

		ReadTimeout       time.Duration `mapstructure:"readTimeout"`
		ReadHeaderTimeout time.Duration `mapstructure:"readHeaderTimeout"`
		WriteTimeout      time.Duration `mapstructure:"writeTimeout"`
	}

	ThirdPartyConfig struct {
		WeatherAPIKey string
	}

	LoggerConfig struct {
		LoggerEnv string
	}

	DatabaseConfig struct {
		Host           string
		Port           string
		User           string
		Password       string
		DBName         string
		SSLMode        string
		DSN            string
		MigrationsPath string `mapstructure:"migrationsPath"`
	}

	SMTPConfig struct {
		Host     string `mapstructure:"host"`
		Port     int    `mapstructure:"port"`
		From     string `mapstructure:"from"`
		FromName string `mapstructure:"from_name"`
		Pass     string
	}

	EmailConfig struct {
		Templates EmailTemplates
		Subjects  EmailSubjects
	}

	EmailTemplates struct {
		Confirmation          string `mapstructure:"confirmation_email"`
		WeatherForecastDaily  string `mapstructure:"weather_forecast_daily"`
		WeatherForecastHourly string `mapstructure:"weather_forecast_hourly"`
	}

	EmailSubjects struct {
		Confirmation    string `mapstructure:"confirmation_email"`
		WeatherForecast string `mapstructure:"weather_forecast"`
	}
)

func Init(configDir, environ string) (*Config, error) {
	populateDefaults()

	if err := parseConfigFile(configDir, environ); err != nil {
		return nil, err
	}

	var cfg Config

	cfg.Environment = environ

	if err := unmarshalConfig(&cfg); err != nil {
		return nil, err
	}

	setFormEnv(&cfg)

	if cfg.Environment == ProdEnvironment {
		cfg.HTTP.Scheme = "https"
		cfg.HTTP.Domain = cfg.HTTP.Host
	} else {
		cfg.HTTP.Scheme = "http"
		cfg.HTTP.Domain = fmt.Sprintf("%s:%s", cfg.HTTP.Host, cfg.HTTP.Port)
	}
	cfg.HTTP.BaseURL = fmt.Sprintf("%s://%s", cfg.HTTP.Scheme, cfg.HTTP.Domain)

	cfg.DB.DSN = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.DBName,
		cfg.DB.SSLMode,
	)
	return &cfg, nil
}

func unmarshalConfig(cfg *Config) error {
	return viper.Unmarshal(cfg)
}

func parseConfigFile(configDir, environ string) error {
	if environ == TestEnvironment {
		viper.SetConfigName("test")
	} else {
		viper.SetConfigName("main")
	}
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	return viper.ReadInConfig()
}

func setFormEnv(cfg *Config) {
	var err error

	switch cfg.Environment {
	case TestEnvironment:
		err = godotenv.Load("../.env.dev.test")
	case ProdEnvironment:
		// Do nothing
	default:
		err = godotenv.Load()
	}

	if err != nil {
		log.Fatalf("error loading .env.dev file")
	}

	// Load and validate required variables
	var requiredVars map[string]*string
	if cfg.Environment == TestEnvironment {
		requiredVars = map[string]*string{
			"DB_HOST":     &cfg.DB.Host,
			"DB_PORT":     &cfg.DB.Port,
			"DB_USER":     &cfg.DB.User,
			"DB_PASSWORD": &cfg.DB.Password,
			"DB_NAME":     &cfg.DB.DBName,
			"DB_SSLMODE":  &cfg.DB.SSLMode,
		}
	} else {
		requiredVars = map[string]*string{
			"LOGG_ENV":        &cfg.Logger.LoggerEnv,
			"WEATHER_API_KEY": &cfg.ThirdParty.WeatherAPIKey,
			"HTTP_HOST":       &cfg.HTTP.Host,
			"SMTP_PASSWORD":   &cfg.SMTP.Pass,

			"DB_HOST":     &cfg.DB.Host,
			"DB_PORT":     &cfg.DB.Port,
			"DB_USER":     &cfg.DB.User,
			"DB_PASSWORD": &cfg.DB.Password,
			"DB_NAME":     &cfg.DB.DBName,
			"DB_SSLMODE":  &cfg.DB.SSLMode,
		}
	}

	for key, ptr := range requiredVars {
		val, exists := os.LookupEnv(key)
		if !exists || val == "" {
			log.Fatalf("environment variable %q is required but not set or empty", key)
		}
		*ptr = val
	}
}

func populateDefaults() {
	viper.SetDefault("http_server.port", defaultHTTPPort)
	viper.SetDefault("db.migrationsPath", defaultMigrationsPath)
}
