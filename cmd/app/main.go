package main

import (
	"log"
	"weather_forecast_sub/internal/app"
	"weather_forecast_sub/internal/config"
)

func main() {
	environment := config.GetEnvironmentOrDefault(config.DevEnvironment)

	application, err := app.NewApplication(environment)
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}

	application.Run()
}
