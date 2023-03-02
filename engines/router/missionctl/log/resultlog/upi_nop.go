package resultlog

import upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"

// UPINopLogger generates instance of NopLog for logging results
type UPINopLogger struct{}

// NewUPINopLogger generates an instance of NewUPINopLogger
func NewUPINopLogger() *UPINopLogger {
	return &UPINopLogger{}
}

// write is a nop method that satisfies the UPILogger interface
func (*UPINopLogger) write(routerLog *upiv1.RouterLog) error {
	return nil
}
