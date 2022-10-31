package log

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
)

// Init the global logger to default prod settings. Calling InitGlobalLogger()
// will reset this.
var globalLogger = newDefaultGlobalLogger()

// Logger interface captures the logging functions exposed for the turing router,
// abstracting away the underlying logging library.
type Logger interface {
	Debugw(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Infof(template string, args ...interface{})
	Infow(template string, args ...interface{})
	Panicf(template string, args ...interface{})
	Warnw(template string, args ...interface{})
	Sync() error
}

func newDefaultGlobalLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	return logger.Sugar()
}

// InitGlobalLogger creates a new SugaredLogger and assigns it as the global logger
func InitGlobalLogger(appCfg *config.AppConfig) {
	cfg := zap.NewProductionConfig()
	cfg.DisableCaller = true
	// Add the App name to the root logger
	cfg.InitialFields = map[string]interface{}{
		"router_version": appCfg.Name,
	}

	setLogLevel(&cfg, appCfg.LogLevel)

	// Build logger
	logger, _ := cfg.Build()
	globalLogger = logger.Sugar()
}

// SetGlobalLogger takes a Logger instance as the input and sets it as the global
// logger, useful for testing.
func SetGlobalLogger(l *zap.SugaredLogger) {
	globalLogger = l
}

// Glob simply returns the global logger
func Glob() *zap.SugaredLogger {
	return globalLogger
}

// WithContext returns a zap logger with as much context as possible
func WithContext(ctx context.Context) *zap.SugaredLogger {
	newLogger := globalLogger
	// If context is passed, return a logger with the available key-value(s)
	if ctx != nil {
		// Retrieve the turing context properties from the context as a list
		// of alternating key, val
		fields := turingctx.GetKeyValsFromContext(ctx)
		if len(fields) > 0 {
			newLogger = newLogger.With(fields...)
		}
	}
	return newLogger
}

// setLogLevel takes in a zap config and a LogLevel and sets the logging
// level in the config accordingly
func setLogLevel(cfg *zap.Config, logLvl config.LogLevel) {
	var zapLevel zapcore.Level

	switch logLvl {
	case config.DebugLevel:
		zapLevel = zap.DebugLevel
	case config.WarnLevel:
		zapLevel = zap.WarnLevel
	case config.ErrorLevel:
		zapLevel = zap.ErrorLevel
	default:
		// Use INFO by default
		zapLevel = zapcore.InfoLevel
	}

	cfg.Level = zap.NewAtomicLevelAt(zapLevel)
}
