package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	WeatherCacheHitCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "weather_cache_hit_count",
		Help: "Total cache hits for weather data",
	})
)
