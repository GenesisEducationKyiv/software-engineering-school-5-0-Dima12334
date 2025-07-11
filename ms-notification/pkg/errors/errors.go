package errors

import (
	"errors"
)

var (
	ErrEmailToRequired      = errors.New("email 'To' field is required")
	ErrEmailSubjectRequired = errors.New("email 'Subject' field is required")
	ErrEmailBodyRequired    = errors.New("email 'Body' field is required")
)
