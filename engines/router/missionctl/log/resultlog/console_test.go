package resultlog

import (
	"context"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
)

// consoleLog is used to Unmarshal the log data
type consoleLog struct {
	Level          string          `json:"level"`
	EventTimestamp string          `json:"event_timestamp"`
	Caller         string          `json:"caller"`
	Msg            string          `json:"msg"`
	TuringReqID    string          `json:"turing_req_id"`
	Request        json.RawMessage `json:"request"`
	Enricher       json.RawMessage `json:"enricher"`
	Router         json.RawMessage `json:"router"`
}

func TestNewConsoleLogger(t *testing.T) {
	testLogger := NewConsoleLogger()
	assert.Equal(t, ConsoleLogger{}, *testLogger)
}

func TestConsoleLoggerWrite(t *testing.T) {
	// Make test request
	req := tu.MakeTestRequest(t, tu.NopHTTPRequestModifier)
	reqBody, err := io.ReadAll(req.Body)
	tu.FailOnError(t, err)

	// Make test context
	ctx := turingctx.NewTuringContext(context.Background())
	turingReqID, err := turingctx.GetRequestID(ctx)
	tu.FailOnError(t, err)

	// Create a logger with a memory sink, for testing console output
	logger, sink, err := tu.NewLoggerWithMemorySink()
	tu.FailOnError(t, err)
	// Patch global logger with the new logger
	globLogger := log.Glob()
	log.SetGlobalLogger(logger)
	// Unpatch
	defer log.SetGlobalLogger(globLogger)

	// Create a new TuringResultLogEntry and send the responses
	timestamp := time.Date(2000, 2, 1, 4, 5, 6, 7, time.UTC)
	entry := NewTuringResultLog(turingReqID, timestamp, req.Header, string(reqBody))
	AddResponse(entry, "enricher", `{"key": "enricher_data"}`, map[string]string{"Content-Encoding": "lz4"},
		"")
	AddResponse(entry, "router", `{"key": "router_data"}`, map[string]string{"Content-Encoding": "gzip"},
		"Error Response")

	// Write the result log using ConsoleLogger
	testLogger := ConsoleLogger{}
	err = testLogger.write(entry)
	tu.FailOnError(t, err)

	// Retrieve the logged contents from the sink
	logData := sink.Bytes()

	// Unmarshal the result
	var logObj consoleLog
	err = json.Unmarshal(logData, &logObj)
	tu.FailOnError(t, err)

	// Validate relevant fields
	assert.Equal(t, "info", logObj.Level)
	assert.Equal(t, "Turing Request Summary", logObj.Msg)
	assert.Equal(t, turingReqID, logObj.TuringReqID)
	assert.Equal(t, "2000-02-01T04:05:06.000000007Z", logObj.EventTimestamp)
	assert.Equal(t,
		json.RawMessage([]byte(
			`{"body":"{\"customer_id\": \"test_customer\"}","header":{"Req_id":"test_req_id"}}`),
		),
		logObj.Request,
	)
	assert.Equal(t,
		json.RawMessage([]byte(
			`{"header":{"Content-Encoding":"lz4"},"response":"{\"key\": \"enricher_data\"}"}`),
		),
		logObj.Enricher,
	)
	assert.Equal(t,
		json.RawMessage([]byte(
			(`{"error":"Error Response","header":{"Content-Encoding":"gzip"},"response":"{\"key\": \"router_data\"}"}`)),
		),
		logObj.Router,
	)
}
