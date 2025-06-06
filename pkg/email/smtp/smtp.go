package smtp

import (
	"github.com/go-gomail/gomail"
	"github.com/pkg/errors"
	"weather_forecast_sub/pkg/email"
	"weather_forecast_sub/pkg/logger"
)

type SMTPSender struct {
	from     string
	fromName string
	password string
	host     string
	port     int
}

func NewSMTPSender(from, fromName, password, host string, port int) *SMTPSender {
	return &SMTPSender{from: from, fromName: fromName, password: password, host: host, port: port}
}

func (s *SMTPSender) Send(input email.SendEmailInput) error {
	if err := input.Validate(); err != nil {
		return err
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", msg.FormatAddress(s.from, s.fromName))
	msg.SetHeader("To", input.To)
	msg.SetHeader("Subject", input.Subject)
	msg.SetBody("text/html", input.Body)

	dialer := gomail.NewDialer(s.host, s.port, s.from, s.password)
	if err := dialer.DialAndSend(msg); err != nil {
		logger.Errorf("failed to sent email: %s", err.Error())
		return errors.Wrap(err, "failed to sent email via smtp")
	}

	return nil
}
