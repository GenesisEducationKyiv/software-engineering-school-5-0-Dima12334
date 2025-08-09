package app

import (
	"common/logger"
	"ms-notification/internal/config"
	"ms-notification/internal/handlers"
	"ms-notification/internal/server"
	"ms-notification/internal/service"
	"ms-notification/pkg/email/smtp"
	"os"
	"os/signal"
	pb "proto_stubs"
	"syscall"
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

func (ab *ApplicationBuilder) setupDependencies(app *Application) error {
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

	grpcServer, err := server.NewServer(&app.config.HTTP)
	if err != nil {
		return err
	}

	grpcNotificationHandler := handlers.NewNotificationGRPCHandler(services.Emails)
	pb.RegisterNotificationServiceServer(grpcServer.GRPCServer(), grpcNotificationHandler)

	app.server = grpcServer

	return nil
}

func (ab *ApplicationBuilder) Build(environment string) (*Application, error) {
	cfg, err := config.Init(config.ConfigsDir, environment)
	if err != nil {
		return nil, err
	}

	if err := logger.Init(cfg.Logger.LoggerEnv, cfg.Logger.FilePath); err != nil {
		return nil, err
	}

	app := &Application{
		config: cfg,
	}
	err = ab.setupDependencies(app)

	return app, err
}

func (a *Application) Run() {
	go func() {
		if err := a.server.Run(); err != nil {
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
	a.server.Stop()
}
