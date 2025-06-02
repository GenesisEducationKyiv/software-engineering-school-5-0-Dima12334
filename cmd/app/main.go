package main

import "weather_forecast_sub/internal/app"

const configsDir = "configs"

func main() {
	app.Run(configsDir)
}
