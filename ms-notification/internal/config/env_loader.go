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
	var allVars []string

	if environment != TestEnvironment {
		allVars = append(
			allVars,
			"LOGG_ENV",
			"SMTP_PASSWORD",
			"RABBITMQ_URL",
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
