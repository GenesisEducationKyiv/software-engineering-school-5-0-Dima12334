package domain

import (
	"fmt"
	"time"
)

const (
	DailyWeatherEmailFrequency  = "daily"
	HourlyWeatherEmailFrequency = "hourly"
)

type Subscription struct {
	ID        string    `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Email     string    `json:"email" db:"email"`
	City      string    `json:"city" db:"city"`
	Token     string    `json:"token" db:"token"`
	Frequency string    `json:"frequency" db:"frequency"`
	Confirmed bool      `json:"confirmed" db:"confirmed"`
}

func NewSubscription(email, city, frequency, token string) Subscription {
	return Subscription{
		CreatedAt: time.Now(),
		Email:     email,
		City:      city,
		Frequency: frequency,
		Token:     token,
		Confirmed: false,
	}
}

func (s *Subscription) CreateConfirmationLink(baseURL string) string {
	return fmt.Sprintf("%s/api/confirm/%s", baseURL, s.Token)
}

func (s *Subscription) CreateUnsubscribeLink(BaseURL string) string {
	return fmt.Sprintf("%s/api/unsubscribe/%s", BaseURL, s.Token)
}

type CreateSubscriptionInput struct {
	Email     string
	City      string
	Frequency string
}

type ConfirmationEmailInput struct {
	Email            string
	ConfirmationLink string
}
