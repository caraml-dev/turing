package log

import "strings"

var global = newZapLogger()

type Level int32

const (
	// NoLevel is a special level used to indicate that no level has been
	// set and allow for a default to be used.
	NoLevel Level = 0

	// DebugLevel information for programmer lowlevel analysis.
	DebugLevel Level = 2

	// InfoLevel information about steady state operations.
	InfoLevel Level = 3

	// WarnLevel information about rare but handled events.
	WarnLevel Level = 4

	// ErrorLevel information about unrecoverable events.
	ErrorLevel Level = 5
)

// LevelFromString returns a Level type for the named log level, or "NoLevel" if
// the level string is invalid. This facilitates setting the log level via
// config or environment variable by name in a predictable way.
func LevelFromString(levelStr string) Level {
	levelStr = strings.ToLower(strings.TrimSpace(levelStr))
	switch levelStr {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return NoLevel
	}
}

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case NoLevel:
		return "none"
	default:
		return "unknown"
	}
}

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

	SetLevel(l Level)
}

func SetGlobalLogger(l Logger) {
	global = l
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

func SetLevel(l Level) {
	global.SetLevel(l)
}
