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
	if environ == TestEnvironment {
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
	case TestEnvironment:
		err = godotenv.Load("../.env")
	case DevEnvironment:
		err = godotenv.Load()
	case ProdEnvironment:
		// Do nothing
	default:
		err = godotenv.Load()
	}

	if err != nil {
		log.Fatalf("error loading .env file")
	}

	// Load and validate required variables
	requiredVars := map[string]*string{
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

	for key, ptr := range requiredVars {
		val, exists := os.LookupEnv(key)
		if !exists || val == "" {
			log.Fatalf("environment variable %q is required but not set or empty", key)
		}
		*ptr = val
	}

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
