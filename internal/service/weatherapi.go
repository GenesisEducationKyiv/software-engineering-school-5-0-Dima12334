package service

import (
	"context"
	"weather_forecast_sub/internal/domain"
)

type WeatherClient interface {
	GetAPICurrentWeather(ctx context.Context, city string) (*domain.WeatherResponse, error)
	GetAPIDayWeather(ctx context.Context, city string) (*domain.DayWeatherResponse, error)
}

type WeatherService struct {
	client WeatherClient
}

func NewWeatherService(client WeatherClient) *WeatherService {
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
