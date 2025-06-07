package main

import (
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"log"
	"os"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/pkg/migrations"
)

func main() {
	environ := os.Getenv("ENV")
	if environ == "" {
		environ = config.DevEnvironment
	}

	cfg, err := config.Init(config.ConfigsDir, environ)
	if err != nil {
		log.Fatalf("failed to init configs: %v", err.Error())
	}

	cmd := os.Args[len(os.Args)-1]
	switch cmd {
	case "up":
		if err := migrations.ApplyMigrations(cfg.DB.DSN, cfg.DB.MigrationsPath, "up"); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Println("Migration up applied successfully.")
	case "down":
		if err := migrations.ApplyMigrations(cfg.DB.DSN, cfg.DB.MigrationsPath, "down"); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Println("Migration down applied successfully.")
	default:
		log.Printf("Unknown command: %s", cmd)
	}
}
