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
	services *service.Services
}

func NewHandler(services *service.Services) *Handler {
	return &Handler{services: services}
}

func (h *Handler) Init(cfg *config.Config) *gin.Engine {
	if cfg.Environment == config.ProdEnvironment {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.LoadHTMLGlob("templates/**/*.html")

	// Init router
	router.GET("ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Swagger docs
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	h.initAPI(router)

	return router
}

func (h *Handler) initAPI(router *gin.Engine) {
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
