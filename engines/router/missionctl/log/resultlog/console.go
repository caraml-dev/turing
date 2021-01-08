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
	logger := log.Glob()

	// Get the loggable data
	kvPairs, err := turLogEntry.Value()
	if err != nil {
		return err
	}

	// Copy keys and values into an array
	data := []interface{}{}
	for k, v := range kvPairs {
		data = append(data, k, v)
	}

	// Write the log
	logger.Infow("Turing Request Summary", data...)
	_ = logger.Sync()
	return nil
}
