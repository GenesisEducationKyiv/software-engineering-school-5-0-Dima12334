package app

import (
	"common/logger"
	amqp "github.com/rabbitmq/amqp091-go"
	"ms-notification/internal/config"
	"ms-notification/internal/consumer"
	"ms-notification/internal/service"
	"ms-notification/pkg/email/smtp"
	"os"
	"os/signal"
	"syscall"
)

type Application struct {
	config   *config.Config
	consumer *consumer.Consumer
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
	})

	rabbitConn, err := amqp.Dial(app.config.RabbitMQ.URL)
	if err != nil {
		return err
	}

	consumer, err := consumer.NewConsumer(rabbitConn, services.Emails)
	if err != nil {
		return err
	}
	consumer.Start()
	app.consumer = consumer

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
	logger.Info("notification consumer started")
	a.waitForShutdown()
}

func (a *Application) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info("shutting down notification service...")
	a.shutdown()
}

func (a *Application) shutdown() {
	if err := a.consumer.Stop(); err != nil {
		logger.Errorf("failed to stop consumer: %s", err)
	} else {
		logger.Info("consumer stopped successfully")
	}
}
