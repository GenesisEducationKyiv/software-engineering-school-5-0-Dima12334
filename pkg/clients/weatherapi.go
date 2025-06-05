package clients

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	customErrors "weather_forecast_sub/pkg/errors"
	"weather_forecast_sub/pkg/logger"
)

const (
	weatherAPICityNotFoundCode = 1006
	weatherAPIClientTimeout    = 10 * time.Second
)

type WeatherAPIClient struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewWeatherAPIClient(apiKey string) *WeatherAPIClient {
	return &WeatherAPIClient{
		APIKey:     apiKey,
		BaseURL:    "https://api.weatherapi.com/v1",
		HTTPClient: &http.Client{Timeout: weatherAPIClientTimeout},
	}
}

type weatherAPIErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type WeatherResponse struct {
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
	Description string  `json:"description"`
}

type currentWeatherAPIResponse struct {
	Current struct {
		TempC     float32 `json:"temp_c"`
		Humidity  float32 `json:"humidity"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

type DayWeatherResponse struct {
	SevenAM WeatherResponse `json:"seven_am"`
	TenAM   WeatherResponse `json:"ten_am"`
	OnePM   WeatherResponse `json:"one_pm"`
	FourPM  WeatherResponse `json:"four_pm"`
	SevenPM WeatherResponse `json:"seven_pm"`
	TenPM   WeatherResponse `json:"ten_pm"`
}

type dayWeatherAPIResponse struct {
	Forecast struct {
		ForecastDay []struct {
			Hour []struct {
				Time      string  `json:"time"` // "2025-05-17 07:00"
				TempC     float32 `json:"temp_c"`
				Humidity  float32 `json:"humidity"`
				Condition struct {
					Text string `json:"text"`
				} `json:"condition"`
			} `json:"hour"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

func (c *WeatherAPIClient) GetAPICurrentWeather(city string) (*WeatherResponse, error) {
	url := fmt.Sprintf("%s/current.json?key=%s&q=%s", c.BaseURL, c.APIKey, city)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		logger.Errorf("error making request to WeatherClient API: %s", err.Error())
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if closeErr := resp.Body.Close(); closeErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to close response body: %w", err, closeErr)
			} else {
				err = fmt.Errorf("failed to close response body: %w", closeErr)
			}
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var apiErr weatherAPIErrorResponse
		if err = json.NewDecoder(resp.Body).Decode(&apiErr); err == nil {
			if resp.StatusCode == http.StatusBadRequest && apiErr.Error.Code == weatherAPICityNotFoundCode {
				return nil, customErrors.ErrCityNotFound
			}
			logger.Errorf(
				"WeatherAPI error. Status code: %d, api code: %d, message: %s",
				resp.StatusCode,
				apiErr.Error.Code,
				apiErr.Error.Message,
			)
			return nil, customErrors.ErrWeatherAPIError
		}
		logger.Errorf("WeatherAPI error. Status code: %d", resp.StatusCode)
		return nil, customErrors.ErrWeatherAPIError
	}

	var result currentWeatherAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		logger.Errorf("error parsing WeatherAPI response: %s", err.Error())
		return nil, customErrors.ErrWeatherAPIError
	}

	return &WeatherResponse{
		Temperature: result.Current.TempC,
		Humidity:    result.Current.Humidity,
		Description: result.Current.Condition.Text,
	}, nil
}

func (c *WeatherAPIClient) GetAPIDayWeather(city string) (*DayWeatherResponse, error) {
	url := fmt.Sprintf("%s/forecast.json?key=%s&q=%s&days=1", c.BaseURL, c.APIKey, city)

	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		logger.Errorf("error making request to WeatherAPI: %s", err.Error())
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		if closeErr := resp.Body.Close(); closeErr != nil {
			if err != nil {
				err = fmt.Errorf("%w; failed to close response body: %w", err, closeErr)
			} else {
				err = fmt.Errorf("failed to close response body: %w", closeErr)
			}
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		var apiErr weatherAPIErrorResponse
		if err = json.NewDecoder(resp.Body).Decode(&apiErr); err == nil {
			if resp.StatusCode == http.StatusBadRequest && apiErr.Error.Code == weatherAPICityNotFoundCode {
				return nil, customErrors.ErrCityNotFound
			}
			logger.Errorf("WeatherAPI error. Status code: %d, api code: %d, message: %s", resp.StatusCode, apiErr.Error.Code, apiErr.Error.Message)
			return nil, customErrors.ErrWeatherAPIError
		}
		logger.Errorf("WeatherAPI error. Status code: %d", resp.StatusCode)
		return nil, customErrors.ErrWeatherAPIError
	}

	var result dayWeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf("error parsing WeatherAPI response: %s", err.Error())
		return nil, customErrors.ErrWeatherAPIError
	}

	// Map of required times
	targetHours := map[string]*WeatherResponse{
		"07:00": {},
		"10:00": {},
		"13:00": {},
		"16:00": {},
		"19:00": {},
		"22:00": {},
	}

	for _, hourData := range result.Forecast.ForecastDay[0].Hour {
		// time format: "2025-05-17 07:00"
		parts := strings.Split(hourData.Time, " ")
		if len(parts) != 2 {
			continue
		}
		timePart := parts[1]

		if target, ok := targetHours[timePart]; ok {
			target.Temperature = hourData.TempC
			target.Humidity = hourData.Humidity
			target.Description = hourData.Condition.Text
		}
	}

	return &DayWeatherResponse{
		SevenAM: *targetHours["07:00"],
		TenAM:   *targetHours["10:00"],
		OnePM:   *targetHours["13:00"],
		FourPM:  *targetHours["16:00"],
		SevenPM: *targetHours["19:00"],
		TenPM:   *targetHours["22:00"],
	}, nil
}
