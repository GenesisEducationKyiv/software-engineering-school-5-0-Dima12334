package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"weather_forecast_sub/internal/server"
	"weather_forecast_sub/pkg/logger"
)

// @title Weather Forecast API
// @version 1.0
// @description Weather API application that allows users to subscribe to weather updates for their city.
// @host weather-forecast-sub-app.onrender.com
// @BasePath /api
// @schemes http https

// @tag.name weather
// @tag.description Weather forecast operations

// @tag.name subscription
// @tag.description Subscription management operations

// Run starts the server.
func Run(srv *server.Server) {
	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("error occurred while running http server: %s\n", err.Error())
		}
	}()
	logger.Info("server started")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const shutdownTimeout = 5 * time.Second
	ctx, shutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		logger.Errorf("failed to stop server: %v", err.Error())
	} else {
		logger.Info("server stopped successfully")
	}
}
