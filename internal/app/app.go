package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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

const (
	devEnvironment = "dev"
)

// @title Weather Forecast API
// @version 1.0
// @description Weather API application that allows users to subscribe to weather updates for their city.
// @host weather-forecast-sub-app.onrender.com
// @BasePath /api
// @schemes http https

// @tag.name weather
// @tag.description Weather forecast operations

// @tag.name subscription
// @tag.description Subscription management operations

// Run initializes the whole application.
func Run(configDir string) {
	environ := os.Getenv("ENV")
	if environ == "" {
		environ = devEnvironment
	}

	cfg, err := config.Init(configDir, environ)
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
	emailSender := smtp.NewSMTPSender(
		cfg.SMTP.From, cfg.SMTP.FromName, cfg.SMTP.Pass, cfg.SMTP.Host, cfg.SMTP.Port,
	)

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

	cronRunner := NewCronRunner(services)
	cronRunner.Start()

	handler := handlers.NewHandler(services)

	srv := server.NewServer(cfg, handler.Init())

	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("error occurred while running http server: %s\n", err.Error())
		}
	}()
	logger.Info("server started")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second
	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		logger.Errorf("failed to stop server: %v", err.Error())
	} else {
		logger.Info("server stopped successfully")
	}
}
