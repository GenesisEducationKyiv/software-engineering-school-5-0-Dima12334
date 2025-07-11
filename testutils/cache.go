package testutils

import (
	"context"
	"ms-weather-subscription/internal/config"
	"ms-weather-subscription/pkg/cache"
	"ms-weather-subscription/pkg/clients"
	"testing"
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
			t.Fatalf("failed to flush redis db: %v", err)
		}
		if err := redisConn.Close(); err != nil {
			t.Fatalf("failed to close redis: %v", err)
		}
	})

	return redisCache
}
