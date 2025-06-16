package testutils

import (
	"testing"
	"weather_forecast_sub/internal/config"
)

func SetupTestConfig(t *testing.T) *config.Config {
	t.Helper()

	cfg, err := config.Init(config.ConfigsDir, config.TestEnvironment)
	if err != nil {
		t.Fatalf("failed to init configs: %v", err.Error())
	}

	return cfg
}
