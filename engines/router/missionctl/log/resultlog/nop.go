package resultlog

import upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"

// NopLogger generates instance of NopLog for logging results
type NopLogger struct{}

// NewNopLogger generates an instance of NewNopLogger
func NewNopLogger() *NopLogger {
	return &NopLogger{}
}

// write is a nop method that satisfies the NopLogger interface
func (*NopLogger) write(turLogEntry *TuringResultLogEntry) error {
	return nil
}

// write is a nop method that satisfies the UPILogger interface
func (*NopLogger) WriteUPIRouterLog(routerLog *upiv1.RouterLog) error {
	return nil
}
