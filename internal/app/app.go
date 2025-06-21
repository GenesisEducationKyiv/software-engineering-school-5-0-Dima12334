package app

import (
	"context"
	"errors"
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

	"github.com/jmoiron/sqlx"
)

type Application struct {
	config     *config.Config
	server     *server.Server
	cronRunner Cron
	dbConn     *sqlx.DB
}

type ApplicationBuilder struct{}

func NewApplication(environment string) (*Application, error) {
	builder := ApplicationBuilder{}
	return builder.Build(environment)
}

func (ab *ApplicationBuilder) setupDependencies(app *Application) {
	hasher := &hash.SHA256Hasher{}
	emailSender := smtp.NewSMTPSender(
		app.config.SMTP.From,
		app.config.SMTP.FromName,
		app.config.SMTP.Pass,
		app.config.SMTP.Host,
		app.config.SMTP.Port,
	)
	thirdPartyClients := clients.NewClients(app.config.ThirdParty)
	repositories := repository.NewRepositories(app.dbConn)

	services := service.NewServices(service.Deps{
		Clients:            thirdPartyClients,
		Repos:              repositories,
		SubscriptionHasher: hasher,
		EmailSender:        emailSender,
		EmailConfig:        app.config.Email,
		HTTPConfig:         app.config.HTTP,
	})

	app.cronRunner = NewCronRunner(services.WeatherForecastSender)

	handler := handlers.NewHandler(services)

	app.server = server.NewServer(&app.config.HTTP, handler.Init(app.config.Environment))
}

func (ab *ApplicationBuilder) Build(environment string) (*Application, error) {
	cfg, err := config.Init(config.ConfigsDir, environment)
	if err != nil {
		return nil, err
	}

	if err := logger.Init(cfg.Logger); err != nil {
		return nil, err
	}

	dbConn, err := db.NewDBConnection(cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	app := &Application{
		config: cfg,
		dbConn: dbConn,
	}
	ab.setupDependencies(app)

	return app, nil
}

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
func (a *Application) Run() {
	a.cronRunner.Start()
	defer a.cronRunner.Stop()

	go func() {
		if err := a.server.Run(); !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("error occurred while running http server: %s\n", err.Error())
		}
	}()
	logger.Info("server started")

	a.waitForShutdown()
}

func (a *Application) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info("shutting down server...")
	a.shutdown()
}

func (a *Application) shutdown() {
	const shutdownTimeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := a.server.Stop(ctx); err != nil {
		logger.Errorf("failed to stop server: %v", err.Error())
	} else {
		logger.Info("server stopped successfully")
	}

	if err := a.dbConn.Close(); err != nil {
		logger.Errorf("error occurred on db connection close: %s", err.Error())
	} else {
		logger.Info("db connection closed successfully")
	}
}
