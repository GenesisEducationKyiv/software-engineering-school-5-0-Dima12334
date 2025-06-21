package errors

import (
	"errors"

	"github.com/lib/pq"
)

const (
	PgUniqueViolationCode = "23505"
)

var (
	ErrSubscriptionNotFound      = errors.New("subscription doesn't exists")
	ErrSubscriptionAlreadyExists = errors.New("subscription with such email already exists")

	ErrCityNotFound    = errors.New("city doesn't exists")
	ErrWeatherAPIError = errors.New("failed to get weather data")

	ErrEmailToRequired      = errors.New("email 'To' field is required")
	ErrEmailSubjectRequired = errors.New("email 'Subject' field is required")
	ErrEmailBodyRequired    = errors.New("email 'Body' field is required")
)

func IsDuplicateDBError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == PgUniqueViolationCode
	}
	return false
}
