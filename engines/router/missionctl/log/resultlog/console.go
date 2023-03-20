package resultlog

import (
	"encoding/json"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog/proto/turing"
)

// ConsoleLogger generates instance of ConsoleLog for logging results
type ConsoleLogger struct{}

// NewConsoleLogger generates an instance of ConsoleLogger
func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

// write logs the given TuringResultLogEntry to the console
func (*ConsoleLogger) write(turLogEntry *turing.TuringResultLogMessage) error {
	// Get context-specific upiLogger
	logger := log.Glob()

	var kvPairs map[string]interface{}
	// Marshal into bytes
	bytes, err := protoJSONMarshaller.Marshal(turLogEntry)
	if err != nil {
		return errors.Wrapf(err, "Error marshaling the result log")
	}
	// Unmarshal into map[string]interface{}
	err = json.Unmarshal(bytes, &kvPairs)
	if err != nil {
		return errors.Wrapf(err, "Error unmarshaling the result log")
	}
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
	return nil
}
