package clients

import (
	"context"
	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/pkg/logger"
)

//go:generate mockgen -source=weather_clients.go -destination=mocks/mock_weather_clients.go

type WeatherClient interface {
	GetAPICurrentWeather(ctx context.Context, city string) (*domain.WeatherResponse, error)
	GetAPIDayWeather(ctx context.Context, city string) (*domain.DayWeatherResponse, error)
}

type ChainWeatherClient struct {
	// The first client in the list is considered the primary client.
	clients []WeatherClient
}

func NewChainWeatherClient(clients []WeatherClient) *ChainWeatherClient {
	return &ChainWeatherClient{
		clients: clients,
	}
}

func (c *ChainWeatherClient) GetAPICurrentWeather(
	ctx context.Context, city string,
) (*domain.WeatherResponse, error) {
	var lastErr error

	for _, client := range c.clients {
		response, err := client.GetAPICurrentWeather(ctx, city)
		if err == nil {
			return response, nil
		}
		logger.Errorf("error from client %T: %s", client, err.Error())
		lastErr = err
	}

	logger.Errorf("all weather clients failed for city %s: %s", city, lastErr.Error())
	return nil, lastErr
}

func (c *ChainWeatherClient) GetAPIDayWeather(
	ctx context.Context, city string,
) (*domain.DayWeatherResponse, error) {
	var lastErr error

	for _, client := range c.clients {
		response, err := client.GetAPIDayWeather(ctx, city)
		if err == nil {
			return response, nil
		}
		logger.Errorf("error from client %T: %s", client, err.Error())
		lastErr = err
	}

	logger.Errorf("all weather clients failed for city %s: %s", city, lastErr.Error())
	return nil, lastErr
}
