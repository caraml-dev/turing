package resultlog

import "github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog/proto/turing"

// NopLogger generates instance of NopLog for logging results
type NopLogger struct{}

// NewNopLogger generates an instance of NewNopLogger
func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

// write is a nop method that satisfies the NopLogger interface
func (*NopLogger) write(turLogEntry *turing.TuringResultLogMessage) error {
	return nil
}
