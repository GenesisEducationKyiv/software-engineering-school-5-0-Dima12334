package main

import (
	"log"
	"os"
	"weather_forecast_sub/internal/app"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/db"
	"weather_forecast_sub/internal/handlers"
	"weather_forecast_sub/internal/repository"
	"weather_forecast_sub/internal/server"
	"weather_forecast_sub/internal/service"
	"weather_forecast_sub/pkg/clients"
	"weather_forecast_sub/pkg/email/smtp"
	"weather_forecast_sub/pkg/hash"
	"weather_forecast_sub/pkg/logger"
)

func main() {
	environ := os.Getenv("ENV")
	if environ == "" {
		environ = config.DevEnvironment
	}

	cfg, err := config.Init(config.ConfigsDir, environ)
	if err != nil {
		log.Fatalf("failed to init configs: %v", err.Error())
		return
	}

	if err := logger.Init(cfg.Logger); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	dbConn, err := db.ConnectDB(cfg.DB)
	if err != nil {
		logger.Errorf("failed to connect to database: %v", err.Error())
		return
	}
	defer func() {
		if err := dbConn.Close(); err != nil {
			logger.Errorf("error occurred on db connection close: %s", err.Error())
		} else {
			logger.Info("db connection closed successfully")
		}
	}()

	hasher := hash.NewSHA256Hasher()
	emailSender := smtp.NewSMTPSender(cfg.SMTP.From, cfg.SMTP.FromName, cfg.SMTP.Pass, cfg.SMTP.Host, cfg.SMTP.Port)

	thirdPartyClients := clients.NewClients(cfg.ThirdParty)
	repositories := repository.NewRepositories(dbConn)
	services := service.NewServices(
		service.Deps{
			Clients:     thirdPartyClients,
			Repos:       repositories,
			EmailHasher: hasher,
			EmailSender: emailSender,
			EmailConfig: cfg.Email,
			HTTPConfig:  cfg.HTTP,
		},
	)

	cronRunner := app.NewCronRunner(services)
	cronRunner.Start()

	handler := handlers.NewHandler(services)

	srv := server.NewServer(cfg, handler.Init(cfg))

	app.Run(srv)
}
