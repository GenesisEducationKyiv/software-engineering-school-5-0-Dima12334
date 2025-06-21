package clients

import (
	"context"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/domain"
)

//go:generate mockgen -source=clients.go -destination=mocks/mock_clients.go

type WeatherClient interface {
	GetAPICurrentWeather(ctx context.Context, city string) (*domain.WeatherResponse, error)
	GetAPIDayWeather(ctx context.Context, city string) (*domain.DayWeatherResponse, error)
}

type Clients struct {
	WeatherAPI WeatherClient
}

func NewClients(thirdPartyCfg config.ThirdPartyConfig) *Clients {
	return &Clients{
		WeatherAPI: NewWeatherAPIClient(thirdPartyCfg.WeatherAPIKey),
	}
}
