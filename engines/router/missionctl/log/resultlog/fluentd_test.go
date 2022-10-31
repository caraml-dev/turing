package resultlog

import (
	"testing"

	"bou.ke/monkey"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
)

// MockFluentClient implements the fluentdClient interface
type MockFluentClient struct {
	mock.Mock
}

func (mf *MockFluentClient) Post(tag string, msg interface{}) error {
	mf.Called(tag, msg)
	return nil
}

// MockBqLogger implements the BigQueryLogger interface
type MockBqLogger struct {
	mock.Mock
}

func (bq *MockBqLogger) write(*TuringResultLogEntry) error {
	return nil
}

func (bq *MockBqLogger) getLogData(t *TuringResultLogEntry) interface{} {
	bq.Called(t)
	return t
}

func TestNewFluentdLogger(t *testing.T) {
	// Create test BQ logger and fluent client
	bqLogger := bigQueryLogger{
		dataset:  "test-dataset",
		table:    "test-table",
		bqClient: nil,
	}
	fc := &fluent.Fluent{}

	tests := map[string]struct {
		cfg      config.FluentdConfig
		expected FluentdLogger
		success  bool
		err      string
	}{
		"success": {
			cfg: config.FluentdConfig{
				Host: "localhost",
				Port: 80,
				Tag:  "test-tag",
			},
			expected: FluentdLogger{
				tag:          "test-tag",
				bqLogger:     &bqLogger,
				fluentLogger: fc,
			},
			success: true,
		},
		"failure | empty tag": {
			cfg: config.FluentdConfig{
				Host: "localhost",
				Port: 80,
				Tag:  "",
			},
			success: false,
			err:     "Fluentd Tag must be configured",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Patch the new fluentd client init
			monkey.Patch(fluent.New,
				func(_ fluent.Config) (*fluent.Fluent, error) { return fc, nil })
			defer monkey.Unpatch(fluent.New)
			// Create the new logger and validate
			testLogger, err := newFluentdLogger(&data.cfg, &bqLogger)
			assert.Equal(t, data.success, err == nil)
			if data.success {
				tu.FailOnNil(t, testLogger)
				assert.Equal(t, data.expected, *testLogger)
			} else {
				tu.FailOnNil(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}

func TestFuentdLoggerWrite(t *testing.T) {
	// Create test log object
	_, entry := makeTestTuringResultLogEntry(t)

	// Create mock BQ Logger
	bqLogger := &MockBqLogger{}
	bqLogger.On("getLogData", mock.Anything).Return(entry)
	// Create mock Fluentd client
	fluentClient := &MockFluentClient{}
	fluentClient.On("Post", mock.Anything, mock.Anything).Return(nil)

	// Create new fluentd logger
	testLogger := &FluentdLogger{
		tag:          "test-tag",
		bqLogger:     bqLogger,
		fluentLogger: fluentClient,
	}

	// Validate
	assert.NoError(t, testLogger.write(entry))
	// Check that the expected function calls occurred
	bqLogger.AssertCalled(t, "getLogData", entry)
	fluentClient.AssertCalled(t, "Post", "test-tag", entry)
}
