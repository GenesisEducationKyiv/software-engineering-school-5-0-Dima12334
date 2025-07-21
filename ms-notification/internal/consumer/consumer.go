package consumer

import (
	"common/logger"
	"ms-notification/internal/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EmailConfirmationQueue   = "email.confirmation"
	EmailDailyForecastQueue  = "email.daily_forecast"
	EmailHourlyForecastQueue = "email.hourly_forecast"
)

type Consumer struct {
	ch            *amqp.Channel
	conn          *amqp.Connection
	emailsService service.Emails
}

func NewConsumer(conn *amqp.Connection, emailsService service.Emails) (*Consumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	queues := []string{EmailConfirmationQueue, EmailDailyForecastQueue, EmailHourlyForecastQueue}
	for _, q := range queues {
		_, err := ch.QueueDeclare(q, true, false, false, false, nil)
		if err != nil {
			return nil, err
		}
	}

	return &Consumer{ch: ch, conn: conn, emailsService: emailsService}, nil
}

func (c *Consumer) Start() {
	go c.consume(EmailConfirmationQueue, c.handleConfirmationEmail)
	go c.consume(EmailDailyForecastQueue, c.handleDailyForecast)
	go c.consume(EmailHourlyForecastQueue, c.handleHourlyForecast)
}

func (c *Consumer) Stop() error {
	var err error

	if err = c.ch.Close(); err != nil {
		logger.Errorf("failed to close consumer channel: %s", err)
	}

	if err = c.conn.Close(); err != nil {
		logger.Errorf("failed to close RabbitMQ connection: %s", err)
	}

	return err
}

func (c *Consumer) consume(queue string, handler func([]byte)) {
	msgs, err := c.ch.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		logger.Errorf("failed to consume from %s: %v", queue, err)
		return
	}

	for msg := range msgs {
		handler(msg.Body)
	}
}
