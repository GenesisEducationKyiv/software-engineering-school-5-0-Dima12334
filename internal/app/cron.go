package app

import (
	"time"
	"weather_forecast_sub/internal/service"
	"weather_forecast_sub/pkg/logger"

	"github.com/robfig/cron/v3"
)

type CronRunner struct {
	services *service.Services
	cron     *cron.Cron
}

func NewCronRunner(services *service.Services) *CronRunner {
	return &CronRunner{
		services: services,
		cron:     cron.New(cron.WithLocation(time.UTC)),
	}
}

func (c *CronRunner) Start() {
	c.registerTasks()
	c.cron.Start()
}

func (c *CronRunner) registerTasks() {
	// Top of each hour (7:00, 8:00, 9:00, etc.)
	c.addTask("0 * * * *", c.hourlyWeatherEmailTask, "hourly weather email sending")
	// Daily at 7AM
	c.addTask("0 7 * * *", c.dailyWeatherEmailTask, "daily weather email sending")
}

func (c *CronRunner) addTask(schedule string, taskFunc func(), taskName string) {
	_, err := c.cron.AddFunc(schedule, func() {
		logger.Debugf("start %s", taskName)
		taskFunc()
	})
	if err != nil {
		logger.Fatalf("failed to schedule %s: %v", taskName, err)
	}
}

func (c *CronRunner) hourlyWeatherEmailTask() {
	err := c.services.Subscriptions.SendHourlyWeatherForecast()
	if err != nil {
		logger.Errorf("hourly weather task error: %s", err.Error())
	}
}

func (c *CronRunner) dailyWeatherEmailTask() {
	err := c.services.Subscriptions.SendDailyWeatherForecast()
	if err != nil {
		logger.Errorf("daily weather task error: %s", err.Error())
	}
}
