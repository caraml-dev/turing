package resultlog

import (
	"github.com/gojek/turing/engines/router/missionctl/log"
)

// ConsoleLogger generates instance of ConsoleLog for logging results
type ConsoleLogger struct{}

// newConsoleLogger generates an instance of ConsoleLogger
func newConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

// write logs the given TuringResultLogEntry to the console
func (*ConsoleLogger) write(turLogEntry *TuringResultLogEntry) error {
	// Get context-specific logger
	logger := log.WithContext(*turLogEntry.ctx)
	// Add request and responses
	data := []interface{}{}
	// Use the timestamp in the log record
	data = append(data, "event_timestamp", turLogEntry.timestamp)
	// Add the request and responses
	data = append(data, "request", turLogEntry.request)
	for k, v := range turLogEntry.responses {
		data = append(data, k, v)
	}
	// Write the log
	logger.Infow("Turing Request Summary", data...)
	_ = logger.Sync()
	return nil
}
