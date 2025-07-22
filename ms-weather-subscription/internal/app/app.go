package app

import (
	"common/logger"
	"context"
	"errors"
	"log"
	"ms-weather-subscription/internal/config"
	"ms-weather-subscription/internal/db"
	"ms-weather-subscription/internal/handlers"
	"ms-weather-subscription/internal/repository"
	"ms-weather-subscription/internal/server"
	"ms-weather-subscription/internal/service"
	"ms-weather-subscription/pkg/cache"
	"ms-weather-subscription/pkg/clients"
	"ms-weather-subscription/pkg/hash"
	"ms-weather-subscription/pkg/publisher"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/redis/go-redis/v9"

	"github.com/jmoiron/sqlx"
)

type Application struct {
	config         *config.Config
	server         *server.Server
	cronRunner     Cron
	dbConn         *sqlx.DB
	redisConn      *redis.Client
	emailPublisher *publisher.EmailPub
}

type ApplicationBuilder struct{}

func NewApplication(environment string) (*Application, error) {
	builder := ApplicationBuilder{}
	return builder.Build(environment)
}

func (ab *ApplicationBuilder) setupDependencies(app *Application) {
	hasher := &hash.SHA256Hasher{}

	redisCache := cache.NewCache(app.redisConn)

	primaryWeatherClient := clients.NewWeatherAPIClient(app.config.ThirdParty.WeatherAPIKey)
	fallbackWeatherClients := []clients.ChainWeatherProvider{
		clients.NewVisualCrossingClient(app.config.ThirdParty.VisualCrossingAPIKey),
	}
	allWeatherClients := append(
		[]clients.ChainWeatherProvider{primaryWeatherClient}, fallbackWeatherClients...,
	)
	chainWeatherClient, err := clients.NewChainWeatherClient(allWeatherClients)
	if err != nil {
		log.Fatalf("failed to create chain weather client: %v", err)
	}
	cachingWeatherClient := clients.NewCachingWeatherClient(chainWeatherClient, redisCache)

	repositories := repository.NewRepositories(app.dbConn)

	services := service.NewServices(service.Deps{
		WeatherClient:      cachingWeatherClient,
		Repos:              repositories,
		SubscriptionHasher: hasher,
		HTTPConfig:         app.config.HTTP,
		EmailPublisher:     app.emailPublisher,
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

	if err := logger.Init(cfg.Logger.LoggerEnv, cfg.Logger.FilePath); err != nil {
		return nil, err
	}

	dbConn, err := db.NewDBConnection(cfg.DB.DSN)
	if err != nil {
		return nil, err
	}
	err = db.ValidateDBConnection(dbConn)
	if err != nil {
		log.Fatal(err)
	}

	redisConn := clients.NewRedisConnection(cfg.Redis)
	err = clients.ValidateRedisConnection(redisConn)
	if err != nil {
		log.Fatal(err)
	}

	var rabbitConn *amqp.Connection
	var emailPublisher *publisher.EmailPub
	if environment != config.TestEnvironment {
		rabbitConn, err = amqp.Dial(cfg.RabbitMQ.URL)
		if err != nil {
			log.Fatalf("failed to create RabbitMQ connection: %v", err)
		}
		emailPublisher, err = publisher.NewEmailPublisher(rabbitConn)
		if err != nil {
			log.Fatalf("failed to create email publisher: %v", err)
		}
	}

	app := &Application{
		config:         cfg,
		dbConn:         dbConn,
		redisConn:      redisConn,
		emailPublisher: emailPublisher,
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

	if err := a.redisConn.Close(); err != nil {
		logger.Errorf("error occurred on redis connection close: %s", err.Error())
	} else {
		logger.Info("redis connection closed successfully")
	}

	if a.emailPublisher != nil {
		if err := a.emailPublisher.Stop(); err != nil {
			logger.Errorf("failed to stop email publisher: %s", err)
		} else {
			logger.Info("email publisher stopped successfully")
		}
	}
}
