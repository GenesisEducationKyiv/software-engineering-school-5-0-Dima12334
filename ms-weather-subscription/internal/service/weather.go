package service

import (
	"context"
	"ms-weather-subscription/internal/domain"
	"ms-weather-subscription/pkg/clients"
)

type WeatherService struct {
	client clients.WeatherClient
}

func NewWeatherService(client clients.WeatherClient) *WeatherService {
	return &WeatherService{client: client}
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
