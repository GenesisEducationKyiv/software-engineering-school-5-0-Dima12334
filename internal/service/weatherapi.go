package service

import (
	"context"
	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/pkg/cache"
	"weather_forecast_sub/pkg/clients"
)

type WeatherService struct {
	client clients.WeatherClient
	cache  cache.Cache
}

func NewWeatherService(client clients.WeatherClient, cache cache.Cache) *WeatherService {
	return &WeatherService{
		client: client,
		cache:  cache,
	}
}

func (s *WeatherService) GetCurrentWeather(ctx context.Context, city string) (
	*domain.WeatherResponse, error,
) {
	return s.client.GetAPICurrentWeather(ctx, city)
}

func (s *WeatherService) GetDayWeather(ctx context.Context, city string) (
	*domain.DayWeatherResponse, error,
) {
	return s.client.GetAPIDayWeather(ctx, city)
}
