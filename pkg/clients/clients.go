package clients

import (
	"weather_forecast_sub/internal/config"
)

//go:generate mockgen -source=clients.go -destination=mocks/mock_clients.go

type WeatherClient interface {
	GetAPICurrentWeather(city string) (*WeatherResponse, error)
	GetAPIDayWeather(city string) (*DayWeatherResponse, error)
}

type Clients struct {
	WeatherAPI WeatherClient
}

func NewClients(thirdPartyCfg config.ThirdPartyConfig) *Clients {
	return &Clients{
		WeatherAPI: NewWeatherAPIClient(thirdPartyCfg.WeatherAPIKey),
	}
}
