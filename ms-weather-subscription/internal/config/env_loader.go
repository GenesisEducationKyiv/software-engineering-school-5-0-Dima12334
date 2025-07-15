package config

import (
	"log"
	"os"

	commonCfg "common/config"

	"github.com/joho/godotenv"
)

type GodotenvLoader struct{}

func (e *GodotenvLoader) LoadEnvFile(fileName string) error {
	if fileName == "" {
		return nil
	}

	envPath := commonCfg.GetOriginalPath(fileName)

	err := godotenv.Load(envPath)
	if err != nil {
		log.Fatalf("error loading %s file: %v", fileName, err)
	}
	return nil
}

func (e *GodotenvLoader) GetRequiredEnvVars(environment string) map[string]string {
	baseVars := []string{
		"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
		"REDIS_ADDRESS", "REDIS_CACHE_DB", "REDIS_PASSWORD",
	}

	var allVars []string
	allVars = append(allVars, baseVars...)

	if environment != TestEnvironment {
		allVars = append(
			allVars,
			"LOGG_ENV",
			"HTTP_HOST",
			"WEATHER_API_KEY",
			"VISUAL_CROSSING_API_KEY",
			"NOTIFICATION_SERVICE_URL",
		)
	}

	result := make(map[string]string)
	for _, key := range allVars {
		val, exists := os.LookupEnv(key)
		if !exists || val == "" {
			log.Fatalf("environment variable %q is required but not set or empty", key)
		}
		result[key] = val
	}

	return result
}

func GetEnvironmentOrDefault(defaultEnvironment string) string {
	environment := os.Getenv("ENV")
	if environment == "" {
		return defaultEnvironment
	}
	return environment
}
