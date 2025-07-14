package clients

import "ms-weather-subscription/internal/domain"

type WeatherResponseType interface {
	*domain.WeatherResponse | *domain.DayWeatherResponse
}

type ConfirmationEmailInput struct {
	Email            string
	ConfirmationLink string
}

type WeatherForecastEmailInput[T WeatherResponseType] struct {
	Subscription    domain.Subscription
	Weather         T
	Date            string
	UnsubscribeLink string
}

type SubscriptionNotificationSender interface {
	SendConfirmationEmail(ConfirmationEmailInput) error
}

type WeatherNotificationSender interface {
	SendWeatherForecastDailyEmail(WeatherForecastEmailInput[*domain.DayWeatherResponse]) error
	SendWeatherForecastHourlyEmail(WeatherForecastEmailInput[*domain.WeatherResponse]) error
}

type NotificationSender interface {
	SubscriptionNotificationSender
	WeatherNotificationSender
}

type NotificationClient struct {
	notificationServiceURL string
}

func NewNotificationClient(NotificationServiceURL string) *NotificationClient {
	return &NotificationClient{
		notificationServiceURL: NotificationServiceURL,
	}
}

func (n *NotificationClient) SendConfirmationEmail(inp ConfirmationEmailInput) error {
	// TODO: Implementation for sending confirmation email
	return nil
}

func (n *NotificationClient) SendWeatherForecastDailyEmail(
	inp WeatherForecastEmailInput[*domain.DayWeatherResponse],
) error {
	// TODO: Implementation for sending daily weather forecast email
	return nil
}

func (n *NotificationClient) SendWeatherForecastHourlyEmail(
	inp WeatherForecastEmailInput[*domain.WeatherResponse],
) error {
	// TODO: Implementation for sending hourly weather forecast email
	return nil
}
