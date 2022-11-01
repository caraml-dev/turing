package resultlog

import (
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/fluent/fluent-logger-golang/fluent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
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
func TestMarshalEmptyLogEntry(t *testing.T) {
	bytes, err := json.Marshal(&TuringResultLogEntry{})
	assert.JSONEq(t, `{}`, string(bytes))
	assert.NoError(t, err)
}

func TestMarshalJSONLogEntry(t *testing.T) {
	_, logEntry := makeTestTuringResultLogEntry(t)
	// Set the Turing Request Id to a known value
	logEntry.TuringReqId = "test-req-id"

	// Marshal and validate
	bytes, err := json.Marshal(logEntry)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"turing_req_id":"test-req-id",
		"event_timestamp":"2000-02-01T04:05:06.000000007Z",
		"router_version":"test-app-name",
		"request":{"header":{"Req_id":"test_req_id"},"body":"{\"customer_id\": \"test_customer\"}"},
		"experiment":{"error":"Error received"},
		"enricher":{
			"response":"{\"key\": \"enricher_data\"}", 
			"header":{"Content-Encoding":"lz4","Content-Type":"text/html,charset=utf-8"}
		},
		"router":{
			"response":"{\"key\": \"router_data\"}",
			"header":{"Content-Encoding":"gzip","Content-Type":"text/html,charset=utf-8"}
		}
	}`, string(bytes))
}

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
					Brokers:             "brokers",
					Topic:               "topic",
					SerializationFormat: config.JSONSerializationFormat,
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
					func(_ *config.BQConfig) (BigQueryLogger, error) {
						return nil, nil
					})
				// Patch init Fluentd Client
				monkey.Patch(fluent.New,
					func(_ fluent.Config) (*fluent.Fluent, error) { return nil, nil })
				// Patch init Kafka Client
				monkey.Patch(newKafkaLogger,
					func(_ *config.KafkaConfig) (*KafkaLogger, error) {
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

func TestTuringResultLogEntryValue(t *testing.T) {
	_, logEntry := makeTestTuringResultLogEntry(t)
	// Set the Turing Request Id to a known value
	logEntry.TuringReqId = "test-req-id"

	// Get loggable data and validate
	kvPairs, err := logEntry.Value()
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"turing_req_id":   "test-req-id",
		"event_timestamp": "2000-02-01T04:05:06.000000007Z",
		"router_version":  "test-app-name",
		"request": map[string]interface{}{
			"header": map[string]interface{}{
				"Req_id": "test_req_id",
			},
			"body": "{\"customer_id\": \"test_customer\"}",
		},
		"experiment": map[string]interface{}{
			"error": "Error received",
		},
		"enricher": map[string]interface{}{
			"response": "{\"key\": \"enricher_data\"}",
			"header":   map[string]interface{}{"Content-Encoding": "lz4", "Content-Type": "text/html,charset=utf-8"},
		},
		"router": map[string]interface{}{
			"response": "{\"key\": \"router_data\"}",
			"header":   map[string]interface{}{"Content-Encoding": "gzip", "Content-Type": "text/html,charset=utf-8"},
		},
	}, kvPairs)
}

func setGlobalLogger(l TuringResultLogger) {
	globalLogger = l
}

// Helper methods for resultlog package tests
func makeTestTuringResultLogEntry(t *testing.T) (context.Context, *TuringResultLogEntry) {
	// Make test request
	req := tu.MakeTestRequest(t, tu.NopHTTPRequestModifier)
	reqBody, err := io.ReadAll(req.Body)
	tu.FailOnError(t, err)

	// Make test context
	ctx := turingctx.NewTuringContext(context.Background())

	// Set the package var for router version
	appName = "test-app-name"

	// Create a TuringResultLogEntry record and add the data
	timestamp := time.Date(2000, 2, 1, 4, 5, 6, 7, time.UTC)
	entry := NewTuringResultLogEntry(ctx, timestamp, req.Header, string(reqBody))
	entry.AddResponse("experiment", "", nil, "Error received")
	entry.AddResponse(
		"router",
		`{"key": "router_data"}`,
		map[string]string{"Content-Encoding": "gzip", "Content-Type": "text/html,charset=utf-8"},
		"",
	)
	entry.AddResponse(
		"enricher",
		`{"key": "enricher_data"}`,
		map[string]string{"Content-Encoding": "lz4", "Content-Type": "text/html,charset=utf-8"},
		"",
	)

	return ctx, entry
}
