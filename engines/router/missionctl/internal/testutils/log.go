package testutils

/*
This file provides type definitions and methods that are useful for testing
the zap logger's output. A memory sink is implemented which captures all output in
the sink, which can be accessed for validation.
*/

import (
	"bytes"
	"net/url"

	"go.uber.org/zap"
)

// MemorySink implements zap.Sink by writing all messages to a buffer.
type MemorySink struct {
	*bytes.Buffer
}

// Close is a nop method to satisfy the zap.Sink interface
func (s *MemorySink) Close() error { return nil }

// Sync is a nop method to satisfy the zap.Sink interface
func (s *MemorySink) Sync() error { return nil }

// Global sink
var globalSink *MemorySink

// NewLoggerWithMemorySink creates a new zap logger with a memory sink, to which all
// output is redirected. Calling sink.String() / sink.Bytes() ... will give access to
// its contents.
func NewLoggerWithMemorySink() (*zap.SugaredLogger, *MemorySink, error) {
	// Init the global sink, and register it with zap
	if globalSink == nil {
		globalSink = &MemorySink{new(bytes.Buffer)}
		_ = zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
			return globalSink, nil
		})
	}
	// Reset sink when making new logger (for consecutive tests, clear the shared sink)
	globalSink.Reset()

	// Using the default prod config with Debug level, set the memory sink as the output
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	cfg.OutputPaths = []string{"memory://"}

	logger, err := cfg.Build()
	if err != nil {
		return nil, globalSink, err
	}
	return logger.Sugar(), globalSink, nil
}
