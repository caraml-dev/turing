package resultlog

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	tu "github.com/gojek/turing/engines/router/missionctl/internal/testutils"
	"github.com/gojek/turing/engines/router/missionctl/log"
	"github.com/gojek/turing/engines/router/missionctl/turingctx"
	"github.com/stretchr/testify/assert"
)

// consoleLog is used to Unmarshal the log data
type consoleLog struct {
	Level       string          `json:"level"`
	Ts          float64         `json:"ts"`
	Caller      string          `json:"caller"`
	Msg         string          `json:"msg"`
	TuringReqID string          `json:"turing_req_id"`
	Request     json.RawMessage `json:"request"`
	Enricher    json.RawMessage `json:"enricher"`
	Router      json.RawMessage `json:"router"`
}

func TestNewConsoleLogger(t *testing.T) {
	testLogger := newConsoleLogger()
	assert.Equal(t, ConsoleLogger{}, *testLogger)
}

func TestConsoleLoggerWrite(t *testing.T) {
	// Make test request
	req := tu.MakeTestRequest(t, tu.NopHTTPRequestModifier)
	reqBody, err := ioutil.ReadAll(req.Body)
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
	entry := NewTuringResultLogEntry(ctx, timestamp, &req.Header, reqBody)
	entry.AddResponse("enricher", []byte(`{"key": "enricher_data"}`), "")
	entry.AddResponse("router", []byte(`{"key": "router_data"}`), "Error Response")

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
	assert.Equal(t, 9.49377906e+08, logObj.Ts)
	assert.Equal(t,
		json.RawMessage([]byte(
			`{"header":{"Req_id":["test_req_id"]},"body":{"customer_id":"test_customer"}}`)),
		logObj.Request,
	)
	assert.Equal(t,
		json.RawMessage([]byte(`{"response":{"key":"enricher_data"}}`)),
		logObj.Enricher,
	)
	assert.Equal(t,
		json.RawMessage([]byte(`{"response":{"key":"router_data"},"error":"Error Response"}`)),
		logObj.Router,
	)
}
