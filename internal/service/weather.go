package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/pkg/cache"
	"weather_forecast_sub/pkg/clients"
	"weather_forecast_sub/pkg/logger"
)

const (
	oneDayDuration = 24 * time.Hour
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
	now := time.Now().UTC()
	key := fmt.Sprintf("%s:%s", strings.ToLower(city), now.Format("2006-01-02:15-00"))

	cached, err := s.cache.Get(ctx, key)
	if err == nil {
		var res domain.WeatherResponse
		if err := json.Unmarshal([]byte(cached), &res); err == nil {
			return &res, nil
		}
	}

	resp, err := s.client.GetAPICurrentWeather(ctx, url.QueryEscape(city))
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("failed to marshal current weather response: %s", err)
		return nil, err
	}

	if err = s.cache.Set(ctx, key, string(data), time.Hour); err != nil {
		logger.Errorf("failed to cache current weather response: %s", err)
		return nil, err
	}

	return resp, nil
}

func (s *WeatherService) GetDayWeather(ctx context.Context, city string) (
	*domain.DayWeatherResponse, error,
) {
	now := time.Now().UTC()
	key := fmt.Sprintf("%s:%s", strings.ToLower(city), now.Format(time.DateOnly))

	cached, err := s.cache.Get(ctx, key)
	if err == nil {
		var res domain.DayWeatherResponse
		if err := json.Unmarshal([]byte(cached), &res); err == nil {
			return &res, nil
		}
	}

	resp, err := s.client.GetAPIDayWeather(ctx, url.QueryEscape(city))
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("failed to marshal day weather response: %s", err)
		return nil, err
	}

	if err = s.cache.Set(ctx, key, string(data), oneDayDuration); err != nil {
		logger.Errorf("failed to cache day weather response: %s", err)
		return nil, err
	}

	return resp, nil
}
