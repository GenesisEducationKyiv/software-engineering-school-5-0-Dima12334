package handlers

import (
	"errors"
	"net/http"
	"net/url"
	customErrors "weather_forecast_sub/pkg/errors"

	"github.com/gin-gonic/gin"
)

type weatherResponse struct {
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
	Description string  `json:"description"`
}

// GetWeather godoc
// @Summary Get current weather for a city
// @Description Returns the current weather forecast for the specified city using WeatherAPI.com.
// @Tags weather
// @Accept json
// @Produce json
// @Param city query string true "City name for weather forecast"
// @Success 200 {object} weatherResponse
// @Failure 400 "Invalid request"
// @Failure 404 "City not found"
// @Router /weather [get]
func (h *Handler) GetWeather(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.Status(http.StatusBadRequest)
		return
	}
	escapedCity := url.QueryEscape(city)

	weather, err := h.services.Weather.GetCurrentWeather(c, escapedCity)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrCityNotFound):
			c.Status(http.StatusNotFound)
		default:
			c.Status(http.StatusBadRequest)
		}
		return
	}

	c.JSON(http.StatusOK, weatherResponse(*weather))
}
