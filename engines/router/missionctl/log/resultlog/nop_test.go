package resultlog

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog/proto/turing"
)

func TestNewNopLogger(t *testing.T) {
	testLogger := NewNopLogger()
	assert.Equal(t, NopLogger{}, *testLogger)
}

func TestNopLoggerWrite(t *testing.T) {
	testLogger := &NopLogger{}
	err := testLogger.write(&turing.TuringResultLogMessage{})
	assert.NoError(t, err)
}
