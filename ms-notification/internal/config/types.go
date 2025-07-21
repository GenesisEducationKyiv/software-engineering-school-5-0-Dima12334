package config

const (
	ProdEnvironment = "prod"
	DevEnvironment  = "dev"
	TestEnvironment = "test"
	ConfigsDir      = "ms-notification/configs"
)

type Config struct {
	Environment string
	Logger      LoggerConfig `mapstructure:"logger"`
	SMTP        SMTPConfig   `mapstructure:"smtp"`
	Email       EmailConfig  `mapstructure:"email"`
	RabbitMQ    RabbitMQConfig
}

type LoggerConfig struct {
	LoggerEnv string
	FilePath  string `mapstructure:"file_path"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	From     string `mapstructure:"from"`
	FromName string `mapstructure:"from_name"`
	Pass     string
}

type RabbitMQConfig struct {
	URL string
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
