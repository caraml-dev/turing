package resultlog

import (
	"testing"

	"github.com/stretchr/testify/assert"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
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

func TestNopLoggerWriteUPIRouterLog(t *testing.T) {
	testLogger := &NopLogger{}
	err := testLogger.WriteUPIRouterLog(&upiv1.RouterLog{})
	assert.NoError(t, err)
}
