package logger

import (
	"weather_forecast_sub/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	prodLogEnv = "prod"
	devLogEnv  = "dev"
)

func Init(loggerCfg config.LoggerConfig) error {
	var cfg zap.Config

	if loggerCfg.LoggerEnv == prodLogEnv {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.DisableStacktrace = true

	baseLogger, err := cfg.Build()
	if err != nil {
		return err
	}

	logger := baseLogger.WithOptions(zap.AddCallerSkip(1))
	zap.ReplaceGlobals(logger)

	return err
}

func Debug(msg string) {
	zap.S().Debug(msg)
}

func Debugf(msg string, args ...interface{}) {
	zap.S().Debugf(msg, args...)
}

func Info(msg string) {
	zap.S().Info(msg)
}

func Infof(msg string, args ...interface{}) {
	zap.S().Infof(msg, args...)
}

func Warn(msg string) {
	zap.S().Warn(msg)
}

func Warnf(msg string, args ...interface{}) {
	zap.S().Warnf(msg, args...)
}

func Error(msg string) {
	zap.S().Error(msg)
}

func Errorf(msg string, args ...interface{}) {
	zap.S().Errorf(msg, args...)
}

func Fatal(msg string) {
	zap.S().Fatal(msg)
}

func Fatalf(msg string, args ...interface{}) {
	zap.S().Fatalf(msg, args...)
}
