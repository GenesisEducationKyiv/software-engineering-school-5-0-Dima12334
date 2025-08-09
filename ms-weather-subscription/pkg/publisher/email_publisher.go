package publisher

import (
	"common/logger"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EmailConfirmationQueue   = "email.confirmation"
	EmailDailyForecastQueue  = "email.daily_forecast"
	EmailHourlyForecastQueue = "email.hourly_forecast"
)

//go:generate mockgen -source=email_publisher.go -destination=mocks/mock_email_publisher.go

type EmailPublisher interface {
	Publish(queue string, msg any) error
}

type EmailPub struct {
	ch   *amqp.Channel
	conn *amqp.Connection
}

func NewEmailPublisher(conn *amqp.Connection) (*EmailPub, error) {
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

	return &EmailPub{ch: ch, conn: conn}, nil
}

func (p *EmailPub) Stop() error {
	var err error

	if err = p.ch.Close(); err != nil {
		logger.Errorf("failed to close publisher channel: %s", err)
	}

	if err = p.conn.Close(); err != nil {
		logger.Errorf("failed to close RabbitMQ connection: %s", err)
	}

	return err
}

func (p *EmailPub) Publish(queue string, msg any) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.ch.Publish("", queue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
}
