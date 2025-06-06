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
	testEnvironment       = "test"
	devEnvironment        = "dev"
	prodEnvironment       = "prod"
	defaultHTTPPort       = "8080"
	defaultMigrationsPath = "file://migrations"
)

type (
	Config struct {
		Environment string
		HTTP        HTTPConfig
		Logger      LoggerConfig
		DB          DatabaseConfig
		TestDB      DatabaseConfig
		ThirdParty  ThirdPartyConfig
		SMTP        SMTPConfig
		Email       EmailConfig
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

	if cfg.Environment == prodEnvironment {
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

	cfg.TestDB.DSN = fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.TestDB.User,
		cfg.TestDB.Password,
		cfg.TestDB.Host,
		cfg.TestDB.Port,
		cfg.TestDB.DBName,
		cfg.TestDB.SSLMode,
	)

	return &cfg, nil
}

func unmarshalConfig(cfg *Config) error {
	if err := viper.UnmarshalKey("http_server", &cfg.HTTP); err != nil {
		return err
	}
	if err := viper.UnmarshalKey("db", &cfg.DB); err != nil {
		return err
	}
	if err := viper.UnmarshalKey("db", &cfg.TestDB); err != nil {
		return err
	}
	if err := viper.UnmarshalKey("smtp", &cfg.SMTP); err != nil {
		return err
	}
	if err := viper.UnmarshalKey("email", &cfg.Email); err != nil {
		return err
	}
	return nil
}

func parseConfigFile(configDir, environ string) error {
	if environ == testEnvironment {
		viper.SetConfigName("test")
	} else {
		viper.SetConfigName("main")
	}
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

func setFormEnv(cfg *Config) {
	var err error

	switch cfg.Environment {
	case testEnvironment:
		err = godotenv.Load("../.env")
	case devEnvironment:
		err = godotenv.Load()
	case prodEnvironment:
		// Do nothing
	default:
		err = godotenv.Load()
	}

	if err != nil {
		log.Fatalf("error loading .env file")
	}

	cfg.Logger.LoggerEnv = os.Getenv("LOGG_ENV")

	cfg.ThirdParty.WeatherAPIKey = os.Getenv("WEATHER_API_KEY")

	cfg.HTTP.Host = os.Getenv("HTTP_HOST")

	cfg.SMTP.Pass = os.Getenv("SMTP_PASSWORD")

	cfg.DB.Host = os.Getenv("DB_HOST")
	cfg.DB.Port = os.Getenv("DB_PORT")
	cfg.DB.User = os.Getenv("DB_USER")
	cfg.DB.Password = os.Getenv("DB_PASSWORD")
	cfg.DB.DBName = os.Getenv("DB_NAME")
	cfg.DB.SSLMode = os.Getenv("DB_SSLMODE")

	cfg.TestDB.Host = os.Getenv("TEST_DB_HOST")
	cfg.TestDB.Port = os.Getenv("TEST_DB_PORT")
	cfg.TestDB.User = os.Getenv("TEST_DB_USER")
	cfg.TestDB.Password = os.Getenv("TEST_DB_PASSWORD")
	cfg.TestDB.DBName = os.Getenv("TEST_DB_NAME")
	cfg.TestDB.SSLMode = os.Getenv("TEST_DB_SSLMODE")
}

func populateDefaults() {
	viper.SetDefault("http_server.port", defaultHTTPPort)
	viper.SetDefault("db.migrationsPath", defaultMigrationsPath)
}
