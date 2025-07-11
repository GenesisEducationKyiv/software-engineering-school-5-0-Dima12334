package app

import (
	"context"
	"errors"
	"ms-notification/internal/config"
	"ms-notification/internal/handlers"
	"ms-notification/internal/server"
	"ms-notification/internal/service"
	"ms-notification/pkg/email/smtp"
	"ms-notification/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Application struct {
	config *config.Config
	server *server.Server
}

type ApplicationBuilder struct{}

func NewApplication(environment string) (*Application, error) {
	builder := ApplicationBuilder{}
	return builder.Build(environment)
}

func (ab *ApplicationBuilder) setupDependencies(app *Application) {
	emailSender := smtp.NewSMTPSender(
		app.config.SMTP.From,
		app.config.SMTP.FromName,
		app.config.SMTP.Pass,
		app.config.SMTP.Host,
		app.config.SMTP.Port,
	)

	services := service.NewServices(service.Deps{
		EmailSender: emailSender,
		EmailConfig: app.config.Email,
		HTTPConfig:  app.config.HTTP,
	})

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

	app := &Application{
		config: cfg,
	}
	ab.setupDependencies(app)

	return app, nil
}

func (a *Application) Run() {
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
}
