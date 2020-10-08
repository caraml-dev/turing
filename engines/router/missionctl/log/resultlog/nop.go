package resultlog

// NopLogger generates instance of NopLog for logging results
type NopLogger struct{}

// newNopLogger generates an instance of NewNopLogger
func newNopLogger() *NopLogger {
	return &NopLogger{}
}

// write is a nop method that satisfies the NopLogger interface
func (*NopLogger) write(turLogEntry *TuringResultLogEntry) error {
	return nil
}
