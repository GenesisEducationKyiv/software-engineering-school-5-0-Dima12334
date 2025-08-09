package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	prodLogEnv = "prod"
	devLogEnv  = "dev"
)

const (
	dirPerm  = 0o750
	filePerm = 0o600
)

func Init(loggerEnv, logFilePath string) error {
	var core zapcore.Core

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	switch loggerEnv {
	case prodLogEnv:
		logDir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(logDir, dirPerm); err != nil {
			return fmt.Errorf("failed to create log dir: %w", err)
		}
		// #nosec G304 -- logFilePath is sanitized and restricted to known safe values
		logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePerm)
		if err != nil {
			return err
		}

		writer := zapcore.AddSync(logFile)
		core = zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), writer, zap.InfoLevel)
	default:
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderCfg), zapcore.AddSync(os.Stdout), zap.DebugLevel,
		)
	}

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	zap.ReplaceGlobals(logger)
	return nil
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
