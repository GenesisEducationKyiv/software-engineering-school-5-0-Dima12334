package clients

import (
	"context"
	"errors"
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
	WeatherClient
}

func NewChainWeatherClient(clients []ChainWeatherProvider) (*ChainWeatherClient, error) {
	if len(clients) == 0 {
		return nil, errors.New("cannot create ChainWeatherClient with empty client list")
	}

	for i := 0; i < len(clients)-1; i++ {
		clients[i].setNext(clients[i+1])
	}

	return &ChainWeatherClient{
		WeatherClient: clients[0],
	}, nil
}
