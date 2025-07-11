package main

import (
	"fmt"
	"log"
	"ms-weather-subscription/internal/config"
	"ms-weather-subscription/pkg/migrations"
	"os"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	cmd, err := parseArgs()
	if err != nil {
		log.Fatalf("failed to parse command: %v", err)
	}

	environment := config.GetEnvironmentOrDefault(config.DevEnvironment)

	cfg, err := config.Init(config.ConfigsDir, environment)
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}

	if err := runMigration(cmd, cfg.DB.DSN, cfg.DB.MigrationsPath); err != nil {
		log.Fatalf("migration %s failed: %v", cmd, err)
	}

	log.Printf("migration %s completed successfully", cmd)
}

func parseArgs() (string, error) {
	cmd := os.Args[len(os.Args)-1]

	switch cmd {
	case "up", "down":
		return cmd, nil
	default:
		return "", fmt.Errorf("unknown command: %s", cmd)
	}
}

func runMigration(direction, dsn, path string) error {
	switch direction {
	case "up", "down":
		return migrations.ApplyMigrations(dsn, path, direction)
	default:
		return fmt.Errorf("unknown migration direction: %s", direction)
	}
}
