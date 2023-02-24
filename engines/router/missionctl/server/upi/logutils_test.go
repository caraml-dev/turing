package upi

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
	"google.golang.org/grpc/metadata"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	fiberProtocol "github.com/gojek/fiber/protocol"
)

func TestCopyResponseToLogChannel(t *testing.T) {
	resp := &upiv1.PredictValuesResponse{}
	key := "test"

	tests := []struct {
		name     string
		err      *errors.TuringError
		expected grpcRouterResponse
	}{
		{
			name: "ok",
			expected: grpcRouterResponse{
				key:    key,
				body:   resp,
				header: metadata.New(map[string]string{"test": "key"}),
			},
		},
		{
			name: "error",
			err:  errors.NewTuringError(fmt.Errorf("test error"), fiberProtocol.GRPC),
			expected: grpcRouterResponse{
				key:     key,
				err:     "test error",
				errCode: 13,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			md := metadata.Pairs("test", "key")
			// Make response channel
			respCh := make(chan grpcRouterResponse, 1)
			copyResponseToLogChannel(respCh, key, md, resp, tt.err)

			close(respCh)
			data := <-respCh
			require.Equal(t, tt.expected, data)
		})
	}
}

// TestLogTuringRouterRequestSummary tests that the response and request are parsed into the expected format
// this include nested proto where underscore separated is expected e.g. prediction_result_table
func TestLogTuringRouterRequestSummary(t *testing.T) {
	table := &upiv1.Table{
		Name: "abc",
		Columns: []*upiv1.Column{
			{
				Name: "col1",
				Type: upiv1.Type_TYPE_DOUBLE,
			},
		},
		Rows: []*upiv1.Row{
			{
				RowId: "row1",
				Values: []*upiv1.Value{
					{
						DoubleValue: 123.456,
					},
				},
			},
		},
	}

	tests := []struct {
		name              string
		reqHeader         metadata.MD
		reqBody           *upiv1.PredictValuesRequest
		resHeader         metadata.MD
		resBody           *upiv1.PredictValuesResponse
		expectedReqBody   string
		expectedReqHeader string
		expectedResHeader string
		expectedResBody   string
	}{
		{
			name: "ok",
			reqHeader: metadata.New(map[string]string{
				"key1": "value1",
				"key2": "value2",
			}),
			reqBody: &upiv1.PredictValuesRequest{
				PredictionTable: table,
			},
			resHeader: metadata.New(map[string]string{
				"key3": "value3",
			}),
			resBody: &upiv1.PredictValuesResponse{
				PredictionResultTable: table,
			},
			expectedReqHeader: "map[key1:value1 key2:value2]",
			expectedResHeader: "map[key3:value3]",
			expectedReqBody: `{"prediction_table":{"name":"abc", "rows":[{"row_id":"row1", 
								"values":[{"double_value":123.456}]}]}}`,
			expectedResBody: `{"prediction_result_table":{"name":"abc", "rows":[{"row_id":"row1", 
								"values":[{"double_value":123.456}]}]}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//Set up observer for test to log entry to console
			core, collectedLogs := observer.New(zap.InfoLevel)
			logger := zap.New(core)
			log.SetGlobalLogger(logger.Sugar())
			err := resultlog.InitTuringResultLogger(&config.AppConfig{
				ResultLogger: config.UPILogger,
			})
			assert.NoError(t, err)

			//Add response to channel and invoke logging of summary
			respCh := make(chan grpcRouterResponse, 1)
			respCh <- grpcRouterResponse{header: tt.resHeader, body: tt.resBody, key: resultlog.ResultLogKeys.Router}
			close(respCh)
			logTuringRouterRequestSummary(tt.reqHeader, tt.reqBody, respCh)

			filteredLogs := collectedLogs.FilterMessage("Turing Request Summary")
			require.NotZero(t, filteredLogs.Len())
			for _, log := range filteredLogs.All() {
				for _, content := range log.Context {
					if content.Key == "request" {
						req := content.Interface.(map[string]interface{})
						header := fmt.Sprintf("%v", req["header"])
						body := fmt.Sprintf("%v", req["body"])
						assert.Equal(t, tt.expectedReqHeader, header)
						assert.JSONEq(t, tt.expectedReqBody, body)
					} else if content.Key == "router" {
						req := content.Interface.(map[string]interface{})
						header := fmt.Sprintf("%v", req["header"])
						body := fmt.Sprintf("%v", req["response"])
						assert.Equal(t, tt.expectedResHeader, header)
						assert.JSONEq(t, tt.expectedResBody, body)
					}
				}
			}
		})
	}
}
