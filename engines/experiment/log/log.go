package log

import (
	"errors"

	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
)

var global = newZapLogger()

// Logger interface defines the common set of logging methods.
// Concrete implementations of Logger interface are zapLogger and hclog.hcLogger
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Panic(args ...interface{})

	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Panicw(msg string, keysAndValues ...interface{})

	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Panicf(template string, args ...interface{})

	With(args ...interface{}) Logger

	SetLevel(lvl string)
}

func SetGlobalLogger(l interface{}) {
	switch l := l.(type) {
	case Logger:
		global = l
	case hclog.Logger:
		global = &hcLogger{
			Logger: l,
		}
	case *zap.SugaredLogger:
		global = &zapLogger{
			SugaredLogger: l,
		}
	default:
		panic(errors.New("unsupported logger type"))
	}
}

func Glob() Logger {
	return global
}

func Debug(args ...interface{}) {
	global.Debug(args...)
}

func Info(args ...interface{}) {
	global.Info(args...)
}

func Warn(args ...interface{}) {
	global.Warn(args...)
}

func Error(args ...interface{}) {
	global.Error(args...)
}

func Panic(args ...interface{}) {
	global.Panic(args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	global.Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	global.Infow(msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	global.Warnw(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	global.Errorw(msg, keysAndValues...)
}

func Panicw(msg string, keysAndValues ...interface{}) {
	global.Panicw(msg, keysAndValues...)
}

func Debugf(template string, args ...interface{}) {
	global.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	global.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	global.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	global.Errorf(template, args...)
}

func Panicf(template string, args ...interface{}) {
	global.Panicf(template, args...)
}

func With(args ...interface{}) Logger {
	return global.With(args...)
}

func SetLevel(lvl string) {
	global.SetLevel(lvl)
}
