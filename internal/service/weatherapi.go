package service

import (
	"weather_forecast_sub/pkg/clients"
)

type WeatherService struct {
	client clients.WeatherClient
}

func NewWeatherService(client clients.WeatherClient) *WeatherService {
	return &WeatherService{client: client}
}

func (s *WeatherService) GetCurrentWeather(city string) (*clients.WeatherResponse, error) {
	return s.client.GetAPICurrentWeather(city)
}

func (s *WeatherService) GetDayWeather(city string) (*clients.DayWeatherResponse, error) {
	return s.client.GetAPIDayWeather(city)
}
