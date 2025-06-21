package clients

import (
	"context"
	"fmt"
	"io"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/domain"
)

//go:generate mockgen -source=weather_clients.go -destination=mocks/mock_weather_clients.go

type WeatherClient interface {
	GetAPICurrentWeather(ctx context.Context, city string) (*domain.WeatherResponse, error)
	GetAPIDayWeather(ctx context.Context, city string) (*domain.DayWeatherResponse, error)
}

type WeatherClients struct {
	WeatherAPI     WeatherClient
	VisualCrossing WeatherClient
}

func NewWeatherClients(thirdPartyCfg config.ThirdPartyConfig) *WeatherClients {
	return &WeatherClients{
		WeatherAPI:     NewWeatherAPIClient(thirdPartyCfg.WeatherAPIKey),
		VisualCrossing: NewVisualCrossingClient(thirdPartyCfg.VisualCrossingAPIKey),
	}
}

func closeBody(body io.Closer, errPtr *error) {
	if closeErr := body.Close(); closeErr != nil {
		if *errPtr != nil {
			*errPtr = fmt.Errorf("%w; failed to close response body: %w", *errPtr, closeErr)
		} else {
			*errPtr = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}
}
