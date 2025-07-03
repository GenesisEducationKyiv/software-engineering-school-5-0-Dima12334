package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
	"weather_forecast_sub/pkg/clients"

	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/pkg/logger"
)

const (
	oneHourDuration = time.Hour
)

type CachingWeatherClient struct {
	clients.WeatherClient
	cache Cache
}

func NewCachingWeatherClient(client clients.WeatherClient, cache Cache) *CachingWeatherClient {
	return &CachingWeatherClient{
		WeatherClient: client,
		cache:         cache,
	}
}

func (s *CachingWeatherClient) GetAPICurrentWeather(
	ctx context.Context, city string,
) (*domain.WeatherResponse, error) {
	now := time.Now().UTC()
	key := fmt.Sprintf("%s:%s", strings.ToLower(city), now.Format("2006-01-02:15-00"))

	if cached, err := s.cache.Get(ctx, key); err == nil {
		var res domain.WeatherResponse
		if err := json.Unmarshal([]byte(cached), &res); err == nil {
			return &res, nil
		}
	}

	resp, err := s.WeatherClient.GetAPICurrentWeather(ctx, url.QueryEscape(city))
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(resp)
	if err != nil {
		logger.Errorf("cache marshal error (weather current): %s", err)
	}

	if err := s.cache.Set(ctx, key, string(data), oneHourDuration); err != nil {
		logger.Errorf("cache set error (weather current): %s", err)
	}

	return resp, nil
}
