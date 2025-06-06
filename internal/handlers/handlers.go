package handlers

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"weather_forecast_sub/internal/service"

	_ "weather_forecast_sub/docs"
)

type Handler struct {
	services *service.Services
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{services: services}
}

func (h *Handler) Init() *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob("templates/**/*.html")

	// Init router
	router.GET("ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Swagger docs
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	h.initApi(router)

	return router
}

func (h *Handler) initApi(router *gin.Engine) {
	router.GET("/subscribe", h.ShowSubscribePage)

	api := router.Group("/api")
	{
		weather := api.Group("/weather")
		{
			weather.GET("/", h.GetWeather)
		}

		subscription := api.Group("")
		{
			subscription.POST("/subscribe", h.SubscribeEmail)
			subscription.GET("/confirm/:token", h.ConfirmEmail)
			subscription.GET("/unsubscribe/:token", h.UnsubscribeEmail)
		}
	}
}
