package service

import (
	"context"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/internal/repository"
	"weather_forecast_sub/pkg/clients"
	"weather_forecast_sub/pkg/email"
	"weather_forecast_sub/pkg/hash"
)

//go:generate mockgen -source=service.go -destination=mocks/mock_service.go

type CreateSubscriptionInput struct {
	Email     string `json:"email"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
}

type Subscription interface {
	Create(ctx context.Context, inp CreateSubscriptionInput) error
	Confirm(ctx context.Context, token string) error
	Delete(ctx context.Context, token string) error

	SendHourlyWeatherForecast() error
	SendDailyWeatherForecast() error
}

type Weather interface {
	GetCurrentWeather(city string) (*clients.WeatherResponse, error)
	GetDayWeather(city string) (*clients.DayWeatherResponse, error)
}

type ConfirmationEmailInput struct {
	Email string
	Token string
}

type WeatherForecastDailyEmailInput struct {
	Subscription domain.Subscription
	Weather      *clients.DayWeatherResponse
	Date         string
}

type WeatherForecastHourlyEmailInput struct {
	Subscription domain.Subscription
	Weather      *clients.WeatherResponse
	Date         string
}

type Emails interface {
	SendConfirmationEmail(ConfirmationEmailInput) error
	SendWeatherForecastDailyEmail(WeatherForecastDailyEmailInput) error
	SendWeatherForecastHourlyEmail(WeatherForecastHourlyEmailInput) error
}

type Deps struct {
	Repos       *repository.Repositories
	Clients     *clients.Clients
	EmailHasher hash.EmailHasher
	EmailSender email.Sender
	EmailConfig config.EmailConfig
	HTTPConfig  config.HTTPConfig
}

type Services struct {
	Subscriptions Subscription
	Weather       Weather
}

func NewServices(deps Deps) *Services {
	emailsService := NewEmailsService(deps.EmailSender, deps.EmailConfig, deps.HTTPConfig)
	weatherService := NewWeatherService(deps.Clients.WeatherAPI)
	return &Services{
		Subscriptions: NewSubscriptionService(
			deps.Repos.Subscription,
			deps.EmailHasher,
			deps.EmailSender,
			deps.EmailConfig,
			deps.HTTPConfig,
			emailsService,
			weatherService,
		),
		Weather: weatherService,
	}
}
