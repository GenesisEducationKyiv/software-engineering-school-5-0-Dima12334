package main

import (
	"fmt"
	"log"
	"os"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/pkg/migrations"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	cmd := parseArgs()

	environ := config.GetEnvironmentOrDefault(config.DevEnvironment)

	cfg, err := config.Init(config.ConfigsDir, environ)
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}

	if err := runMigration(cmd, cfg.DB.DSN, cfg.DB.MigrationsPath); err != nil {
		log.Fatalf("migration %s failed: %v", cmd, err)
	}

	log.Printf("migration %s completed successfully", cmd)
}

func parseArgs() string {
	cmd := os.Args[len(os.Args)-1]

	switch cmd {
	case "up", "down":
		return cmd
	default:
		log.Printf("unknown command: %s", cmd)
		return ""
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
