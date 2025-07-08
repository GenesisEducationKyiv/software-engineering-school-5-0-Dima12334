package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/pkg/clients"
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

		logger.Warnf("cache unmarshal error (weather current): %v", err)
	} else {
		HandleRedisError(err)
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
		HandleRedisError(err)
	}

	return resp, nil
}

func HandleRedisError(err error) {
	if err == nil {
		return
	}

	if errors.Is(err, redis.Nil) {
		return
	}

	if errors.Is(err, redis.ErrClosed) || strings.Contains(err.Error(), "connection refused") {
		logger.Errorf("redis not available: %v", err)
	} else {
		logger.Errorf("redis get error: %v", err)
	}
}
