package service

import (
	"context"
	"ms-weather-subscription/internal/config"
	"ms-weather-subscription/internal/domain"
	"ms-weather-subscription/internal/repository"
	"ms-weather-subscription/pkg/clients"
	"ms-weather-subscription/pkg/hash"
	"ms-weather-subscription/pkg/publisher"
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

type Deps struct {
	Repos              *repository.Repositories
	WeatherClient      clients.WeatherClient
	SubscriptionHasher hash.SubscriptionHasher
	HTTPConfig         config.HTTPConfig
	EmailPublisher     publisher.EmailPublisher
}

type Services struct {
	Subscriptions         Subscription
	Weather               Weather
	WeatherForecastSender WeatherForecastSender
}

func NewServices(deps Deps) *Services {
	weatherService := NewWeatherService(deps.WeatherClient)
	return &Services{
		Subscriptions: NewSubscriptionService(
			deps.HTTPConfig,
			deps.Repos.Subscription,
			deps.SubscriptionHasher,
			deps.EmailPublisher,
		),
		Weather: weatherService,
		WeatherForecastSender: NewWeatherForecastSenderService(
			deps.HTTPConfig,
			weatherService,
			deps.Repos.Subscription,
			deps.EmailPublisher,
		),
	}
}
