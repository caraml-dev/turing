package resultlog

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/gojek/turing/engines/router/missionctl/config"
	tu "github.com/gojek/turing/engines/router/missionctl/internal/testutils"
	"github.com/gojek/turing/engines/router/missionctl/turingctx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testSuiteResultLogger struct {
	cfg     *config.AppConfig
	patch   bool
	success bool
}

// mockResultLogger satisfies the TuringResultLogger interface
type mockResultLogger struct {
	mock.Mock
}

// write satisfies the TuringResultLogger interface
func (l *mockResultLogger) write(*TuringResultLogEntry) error {
	l.Called()
	return nil
}

// Tests
func TestInitTuringResultLogger(t *testing.T) {
	// Define test cases
	tests := map[string]testSuiteResultLogger{
		"nop": {
			cfg: &config.AppConfig{
				ResultLogger: config.NopLogger,
			},
			patch:   false,
			success: true,
		},
		"console": {
			cfg: &config.AppConfig{
				ResultLogger: config.ConsoleLogger,
			},
			patch:   false,
			success: true,
		},
		// Dummy BQ project will result in init failure
		"bigquery_failure": {
			cfg: &config.AppConfig{
				ResultLogger: config.BigqueryLogger,
				BigQuery: &config.BQConfig{
					Project:   "test-turing-project",
					BatchLoad: false,
				},
			},
			patch:   false,
			success: false,
		},
		// Patch init BQ method for success
		"bigquery_success": {
			cfg: &config.AppConfig{
				ResultLogger: config.BigqueryLogger,
				BigQuery: &config.BQConfig{
					BatchLoad: false,
				},
			},
			patch:   true,
			success: true,
		},
		// Patch init BQ and init fluentd methods for success
		"bigquery_batch": {
			cfg: &config.AppConfig{
				ResultLogger: config.BigqueryLogger,
				BigQuery: &config.BQConfig{
					BatchLoad: true,
				},
				Fluentd: &config.FluentdConfig{
					Host: "localhost",
					Port: 0,
					Tag:  "test-tag",
				},
			},
			patch:   true,
			success: true,
		},
		"kafka": {
			cfg: &config.AppConfig{
				ResultLogger: config.KafkaLogger,
				Kafka: &config.KafkaConfig{
					Brokers: "brokers",
					Topic:   "topic",
				},
			},
			patch:   true,
			success: true,
		},
		"unrecognised_failure": {
			cfg:     &config.AppConfig{},
			success: false,
		},
	}

	// Test
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			if data.patch {
				// Patch init BQ Client
				monkey.Patch(newBigQueryLogger,
					func(_ string, _ *config.BQConfig) (BigQueryLogger, error) {
						return nil, nil
					})
				// Patch init Fluentd Client
				monkey.Patch(fluent.New,
					func(_ fluent.Config) (*fluent.Fluent, error) { return nil, nil })
				// Patch init Kafka Client
				monkey.Patch(newKafkaLogger,
					func(_ string, _ *config.KafkaConfig) (*KafkaLogger, error) {
						return nil, nil
					})
			}
			err := InitTuringResultLogger(data.cfg)
			if data.patch {
				monkey.UnpatchAll()
			}
			// Assert error status
			assert.Equal(t, data.success, err == nil)
		})
	}
}

func TestGlobalLoggerLog(t *testing.T) {
	// Create test logger
	logger := &mockResultLogger{}
	logger.On("write").Return(nil)

	// Patch global logger
	savedGlobalLogger := globalLogger
	setGlobalLogger(logger)
	// Unpatch
	defer setGlobalLogger(savedGlobalLogger)

	// Call Log
	_ = LogEntry(&TuringResultLogEntry{})

	// Test that the global logger's write was called
	logger.AssertCalled(t, "write")
}

func TestMarshalEmptyRequestLogEntry(t *testing.T) {
	bytes, err := json.Marshal(requestLogEntry{})
	assert.JSONEq(t, `{}`, string(bytes))
	assert.NoError(t, err)
}

func setGlobalLogger(l TuringResultLogger) {
	globalLogger = l
}

// Helper methods for resultlog package tests
func makeTestTuringResultLogEntry(t *testing.T) (context.Context, *TuringResultLogEntry) {
	// Make test request
	req := tu.MakeTestRequest(t, tu.NopHTTPRequestModifier)
	reqBody, err := ioutil.ReadAll(req.Body)
	tu.FailOnError(t, err)

	// Make test context
	ctx := turingctx.NewTuringContext(context.Background())

	// Create a TuringResultLogEntry record and add the data
	timestamp := time.Date(2000, 2, 1, 4, 5, 6, 7, time.UTC)
	entry := NewTuringResultLogEntry(ctx, timestamp, &req.Header, reqBody)
	entry.AddResponse("experiment", []byte(`{"key": "experiment_data"}`), "")
	entry.AddResponse("router", []byte(`{"key": "router_data"}`), "")
	entry.AddResponse("enricher", []byte(`{"key": "enricher_data"}`), "")

	return ctx, entry
}
