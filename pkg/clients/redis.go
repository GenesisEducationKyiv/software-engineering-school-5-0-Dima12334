package clients

import (
	"context"
	"strings"
	"time"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/pkg/logger"

	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

const (
	pingTimeout = 5 * time.Second
)

func NewRedisConnection(redisCfg config.RedisConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     redisCfg.Address,
		DB:       redisCfg.CacheDB,
		Password: redisCfg.Password,
	})
}

func ValidateRedisConnection(redisClient *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()

	return errors.Wrap(redisClient.Ping(ctx).Err(), "ping redis wasn't successful")
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
