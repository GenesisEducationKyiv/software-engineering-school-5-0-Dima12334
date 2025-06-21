package handlers

import (
	"context"
	"errors"
	"net/http"
	"weather_forecast_sub/internal/domain"
	customErrors "weather_forecast_sub/pkg/errors"
	"weather_forecast_sub/pkg/hash"

	"github.com/gin-gonic/gin"
)

type Subscription interface {
	Create(ctx context.Context, inp domain.CreateSubscriptionInput) error
	Confirm(ctx context.Context, token string) error
	Delete(ctx context.Context, token string) error
}

type SubscriptionHandler struct {
	subscriptionService Subscription
}

func NewSubscriptionHandler(subscriptionService Subscription) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
	}
}

type subscribeEmailInput struct {
	Email     string `form:"email" json:"email" binding:"required,email,max=255"`
	City      string `form:"city" json:"city" binding:"required,max=255"`
	Frequency string `form:"frequency" json:"frequency" binding:"oneof=hourly daily"`
}

func (h *SubscriptionHandler) ShowSubscribePage(c *gin.Context) {
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
func (h *SubscriptionHandler) SubscribeEmail(c *gin.Context) {
	var inp subscribeEmailInput

	if err := c.ShouldBind(&inp); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	err := h.subscriptionService.Create(
		c,
		domain.CreateSubscriptionInput{
			Email:     inp.Email,
			City:      inp.City,
			Frequency: inp.Frequency,
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrSubscriptionAlreadyExists):
			c.Status(http.StatusConflict)
		default:
			c.Status(http.StatusInternalServerError)
		}
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
func (h *SubscriptionHandler) ConfirmEmail(c *gin.Context) {
	token := c.Param("token")

	if !hash.IsValidSHA256Hex(token) {
		c.Status(http.StatusBadRequest)
		return
	}

	err := h.subscriptionService.Confirm(c, token)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrSubscriptionNotFound):
			c.Status(http.StatusNotFound)
		default:
			c.Status(http.StatusInternalServerError)
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
func (h *SubscriptionHandler) UnsubscribeEmail(c *gin.Context) {
	token := c.Param("token")

	if !hash.IsValidSHA256Hex(token) {
		c.Status(http.StatusBadRequest)
		return
	}

	err := h.subscriptionService.Delete(c, token)
	if err != nil {
		switch {
		case errors.Is(err, customErrors.ErrSubscriptionNotFound):
			c.Status(http.StatusNotFound)
		default:
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	c.Status(http.StatusOK)
}
