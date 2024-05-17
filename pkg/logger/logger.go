package logger

import (
	"fmt"

	"go.uber.org/zap"
)

type Logger interface {
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
	Warning(format string, v ...interface{})
	Debug(format string, v ...interface{})
}

type ZapLogger struct {
	logger *zap.SugaredLogger
	zap    *zap.Logger
}

var release string

func NewLogger(env string) (*ZapLogger, error) {
	var logger *zap.Logger
	var err error
	switch env {
	case "production":
		logger, err = zap.NewProduction()
	case "development":
		logger, err = zap.NewDevelopment()
	default:
		logger = zap.NewNop()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create zap logger: %w", err)
	}

	logger = logger.With(zap.String("release", release)).WithOptions(zap.AddCallerSkip(1))
	defer logger.Sync()
	sugar := logger.Sugar()

	return &ZapLogger{sugar, logger}, nil
}

func (z *ZapLogger) GetZapLogger() *zap.Logger {
	return z.zap
}

func (z *ZapLogger) Info(format string, v ...interface{}) {
	z.logger.Infof(format, v...)
}

func (z *ZapLogger) Error(format string, v ...interface{}) {
	z.logger.Errorf(format, v...)
}

func (z *ZapLogger) Warning(format string, v ...interface{}) {
	z.logger.Warnf(format, v...)
}

func (z *ZapLogger) Debug(format string, v ...interface{}) {
	z.logger.Debugf(format, v...)
}
