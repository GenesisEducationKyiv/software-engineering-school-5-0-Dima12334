package consumer

import (
	"common/logger"
	"encoding/json"
	"fmt"
	"ms-notification/internal/domain"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

type confirmationEmailCommand struct {
	Email            string `json:"email"`
	ConfirmationLink string `json:"confirmation_link"`
}

type baseForecastCommand struct {
	Subscription    domain.Subscription `json:"subscription"`
	Date            string              `json:"date"`
	UnsubscribeLink string              `json:"unsubscribe_link"`
}

type dailyForecastCommand struct {
	baseForecastCommand
	Weather domain.DayWeather `json:"weather"`
}

type hourlyForecastCommand struct {
	baseForecastCommand
	Weather domain.Weather `json:"weather"`
}

type MessageHandlerFunc func(msg amqp.Delivery) error

func (c *Consumer) wrapHandler(handler MessageHandlerFunc) func(amqp.Delivery) {
	return func(msg amqp.Delivery) {
		err := handler(msg)
		if err != nil {
			logger.Errorf("handler error: %s", err)

			var nackErr error
			if strings.Contains(err.Error(), "invalid") && strings.Contains(err.Error(), "payload") {
				nackErr = msg.Nack(false, false)
			} else {
				nackErr = msg.Nack(false, true)
			}

			if nackErr != nil {
				logger.Errorf("failed to Nack message: %s", nackErr)
			}
			return
		}

		if ackErr := msg.Ack(false); ackErr != nil {
			logger.Errorf("failed to Ack message: %s", ackErr)
		}
	}
}

func (c *Consumer) handleConfirmationEmail(msg amqp.Delivery) error {
	var cmd confirmationEmailCommand
	if err := json.Unmarshal(msg.Body, &cmd); err != nil {
		return fmt.Errorf("invalid confirmation email payload: %w", err)
	}

	inp := domain.ConfirmationEmailInput(cmd)
	if err := c.emailService.SendConfirmationEmail(inp); err != nil {
		return fmt.Errorf("confirmation email send error: %w", err)
	}

	return nil
}

func (c *Consumer) handleDailyForecast(msg amqp.Delivery) error {
	var cmd dailyForecastCommand
	if err := json.Unmarshal(msg.Body, &cmd); err != nil {
		return fmt.Errorf("invalid daily forecast email payload: %w", err)
	}

	inp := domain.WeatherForecastEmailInput[*domain.DayWeather]{
		Subscription:    cmd.Subscription,
		Weather:         &cmd.Weather,
		Date:            cmd.Date,
		UnsubscribeLink: cmd.UnsubscribeLink,
	}
	if err := c.emailService.SendWeatherForecastDailyEmail(inp); err != nil {
		return fmt.Errorf("daily forecast email send error: %w", err)
	}

	return nil
}

func (c *Consumer) handleHourlyForecast(msg amqp.Delivery) error {
	var cmd hourlyForecastCommand
	if err := json.Unmarshal(msg.Body, &cmd); err != nil {
		return fmt.Errorf("invalid hourly forecast email payload: %w", err)
	}

	inp := domain.WeatherForecastEmailInput[*domain.Weather]{
		Subscription:    cmd.Subscription,
		Weather:         &cmd.Weather,
		Date:            cmd.Date,
		UnsubscribeLink: cmd.UnsubscribeLink,
	}
	if err := c.emailService.SendWeatherForecastHourlyEmail(inp); err != nil {
		return fmt.Errorf("hourly forecast email send error: %w", err)
	}

	return nil
}
