package resultlog

import (
	"testing"

	"bou.ke/monkey"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gojek/turing/engines/router/missionctl/config"
	tu "github.com/gojek/turing/engines/router/missionctl/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFluentClient implements the fluentdClient interface
type MockFluentClient struct {
	mock.Mock
}

func (mf *MockFluentClient) Post(tag string, msg interface{}) error {
	mf.Called(tag, msg)
	return nil
}

func TestNewFluentdLogger(t *testing.T) {
	// Create a fluentd client
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
			testLogger, err := newFluentdLogger(&data.cfg)
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

	// Create mock Fluentd client
	fluentClient := &MockFluentClient{}
	fluentClient.On("Post", mock.Anything, mock.Anything).Return(nil)

	// Create new fluentd logger
	testLogger := &FluentdLogger{
		tag:          "test-tag",
		fluentLogger: fluentClient,
	}

	// Validate
	assert.NoError(t, testLogger.write(entry))

	// Check that the expected function call occurred
	record, err := entry.Value()
	tu.FailOnError(t, err)
	fluentClient.AssertCalled(t, "Post", "test-tag", record)
}
