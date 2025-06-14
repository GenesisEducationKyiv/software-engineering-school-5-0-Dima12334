package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type GodotenvLoader struct{}

func NewGodotenvLoader() *GodotenvLoader {
	return &GodotenvLoader{}
}

func (e *GodotenvLoader) LoadEnvFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	err := godotenv.Load(filePath)
	if err != nil {
		log.Fatalf("error loading %s file: %v", filePath, err)
	}
	return nil
}

func (e *GodotenvLoader) GetRequiredEnvVars(environment string) map[string]string {
	baseVars := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE"}

	var allVars []string
	allVars = append(allVars, baseVars...)

	if environment != TestEnvironment {
		allVars = append(allVars, "LOGG_ENV", "HTTP_HOST", "WEATHER_API_KEY", "SMTP_PASSWORD")
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
