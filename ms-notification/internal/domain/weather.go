package domain

type Weather struct {
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
	Description string  `json:"description"`
}

type DayWeather struct {
	SevenAM Weather `json:"seven_am"`
	TenAM   Weather `json:"ten_am"`
	OnePM   Weather `json:"one_pm"`
	FourPM  Weather `json:"four_pm"`
	SevenPM Weather `json:"seven_pm"`
	TenPM   Weather `json:"ten_pm"`
}

type WeatherType interface {
	*Weather | *DayWeather
}

type WeatherForecastEmailInput[T WeatherType] struct {
	Subscription    Subscription
	Weather         T
	Date            string
	UnsubscribeLink string
}
