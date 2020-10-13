package log

import (
	"go.uber.org/zap"
)

var globalLogger = NewLogger()

// NewLogger create a new SugaredLogger
func NewLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction(zap.AddCallerSkip(1))
	return logger.Sugar()
}

// Infof uses fmt.Sprintf to log a templated message.
func Infof(template string, args ...interface{}) {
	globalLogger.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func Warnf(template string, args ...interface{}) {
	globalLogger.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func Errorf(template string, args ...interface{}) {
	globalLogger.Errorf(template, args...)
}

// Debugf uses fmt.Sprintf to log a templated message.
func Debugf(template string, args ...interface{}) {
	globalLogger.Debugf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message.
func Fatalf(template string, args ...interface{}) {
	globalLogger.Fatalf(template, args...)
}

// Panicf uses fmt.Sprintf to log a templated message.
func Panicf(template string, args ...interface{}) {
	globalLogger.Panicf(template, args...)
}
