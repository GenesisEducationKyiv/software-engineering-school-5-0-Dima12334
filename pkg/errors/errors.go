package errors

import (
	"errors"

	"github.com/lib/pq"
)

var (
	ErrSubscriptionNotFound      = errors.New("subscription doesn't exists")
	ErrSubscriptionAlreadyExists = errors.New("subscription with such email already exists")
	ErrCityNotFound              = errors.New("city doesn't exists")
	ErrWeatherAPIError           = errors.New("failed to get weather data")
)

func IsDuplicateDBError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505" // Unique violation code
	}
	return false
}
