package domain

type WeatherResponse struct {
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
	Description string  `json:"description"`
}

type DayWeatherResponse struct {
	SevenAM WeatherResponse `json:"seven_am"`
	TenAM   WeatherResponse `json:"ten_am"`
	OnePM   WeatherResponse `json:"one_pm"`
	FourPM  WeatherResponse `json:"four_pm"`
	SevenPM WeatherResponse `json:"seven_pm"`
	TenPM   WeatherResponse `json:"ten_pm"`
}
