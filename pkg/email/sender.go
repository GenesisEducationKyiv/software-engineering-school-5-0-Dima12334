package email

import (
	"bytes"
	"errors"
	"html/template"
	"weather_forecast_sub/pkg/logger"
)

type SendEmailInput struct {
	To      string
	Subject string
	Body    string
}

//go:generate mockgen -source=sender.go -destination=mocks/mock_sender.go

type Sender interface {
	Send(input SendEmailInput) error
}

func (e *SendEmailInput) GenerateBodyFromHTML(templateFileName string, data any) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		logger.Errorf("failed to parse file %s:%s", templateFileName, err.Error())
		return err
	}

	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}

	e.Body = buf.String()

	return nil
}

func (e *SendEmailInput) Validate() error {
	if e.To == "" {
		return errors.New("empty email to")
	}

	if e.Subject == "" || e.Body == "" {
		return errors.New("empty email subject/body")
	}

	return nil
}
