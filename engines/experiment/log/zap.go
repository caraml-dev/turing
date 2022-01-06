package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	*zap.SugaredLogger
}

func newZapLogger(logLevel ...zapcore.Level) Logger {
	cfg := zap.NewProductionConfig()

	if len(logLevel) > 0 {
		cfg.Level = zap.NewAtomicLevelAt(logLevel[0])
	}

	logger, _ := cfg.Build(zap.AddCallerSkip(1))
	return &zapLogger{logger.Sugar()}
}

func (l *zapLogger) With(args ...interface{}) Logger {
	l.SugaredLogger.Desugar().WithOptions()
	return &zapLogger{l.SugaredLogger.With(args...)}
}
