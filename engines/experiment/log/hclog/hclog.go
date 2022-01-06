package hclog

import (
	"fmt"
	"github.com/gojek/turing/engines/experiment/log"
	"github.com/hashicorp/go-hclog"
)

func init() {
	log.SetGlobalLogger(newHCLogger())
}

func newHCLogger() log.Logger {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:       hclog.Warn,
		JSONFormat:  true,
		DisableTime: true,
	})

	return &hcLogger{logger}
}

type hcLogger struct {
	hclog.Logger
}

func (l *hcLogger) Debug(args ...interface{}) {
	l.Logger.Debug(fmt.Sprint(args...))
}

func (l *hcLogger) Info(args ...interface{}) {
	l.Logger.Info(fmt.Sprint(args...))
}

func (l *hcLogger) Warn(args ...interface{}) {
	l.Logger.Warn(fmt.Sprint(args...))
}

func (l *hcLogger) Error(args ...interface{}) {
	l.Logger.Error(fmt.Sprint(args...))
}

func (l *hcLogger) Panic(args ...interface{}) {
	msg := fmt.Sprint(args...)
	l.Logger.Error(msg)
	panic(msg)
}

func (l *hcLogger) Debugw(msg string, keysAndValues ...interface{}) {
	l.Logger.Debug(msg, keysAndValues...)
}

func (l *hcLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.Logger.Info(msg, keysAndValues...)
}

func (l *hcLogger) Warnw(msg string, keysAndValues ...interface{}) {
	l.Logger.Warn(msg, keysAndValues...)
}

func (l *hcLogger) Errorw(msg string, keysAndValues ...interface{}) {
	l.Logger.Error(msg, keysAndValues...)
}

func (l *hcLogger) Panicw(msg string, keysAndValues ...interface{}) {
	l.Logger.Error(msg, keysAndValues...)
	panic(msg)
}

func (l *hcLogger) Debugf(template string, args ...interface{}) {
	l.Logger.Debug(fmt.Sprintf(template, args...))
}

func (l *hcLogger) Infof(template string, args ...interface{}) {
	l.Logger.Info(fmt.Sprintf(template, args...))
}

func (l *hcLogger) Warnf(template string, args ...interface{}) {
	l.Logger.Warn(fmt.Sprintf(template, args...))
}

func (l *hcLogger) Errorf(template string, args ...interface{}) {
	l.Logger.Error(fmt.Sprintf(template, args...))
}

func (l *hcLogger) Panicf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	l.Logger.Error(msg)
	panic(msg)
}

func (l *hcLogger) With(args ...interface{}) log.Logger {
	return &hcLogger{l.Logger.With(args...)}
}
