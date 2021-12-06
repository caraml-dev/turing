package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var globalLogger = newLogger()

func newLogger(logLevel ...zapcore.Level) *zap.SugaredLogger {
	cfg := zap.NewProductionConfig()

	if len(logLevel) > 0 {
		cfg.Level = zap.NewAtomicLevelAt(logLevel[0])
	}

	logger, _ := cfg.Build(zap.AddCallerSkip(1))
	return logger.Sugar()
}

// SetLogLevelAt creates a new SugaredLogger with provided log level enabled
// and assigns it as the global logger
func SetLogLevelAt(level string) error {
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		return err
	}
	globalLogger = newLogger(lvl)
	return nil
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

func Global() *zap.Logger {
	return globalLogger.Desugar()
}
