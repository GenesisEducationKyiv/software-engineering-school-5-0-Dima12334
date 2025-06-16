package email

import (
	"bytes"
	"errors"
	"html/template"
	"strings"
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
	switch {
	case strings.TrimSpace(e.To) == "":
		return errors.New("email 'To' field is required")
	case strings.TrimSpace(e.Subject) == "":
		return errors.New("email 'Subject' field is required")
	case strings.TrimSpace(e.Body) == "":
		return errors.New("email 'Body' field is required")
	}

	return nil
}
