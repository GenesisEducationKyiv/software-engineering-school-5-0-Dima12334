package testutils

import (
	"testing"
	"weather_forecast_sub/pkg/cache"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func SetupTestCache(t *testing.T) *cache.RedisCache {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	redisConn := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
		DB:   0,
	})
	cache := cache.NewCache(redisConn)

	t.Cleanup(func() {
		if err := redisConn.Close(); err != nil {
			t.Logf("failed to close redis client: %v", err)
		}
		mr.Close()
	})

	return cache
}
