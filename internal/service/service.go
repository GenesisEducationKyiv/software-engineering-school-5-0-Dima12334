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
}

type CronJobs interface {
	SendHourlyWeatherForecast(ctx context.Context) error
	SendDailyWeatherForecast(ctx context.Context) error
}

type Weather interface {
	GetCurrentWeather(ctx context.Context, city string) (*clients.WeatherResponse, error)
	GetDayWeather(ctx context.Context, city string) (*clients.DayWeatherResponse, error)
}

type WeatherResponseType interface {
	*clients.WeatherResponse | *clients.DayWeatherResponse
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

type Emails interface {
	SendConfirmationEmail(ConfirmationEmailInput) error
	SendWeatherForecastDailyEmail(WeatherForecastEmailInput[*clients.DayWeatherResponse]) error
	SendWeatherForecastHourlyEmail(WeatherForecastEmailInput[*clients.WeatherResponse]) error
}

type Deps struct {
	Repos              *repository.Repositories
	Clients            *clients.Clients
	SubscriptionHasher hash.SubscriptionHasher
	EmailSender        email.Sender
	EmailConfig        config.EmailConfig
	HTTPConfig         config.HTTPConfig
}

type Services struct {
	Subscriptions Subscription
	Weather       Weather
	CronJobs      CronJobs
}

func NewServices(deps Deps) *Services {
	emailsService := NewEmailsService(deps.EmailSender, deps.EmailConfig, deps.HTTPConfig)
	weatherService := NewWeatherService(deps.Clients.WeatherAPI)
	return &Services{
		Subscriptions: NewSubscriptionService(
			deps.Repos.Subscription,
			deps.SubscriptionHasher,
			deps.EmailSender,
			deps.EmailConfig,
			deps.HTTPConfig,
			emailsService,
			weatherService,
		),
		Weather: weatherService,
		CronJobs: NewCronJobsService(
			emailsService,
			weatherService,
			deps.Repos.Subscription,
		),
	}
}
