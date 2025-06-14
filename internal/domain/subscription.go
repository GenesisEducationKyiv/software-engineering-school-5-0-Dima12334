package domain

import "time"

const (
	DailyWeatherEmailFrequency  = "daily"
	HourlyWeatherEmailFrequency = "hourly"
)

type Subscription struct {
	ID         string     `json:"id" db:"id"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	Email      string     `json:"email" db:"email"`
	City       string     `json:"city" db:"city"`
	Token      string     `json:"token" db:"token"`
	Frequency  string     `json:"frequency" db:"frequency"`
	Confirmed  bool       `json:"confirmed" db:"confirmed"`
	LastSentAt *time.Time `json:"last_sent_at" db:"last_sent_at"`
}
