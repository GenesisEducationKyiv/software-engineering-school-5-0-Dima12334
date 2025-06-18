package app

import (
	"context"
	"time"
	"weather_forecast_sub/internal/service"
	"weather_forecast_sub/pkg/logger"

	"github.com/robfig/cron/v3"
)

type Cron interface {
	Start()
	Stop()
	AddTask(schedule string, taskFunc func(), taskName string)
}

type CronRunner struct {
	service service.WeatherForecastSender
	cron    *cron.Cron
}

func NewCronRunner(service service.WeatherForecastSender) *CronRunner {
	return &CronRunner{
		service: service,
		cron:    cron.New(cron.WithLocation(time.UTC)),
	}
}

func (c *CronRunner) Start() {
	c.registerTasks()
	c.cron.Start()
}

func (c *CronRunner) Stop() {
	c.cron.Stop().Done()
}

func (c *CronRunner) registerTasks() {
	// Top of each hour (7:00, 8:00, 9:00, etc.)
	c.AddTask("0 * * * *", c.hourlyWeatherEmailTask, "hourly weather email sending")
	// Daily at 7AM
	c.AddTask("0 7 * * *", c.dailyWeatherEmailTask, "daily weather email sending")
}

func (c *CronRunner) AddTask(schedule string, taskFunc func(), taskName string) {
	_, err := c.cron.AddFunc(schedule, func() {
		logger.Debugf("start %s", taskName)
		taskFunc()
	})
	if err != nil {
		logger.Errorf("failed to schedule %s: %v", taskName, err)
	}
}

func (c *CronRunner) hourlyWeatherEmailTask() {
	ctx := context.Background()
	err := c.service.SendHourlyWeatherForecast(ctx)
	if err != nil {
		logger.Errorf("hourly weather task error: %s", err.Error())
	}
}

func (c *CronRunner) dailyWeatherEmailTask() {
	ctx := context.Background()
	err := c.service.SendDailyWeatherForecast(ctx)
	if err != nil {
		logger.Errorf("daily weather task error: %s", err.Error())
	}
}
