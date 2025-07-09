package clients

import (
	"context"
	"time"
	"weather_forecast_sub/internal/config"

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
