package email

import (
	"bytes"
	"html/template"
	customErrors "ms-weather-subscription/pkg/errors"
	"ms-weather-subscription/pkg/logger"
	"strings"
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
		return customErrors.ErrEmailToRequired
	case strings.TrimSpace(e.Subject) == "":
		return customErrors.ErrEmailSubjectRequired
	case strings.TrimSpace(e.Body) == "":
		return customErrors.ErrEmailBodyRequired
	}

	return nil
}
