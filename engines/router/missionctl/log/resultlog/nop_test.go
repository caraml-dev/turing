package resultlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNopLogger(t *testing.T) {
	testLogger := NewNopLogger()
	assert.Equal(t, NopLogger{}, *testLogger)
}

func TestNopLoggerWrite(t *testing.T) {
	testLogger := &NopLogger{}
	err := testLogger.write(&TuringResultLogEntry{})
	assert.NoError(t, err)
}
