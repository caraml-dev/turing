package resultlog

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/gojek/turing/engines/router/missionctl/log"
	"github.com/gojek/turing/engines/router/missionctl/log/resultlog/proto/turing"
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

	// Marshal the log entry and unmarshal to get a map of key, value pairs
	m := &protojson.MarshalOptions{
		UseProtoNames: true, // Use the json field name instead of the camel case struct field name
	}
	message := turing.TuringResultLogMessage(*turLogEntry)
	bytes, err := m.Marshal(&message)
	if err != nil {
		return err
	}
	var kvPairs map[string]interface{}
	err = json.Unmarshal(bytes, &kvPairs)
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
