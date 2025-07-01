package testutils

import (
	"context"
	"testing"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/pkg/cache"
	"weather_forecast_sub/pkg/clients"
)

func SetupTestCache(t *testing.T) *cache.RedisCache {
	t.Helper()

	cfg, err := config.Init(config.ConfigsDir, config.TestEnvironment)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	redisConn := clients.NewRedisConnection(cfg.Redis)
	redisCache := cache.NewCache(redisConn)

	t.Cleanup(func() {
		if err := redisConn.FlushDB(context.Background()).Err(); err != nil {
			t.Logf("failed to flush redis db: %v", err)
		}
		if err := redisConn.Close(); err != nil {
			t.Logf("failed to close redis: %v", err)
		}
	})

	return redisCache
}
