package resultlog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog/proto/turing"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"

	fiberProtocol "github.com/gojek/fiber/protocol"
)

// mockResultLogger satisfies the TuringResultLogger interface
type mockResultLogger struct {
	numOfCalls int32
	result     *turing.TuringResultLogMessage
}

// write satisfies the TuringResultLogger interface
func (l *mockResultLogger) write(log *turing.TuringResultLogMessage) error {
	l.result = log
	atomic.AddInt32(&l.numOfCalls, 1)
	return nil
}

// Helper methods for resultlog package tests
func makeTestTuringResultLog(t *testing.T) (context.Context, *turing.TuringResultLogMessage) {
	// Make test request
	req := tu.MakeTestRequest(t, tu.NopHTTPRequestModifier)
	reqBody, err := io.ReadAll(req.Body)
	tu.FailOnError(t, err)

	// Make test context
	ctx := turingctx.NewTuringContext(context.Background())

	// Get Turing request id
	turingReqID, err := turingctx.GetRequestID(ctx)
	tu.FailOnError(t, err)

	// Create a TuringResultLogEntry record and add the data
	timestamp := time.Date(2000, 2, 1, 4, 5, 6, 7, time.UTC)
	entry := NewTuringResultLog(turingReqID, timestamp, req.Header, string(reqBody))
	AddResponse(entry, "experiment", "", nil, "Error received")
	AddResponse(
		entry,
		"router",
		`{"key": "router_data"}`,
		map[string]string{"Content-Encoding": "gzip", "Content-Type": "text/html,charset=utf-8"},
		"",
	)
	AddResponse(
		entry,
		"enricher",
		`{"key": "enricher_data"}`,
		map[string]string{"Content-Encoding": "lz4", "Content-Type": "text/html,charset=utf-8"},
		"",
	)

	return ctx, entry
}

// Tests
func TestMarshalEmptyLogEntry(t *testing.T) {
	bytes, err := protoJSONMarshaller.Marshal(&turing.TuringResultLogMessage{})
	assert.JSONEq(t, `{}`, string(bytes))
	assert.NoError(t, err)
}

// TestMarshalJSONLogEntry test that the output json into the expected format
func TestMarshalJSONLogEntry(t *testing.T) {
	ctx, logEntry := makeTestTuringResultLog(t)
	turingReqID, err := turingctx.GetRequestID(ctx)
	require.NoError(t, err)

	// Marshal and validate
	bytes, err := protoJSONMarshaller.Marshal(logEntry)
	require.NoError(t, err)
	assert.JSONEq(t, fmt.Sprintf(`{
		"turing_req_id":"%s",
		"event_timestamp":"2000-02-01T04:05:06.000000007Z",
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
	}`, turingReqID), string(bytes))
}

// TestMarshalJSONLogEntry test that the entry can be unmarshall into map[string]interface{}
// with the json unmarshaller into the expected format
func TestTuringResultLogEntryValue(t *testing.T) {
	ctx, logEntry := makeTestTuringResultLog(t)
	turingReqID, err := turingctx.GetRequestID(ctx)
	require.NoError(t, err)

	// Get loggable data and validate

	var kvPairs map[string]interface{}
	// Marshal into bytes
	bytes, err := protoJSONMarshaller.Marshal(logEntry)
	assert.NoError(t, err)
	// Unmarshal into map[string]interface{}
	err = json.Unmarshal(bytes, &kvPairs)
	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"turing_req_id":   turingReqID,
		"event_timestamp": "2000-02-01T04:05:06.000000007Z",
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

func TestResultLogger_LogTuringRouterRequestSummary(t *testing.T) {

	// common var that can be reused to verify mapping
	logger := log.Glob()
	testTime := time.Now()
	predictionID := "123"
	appName := "app-name"

	type args struct {
		predictionID   string
		timestamp      time.Time
		reqHeader      http.Header
		reqBody        []byte
		routerResponse []RouterResponse
	}
	tests := []struct {
		name string
		args args
		want *turing.TuringResultLogMessage
	}{
		{
			name: "ok request response",
			args: args{
				predictionID: predictionID,
				timestamp:    testTime,
				reqHeader:    http.Header{"req": []string{"header"}},
				reqBody:      []byte("req body"),
				routerResponse: []RouterResponse{
					{
						key:    ResultLogKeys.Router,
						header: http.Header{"X-Region": []string{"region-a", "region-b"}},
						body:   []byte("resp body"),
					},
					{
						key:  ResultLogKeys.Enricher,
						body: []byte("enricher body"),
					},
				},
			},
			want: &turing.TuringResultLogMessage{
				TuringReqId:    predictionID,
				EventTimestamp: timestamppb.New(testTime),
				RouterVersion:  appName,
				Request: &turing.Request{
					Header: map[string]string{"req": "header"},
					Body:   "req body",
				},
				Enricher: &turing.Response{
					Response: "enricher body",
					Header:   map[string]string{},
				},
				Router: &turing.Response{
					Response: "resp body",
					Error:    "",
					Header:   map[string]string{"X-Region": "region-a,region-b"},
				},
			},
		},
		{
			name: "error resp",
			args: args{
				predictionID: "",
				timestamp:    time.Time{},
				reqHeader:    nil,
				reqBody:      nil,
				routerResponse: []RouterResponse{
					{
						key:    ResultLogKeys.Router,
						header: http.Header{"fa": []string{"il"}},
						err:    "fail fail",
					},
				},
			},
			want: &turing.TuringResultLogMessage{
				EventTimestamp: timestamppb.New(time.Time{}),
				RouterVersion:  appName,
				Request: &turing.Request{
					Header: map[string]string{},
				},
				Router: &turing.Response{
					Error: "fail fail",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &mockResultLogger{}
			rl := &ResultLogger{
				trl:     mockLogger,
				appName: appName,
			}

			respCh := make(chan RouterResponse, len(tt.args.routerResponse))
			for _, routerResponse := range tt.args.routerResponse {
				respCh <- routerResponse
			}
			close(respCh)

			rl.LogTuringRouterRequestSummary(
				tt.args.predictionID,
				logger,
				tt.args.timestamp,
				tt.args.reqHeader,
				tt.args.reqBody,
				respCh)
		wait:
			for {
				select {
				case <-time.After(3 * time.Second):
					t.Log("logTuringRouterRequestSummary not called")
					t.Fail()
				case <-time.NewTicker(500 * time.Millisecond).C:
					if atomic.LoadInt32(&mockLogger.numOfCalls) == 1 {
						break wait
					}
				}
			}

			assert.NotNil(t, mockLogger.result)
			assert.Equal(t, tt.want, mockLogger.result)
			abc := mockLogger.result
			fmt.Println(abc)
		})
	}
}

// TestSendResponseToLogChannel tests the copyResponseToLogChannel method in logutils.
// Verify that when an error is set, the error message is copied, response is set to null;
// when the error is empty, the response is copied an error field in the log is empty.
// Additionally, verify that the response body is still open for reading after the operation.
func TestSendResponseToLogChannel(t *testing.T) {
	tests := map[string]*errors.TuringError{
		"success": nil,
		"error":   errors.NewTuringError(fmt.Errorf("test error"), fiberProtocol.HTTP),
	}

	rl := InitTuringResultLogger("", NewNopLogger())

	for name, httpErr := range tests {
		t.Run(name, func(t *testing.T) {
			// Make test response
			resp := tu.MakeTestMisisonControlResponse()
			// Capture expected body, for validation
			expectedRespBody := resp.Body()

			// Make response channel
			respCh := make(chan RouterResponse, 1)

			// Push message to channel and close the channel
			rl.SendResponseToLogChannel(context.Background(), respCh, "test", resp, httpErr)
			close(respCh)

			// Read from the channel and validate
			data := <-respCh
			// Check key
			assert.Equal(t, data.key, "test")
			// Check error and response
			if httpErr == nil {
				assert.Empty(t, data.err)
				assert.Equal(t, expectedRespBody, data.body)
			} else {
				assert.Equal(t, data.err, httpErr.Error())
				assert.Empty(t, data.body)
			}
			// Check that the original response body is still readable
			assert.Equal(t, expectedRespBody, resp.Body())
		})
	}
}

// TestNewTuringResultLog verify that the generated proto have mapped values from input and
// FormatHeader is parsing correctly
func TestNewTuringResultLog(t *testing.T) {
	testTime := time.Now()
	predictionID := "123"
	type args[h interface{ http.Header | metadata.MD }] struct {
		predictionID string
		timestamp    time.Time
		header       h
		body         string
	}
	type testCase[h interface{ http.Header | metadata.MD }] struct {
		name string
		args args[h]
		want *turing.TuringResultLogMessage
	}
	tests := []testCase[http.Header]{
		{
			name: "http header",
			args: args[http.Header]{
				header:       http.Header{"X-Region": []string{"region-a", "region-b"}},
				body:         "test",
				timestamp:    testTime,
				predictionID: predictionID,
			},
			want: &turing.TuringResultLogMessage{
				TuringReqId:    predictionID,
				EventTimestamp: timestamppb.New(testTime),
				Request: &turing.Request{
					Header: map[string]string{"X-Region": "region-a,region-b"},
					Body:   "test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t,
				tt.want,
				NewTuringResultLog(
					tt.args.predictionID,
					tt.args.timestamp,
					tt.args.header,
					tt.args.body))
		})
	}
}
