package main

import (
	"log"
	"os"
	"weather_forecast_sub/internal/app"
	"weather_forecast_sub/internal/config"
)

func main() {
	environ := getEnvironment()

	application, err := app.NewApplication(environ)
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}

	application.Run()
}

func getEnvironment() string {
	environ := os.Getenv("ENV")
	if environ == "" {
		return config.DevEnvironment
	}
	return environ
}
