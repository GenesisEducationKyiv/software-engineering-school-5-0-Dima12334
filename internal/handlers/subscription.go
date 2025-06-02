package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"weather_forecast_sub/internal/service"
	customErrors "weather_forecast_sub/pkg/errors"
	"weather_forecast_sub/pkg/hash"
)

type subscribeEmailInput struct {
	Email     string `form:"email" json:"email" binding:"required,email,max=255"`
	City      string `form:"city" json:"city" binding:"required,max=255"`
	Frequency string `form:"frequency" json:"frequency" binding:"oneof=hourly daily"`
}

func (h *Handler) ShowSubscribePage(c *gin.Context) {
	c.HTML(http.StatusOK, "subscribe.html", gin.H{})
}

// SubscribeEmail godoc
// @Summary Subscribe to weather updates
// @Description Subscribe an email to receive weather updates for a specific city with chosen frequency.
// @Tags subscription
// @Accept  json
// @Accept  x-www-form-urlencoded
// @Produce json
// @Param email formData string true "Email address to subscribe"
// @Param city formData string true "City for weather updates"
// @Param frequency formData string true "Frequency of updates (hourly or daily)" Enums(hourly, daily)
// @Success 200 "Subscription successful. Confirmation email sent."
// @Failure 400 "Invalid input"
// @Failure 409 "Email already subscribed"
// @Router /subscribe [post]
func (h *Handler) SubscribeEmail(c *gin.Context) {
	var inp subscribeEmailInput

	if err := c.ShouldBind(&inp); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err := h.services.Subscriptions.Create(
		c,
		service.CreateSubscriptionInput{
			Email:     inp.Email,
			City:      inp.City,
			Frequency: inp.Frequency,
		},
	)
	if err != nil {
		if errors.Is(err, customErrors.ErrSubscriptionAlreadyExists) {
			c.Status(http.StatusConflict)
			return
		}

		c.Status(http.StatusBadRequest)
		return
	}

	c.Status(http.StatusOK)
}

// ConfirmEmail godoc
// @Summary Confirm email subscription
// @Description Confirms a subscription using the token sent in the confirmation email.
// @Tags subscription
// @Accept json
// @Produce json
// @Param token path string true "Confirmation token"
// @Success 200 "Subscription confirmed successfully"
// @Failure 400 "Invalid token"
// @Failure 404 "Token not found"
// @Router /confirm/{token} [get]
func (h *Handler) ConfirmEmail(c *gin.Context) {
	token := c.Param("token")

	if !hash.IsValidSHA256Hex(token) {
		c.Status(http.StatusBadRequest)
		return
	}

	err := h.services.Subscriptions.Confirm(c, token)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrSubscriptionNotFound):
			c.Status(http.StatusNotFound)
		default:
			c.Status(http.StatusBadRequest)
		}
		return
	}

	c.Status(http.StatusOK)
}

// UnsubscribeEmail godoc
// @Summary Unsubscribe from weather updates
// @Description Unsubscribes an email from weather updates using the token sent in emails.
// @Tags subscription
// @Accept json
// @Produce json
// @Param token path string true "Unsubscribe token"
// @Success 200 "Unsubscribed successfully"
// @Failure 400 "Invalid token"
// @Failure 404 "Token not found"
// @Router /unsubscribe/{token} [get]
func (h *Handler) UnsubscribeEmail(c *gin.Context) {
	token := c.Param("token")

	if !hash.IsValidSHA256Hex(token) {
		c.Status(http.StatusBadRequest)
		return
	}

	err := h.services.Subscriptions.Delete(c, token)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrSubscriptionNotFound):
			c.Status(http.StatusNotFound)
		default:
			c.Status(http.StatusBadRequest)
		}
		return
	}

	c.Status(http.StatusOK)
}
