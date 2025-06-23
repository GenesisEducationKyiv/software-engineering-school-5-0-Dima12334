package clients

import (
	"context"
	"weather_forecast_sub/internal/domain"
)

type WeatherClient interface {
	GetAPICurrentWeather(ctx context.Context, city string) (*domain.WeatherResponse, error)
	GetAPIDayWeather(ctx context.Context, city string) (*domain.DayWeatherResponse, error)
}

type ChainWeatherProvider interface {
	WeatherClient
	setNext(ChainWeatherProvider)
}

type ChainWeatherClient struct {
	primaryClient WeatherClient
}

func NewChainWeatherClient(clients []ChainWeatherProvider) *ChainWeatherClient {
	for i := 0; i < len(clients)-1; i++ {
		clients[i].setNext(clients[i+1])
	}

	return &ChainWeatherClient{
		primaryClient: clients[0],
	}
}

func (c *ChainWeatherClient) GetAPICurrentWeather(
	ctx context.Context, city string,
) (*domain.WeatherResponse, error) {
	return c.primaryClient.GetAPICurrentWeather(ctx, city)
}

func (c *ChainWeatherClient) GetAPIDayWeather(
	ctx context.Context, city string,
) (*domain.DayWeatherResponse, error) {
	return c.primaryClient.GetAPIDayWeather(ctx, city)
}
