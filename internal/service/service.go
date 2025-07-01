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

type Subscription interface {
	Create(ctx context.Context, inp domain.CreateSubscriptionInput) error
	Confirm(ctx context.Context, token string) error
	Delete(ctx context.Context, token string) error
}

type WeatherForecastSender interface {
	SendHourlyWeatherForecast(ctx context.Context) error
	SendDailyWeatherForecast(ctx context.Context) error
}

type Weather interface {
	GetCurrentWeather(ctx context.Context, city string) (*domain.WeatherResponse, error)
	GetDayWeather(ctx context.Context, city string) (*domain.DayWeatherResponse, error)
}

type WeatherResponseType interface {
	*domain.WeatherResponse | *domain.DayWeatherResponse
}

type ConfirmationEmailInput struct {
	Email string
	Token string
}

type WeatherForecastEmailInput[T WeatherResponseType] struct {
	Subscription domain.Subscription
	Weather      T
	Date         string
}

type SubscriptionEmails interface {
	SendConfirmationEmail(ConfirmationEmailInput) error
}

type WeatherEmails interface {
	SendWeatherForecastDailyEmail(WeatherForecastEmailInput[*domain.DayWeatherResponse]) error
	SendWeatherForecastHourlyEmail(WeatherForecastEmailInput[*domain.WeatherResponse]) error
}

type Deps struct {
	Repos              *repository.Repositories
	WeatherClient      clients.WeatherClient
	SubscriptionHasher hash.SubscriptionHasher
	EmailSender        email.Sender
	EmailConfig        config.EmailConfig
	HTTPConfig         config.HTTPConfig
}

type Services struct {
	Subscriptions         Subscription
	Weather               Weather
	WeatherForecastSender WeatherForecastSender
}

func NewServices(deps Deps) *Services {
	emailsService := NewEmailsService(deps.EmailSender, deps.EmailConfig, deps.HTTPConfig)
	weatherService := NewWeatherService(deps.WeatherClient)
	return &Services{
		Subscriptions: NewSubscriptionService(
			deps.Repos.Subscription,
			deps.SubscriptionHasher,
			emailsService,
		),
		Weather: weatherService,
		WeatherForecastSender: NewWeatherForecastSenderService(
			emailsService,
			weatherService,
			deps.Repos.Subscription,
		),
	}
}
