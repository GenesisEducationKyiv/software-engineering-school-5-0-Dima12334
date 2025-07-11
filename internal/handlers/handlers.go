package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "weather_forecast_sub/docs"
	"weather_forecast_sub/internal/config"
	"weather_forecast_sub/internal/service"
)

type Handler struct {
	SubscriptionHandler *SubscriptionHandler
	WeatherHandler      *WeatherHandler
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{
		SubscriptionHandler: NewSubscriptionHandler(services.Subscriptions),
		WeatherHandler:      NewWeatherHandler(services.Weather),
	}
}

func (h *Handler) Init(environment string) *gin.Engine {
	router := h.initGinRouter(environment)

	h.initBaseRoutes(router)
	h.initHTMLRoutes(router)
	h.initAPI(router)

	return router
}

func (h *Handler) initGinRouter(environment string) *gin.Engine {
	if environment == config.ProdEnvironment {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.LoadHTMLGlob(config.GetOriginalPath("templates/**/*.html"))

	return router
}

func (h *Handler) initBaseRoutes(router *gin.Engine) {
	// Health check endpoint
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Swagger docs
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

func (h *Handler) initHTMLRoutes(router *gin.Engine) {
	router.GET("/subscribe", h.SubscriptionHandler.ShowSubscribePage)
}

func (h *Handler) initAPI(router *gin.Engine) {
	api := router.Group("/api")
	{
		weather := api.Group("/weather")
		{
			weather.GET("/", h.WeatherHandler.GetWeather)
		}

		subscription := api.Group("")
		{
			subscription.POST("/subscribe", h.SubscriptionHandler.SubscribeEmail)
			subscription.GET("/confirm/:token", h.SubscriptionHandler.ConfirmEmail)
			subscription.GET("/unsubscribe/:token", h.SubscriptionHandler.UnsubscribeEmail)
		}
	}
}
