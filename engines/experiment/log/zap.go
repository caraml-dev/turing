package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	*zap.SugaredLogger
	cfg *zap.Config
}

func newZapLogger(logLevel ...zapcore.Level) Logger {
	cfg := zap.NewProductionConfig()

	if len(logLevel) > 0 {
		cfg.Level = zap.NewAtomicLevelAt(logLevel[0])
	}

	logger, _ := cfg.Build(zap.AddCallerSkip(1))
	return &zapLogger{logger.Sugar(), &cfg}
}

func (l *zapLogger) With(args ...interface{}) Logger {
	return &zapLogger{l.SugaredLogger.With(args...), l.cfg}
}

func (l *zapLogger) SetLevel(lvl string) {
	var zapLvl zapcore.Level
	if err := zapLvl.UnmarshalText([]byte(lvl)); err != nil {
		l.Warnf("failed to set %s log level: %v", lvl, err)
	} else {
		l.cfg.Level = zap.NewAtomicLevelAt(zapLvl)
	}
}
