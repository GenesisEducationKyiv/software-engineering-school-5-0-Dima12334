package main

import (
	"log"
	"ms-weather-subscription/internal/app"
	"ms-weather-subscription/internal/config"
)

func main() {
	environment := config.GetEnvironmentOrDefault(config.DevEnvironment)

	application, err := app.NewApplication(environment)
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}

	application.Run()
}
