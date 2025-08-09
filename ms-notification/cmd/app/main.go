package main

import (
	"log"
	"ms-notification/internal/app"
	"ms-notification/internal/config"
)

func main() {
	environment := config.GetEnvironmentOrDefault(config.DevEnvironment)

	application, err := app.NewApplication(environment)
	if err != nil {
		log.Fatalf("failed to create application: %v", err)
	}

	application.Run()
}
