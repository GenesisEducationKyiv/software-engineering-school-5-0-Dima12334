package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
	"weather_forecast_sub/internal/domain"
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

type currentWeatherAPIResponse struct {
	Current struct {
		TempC     float32 `json:"temp_c"`
		Humidity  float32 `json:"humidity"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
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

func (c *WeatherAPIClient) processErrorResponse(resp *http.Response) error {
	var apiErr weatherAPIErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
		logger.Errorf("error decoding WeatherAPI error response: %s", err)
		return customErrors.ErrWeatherDataError
	}

	if resp.StatusCode == http.StatusBadRequest && apiErr.Error.Code == weatherAPICityNotFoundCode {
		return customErrors.ErrCityNotFound
	}

	logger.Errorf(
		"WeatherAPI error. Status code: %d, api code: %d, message: %s",
		resp.StatusCode,
		apiErr.Error.Code,
		apiErr.Error.Message,
	)
	return customErrors.ErrWeatherDataError
}

func (c *WeatherAPIClient) GetAPICurrentWeather(
	ctx context.Context, city string,
) (*domain.WeatherResponse, error) {
	url := fmt.Sprintf("%s/current.json?key=%s&q=%s", c.BaseURL, c.APIKey, city)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Errorf("failed to create WeatherAPI request: %s", err)
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		logger.Errorf("error making request to WeatherClient API: %s", err.Error())
		return nil, err
	}
	defer closeBody(resp.Body, &err)

	if resp.StatusCode != http.StatusOK {
		err = c.processErrorResponse(resp)
		return nil, err
	}

	var result currentWeatherAPIResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		logger.Errorf("error parsing WeatherAPI response: %s", err.Error())
		return nil, customErrors.ErrWeatherDataError
	}

	return &domain.WeatherResponse{
		Temperature: result.Current.TempC,
		Humidity:    result.Current.Humidity,
		Description: result.Current.Condition.Text,
	}, nil
}

func (c *WeatherAPIClient) GetAPIDayWeather(
	ctx context.Context, city string,
) (*domain.DayWeatherResponse, error) {
	url := fmt.Sprintf("%s/forecast.json?key=%s&q=%s&days=1", c.BaseURL, c.APIKey, city)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		logger.Errorf("failed to create WeatherAPI request: %s", err)
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		logger.Errorf("error making request to WeatherAPI: %s", err.Error())
		return nil, err
	}
	defer closeBody(resp.Body, &err)

	if resp.StatusCode != http.StatusOK {
		err = c.processErrorResponse(resp)
		return nil, err
	}

	var result dayWeatherAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		logger.Errorf("error parsing WeatherAPI response: %s", err.Error())
		return nil, customErrors.ErrWeatherDataError
	}

	// Map of required times
	targetHours := map[string]*domain.WeatherResponse{
		"07:00": {},
		"10:00": {},
		"13:00": {},
		"16:00": {},
		"19:00": {},
		"22:00": {},
	}

	const dateTimeSplitStrLength = 2 // "2025-05-17 07:00" -> ["2025-05-17", "07:00"]
	for _, hourData := range result.Forecast.ForecastDay[0].Hour {
		dateTimeParts := strings.Split(hourData.Time, " ")
		if len(dateTimeParts) != dateTimeSplitStrLength {
			continue
		}
		timePart := dateTimeParts[1]

		if target, ok := targetHours[timePart]; ok {
			target.Temperature = hourData.TempC
			target.Humidity = hourData.Humidity
			target.Description = hourData.Condition.Text
		}
	}

	return &domain.DayWeatherResponse{
		SevenAM: *targetHours["07:00"],
		TenAM:   *targetHours["10:00"],
		OnePM:   *targetHours["13:00"],
		FourPM:  *targetHours["16:00"],
		SevenPM: *targetHours["19:00"],
		TenPM:   *targetHours["22:00"],
	}, nil
}
