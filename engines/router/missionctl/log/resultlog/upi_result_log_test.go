package resultlog

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog/proto/turing"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/caraml-dev/universal-prediction-interface/pkg/converter"

	fiberProtocol "github.com/gojek/fiber/protocol"
)

func Test_InitUPIResultLogger(t *testing.T) {

	type args struct {
		appName string
	}
	tests := []struct {
		name   string
		args   args
		want   *UPIResultLogger
		errMsg string
	}{
		{
			name: "valid app name",
			args: args{
				appName: "a-b-c-1111.my-project",
			},
		},
		{
			name: "invalid app name",
			args: args{
				appName: "abc-asdaf-1",
			},
			errMsg: "invalid router name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := InitUPIResultLogger(tt.args.appName, "", nil, nil)
			if tt.errMsg != "" {
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUPIResultLogger_CopyResponseToLogChannel(t *testing.T) {
	type args struct {
		key string
		md  metadata.MD
		r   *upiv1.PredictValuesResponse
		err *errors.TuringError
	}
	tests := []struct {
		name string
		args args
		want GrpcRouterResponse
	}{
		{
			name: "log valid response",
			args: args{
				key: "abc",
				md:  metadata.Pairs("key1", "value1", "key2", "value2"),
				r: &upiv1.PredictValuesResponse{
					PredictionResultTable: &upiv1.Table{
						Name: "table",
					},
				},
				err: nil,
			},
			want: GrpcRouterResponse{
				Key:    "abc",
				Header: metadata.New(map[string]string{"key1": "value1", "key2": "value2"}),
				Body: &upiv1.PredictValuesResponse{
					PredictionResultTable: &upiv1.Table{
						Name: "table",
					},
				},
				Err:     "",
				ErrCode: 0,
			},
		},
		{
			name: "log error response",
			args: args{
				key: "efg",
				md:  metadata.Pairs("key1", "value1"),
				err: errors.NewTuringError(fmt.Errorf("fail request"), fiberProtocol.GRPC),
			},
			want: GrpcRouterResponse{
				Key:     "efg",
				Header:  metadata.New(map[string]string{"key1": "value1"}),
				Err:     "fail request",
				ErrCode: 13,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			respCh := make(chan GrpcRouterResponse, 1)
			ul := &UPIResultLogger{}

			ul.SendResponseToLogChannel(respCh, tt.args.key, tt.args.md, tt.args.r, tt.args.err)
			close(respCh)
			data := <-respCh
			assert.Equal(t, tt.want, data)
		})
	}
}

type mockUPILogger struct {
	numOfCalls int32
	routerLog  *upiv1.RouterLog
}

func (l *mockUPILogger) write(log *upiv1.RouterLog) error {
	l.routerLog = log
	atomic.AddInt32(&l.numOfCalls, 1)
	return nil
}

// mockUPILogger is injected to the upiLogger, so that the mapping logic of this function can be
// tested and verified
func TestUPIResultLogger_LogTuringRouterRequestSummary_logRouterLog(t *testing.T) {

	// Variables are predefined, as the test validate the mapping of the responding
	// input to fields of Router Log and the data should not be mutated
	routerName := "router-name"
	routerVersion := "router-version"
	projectName := "project-name"

	time := timestamppb.Now()
	predictionContext := []*upiv1.Variable{
		{
			Name:        "var 1",
			Type:        upiv1.Type_TYPE_DOUBLE,
			DoubleValue: 12.34,
		},
	}
	predictionTable := &upiv1.Table{
		Name: "t1",
		Columns: []*upiv1.Column{
			{
				Name: "col1",
				Type: upiv1.Type_TYPE_DOUBLE,
			},
		},
		Rows: []*upiv1.Row{
			{
				RowId: "1",
				Values: []*upiv1.Value{
					{
						DoubleValue: 12.34,
					},
				},
			},
		},
	}
	predictionTableStruct, err := converter.TableToStruct(predictionTable, convertorTableSchema)
	require.NoError(t, err)
	modelMedata := []*upiv1.ModelMetadata{
		{
			Name:    "model3",
			Version: "3",
		},
	}

	variable := &upiv1.Variable{
		Name:        "var1",
		Type:        upiv1.Type_TYPE_STRING,
		StringValue: "var1 value",
	}

	type args struct {
		reqHeader    metadata.MD
		upiReq       *upiv1.PredictValuesRequest
		routerResp   GrpcRouterResponse
		resultLogger *UPIResultLogger
	}
	tests := []struct {
		name string
		args args
		want *upiv1.RouterLog
	}{
		{
			name: "empty request and response",
			args: args{
				resultLogger: &UPIResultLogger{
					upiLogger:     &mockUPILogger{},
					loggerType:    config.UPILogger,
					routerName:    routerName,
					routerVersion: routerVersion,
					projectName:   projectName,
				},
			},
			want: &upiv1.RouterLog{
				ProjectName:        projectName,
				RouterName:         routerName,
				RouterVersion:      routerVersion,
				TableSchemaVersion: convertorTableSchema,
				RoutingLogic:       &upiv1.RoutingLogic{},
				RouterInput:        &upiv1.RouterInput{},
				RouterOutput:       &upiv1.RouterOutput{},
			},
		},
		{
			name: "predict request only",
			args: args{
				reqHeader: metadata.Pairs("k1", "v1"),
				upiReq: &upiv1.PredictValuesRequest{
					PredictionTable: predictionTable,
					TransformerInput: &upiv1.TransformerInput{
						Tables:    []*upiv1.Table{predictionTable},
						Variables: []*upiv1.Variable{variable},
					},
					TargetName:        "target-name",
					PredictionContext: predictionContext,
					Metadata: &upiv1.RequestMetadata{
						PredictionId:     "123",
						RequestTimestamp: time,
					},
				},
				resultLogger: &UPIResultLogger{upiLogger: &mockUPILogger{}, loggerType: config.UPILogger},
			},
			want: &upiv1.RouterLog{
				PredictionId: "123",
				TargetName:   "target-name",
				RoutingLogic: &upiv1.RoutingLogic{},
				RouterInput: &upiv1.RouterInput{
					PredictionTable:      predictionTableStruct,
					TransformerTables:    []*structpb.Struct{predictionTableStruct},
					TransformerVariables: []*upiv1.Variable{variable},
					PredictionContext:    predictionContext,
					Headers: []*upiv1.Header{
						{
							Key:   "k1",
							Value: "v1",
						},
					},
				},
				RouterOutput:       &upiv1.RouterOutput{},
				RequestTimestamp:   time,
				TableSchemaVersion: convertorTableSchema,
			},
		},
		{
			name: "predict request only without err",
			args: args{
				resultLogger: &UPIResultLogger{upiLogger: &mockUPILogger{}, loggerType: config.UPILogger},
				routerResp: GrpcRouterResponse{
					Header: metadata.Pairs("traffic-rule", "rule3"),
					Body: &upiv1.PredictValuesResponse{
						PredictionResultTable: predictionTable,
						PredictionContext:     predictionContext,
						Metadata: &upiv1.ResponseMetadata{
							PredictionId:   "",
							Models:         modelMedata,
							ExperimentName: "experiment",
							TreatmentName:  "treatment",
						},
					},
				},
			},
			want: &upiv1.RouterLog{
				TableSchemaVersion: convertorTableSchema,
				RoutingLogic: &upiv1.RoutingLogic{
					Models:         modelMedata,
					TrafficRule:    "rule3",
					ExperimentName: "experiment",
					TreatmentName:  "treatment",
				},
				RouterInput: &upiv1.RouterInput{},
				RouterOutput: &upiv1.RouterOutput{
					PredictionResultsTable: predictionTableStruct,
					PredictionContext:      predictionContext,
					Headers: []*upiv1.Header{
						{
							Key:   "traffic-rule",
							Value: "rule3",
						},
					},
					Status: 0,
				},
			},
		},
		{
			name: "predict request only with err",
			args: args{
				resultLogger: &UPIResultLogger{upiLogger: &mockUPILogger{}, loggerType: config.UPILogger},
				routerResp: GrpcRouterResponse{
					Err:     "no response from model",
					ErrCode: 13,
				},
			},
			want: &upiv1.RouterLog{
				TableSchemaVersion: convertorTableSchema,
				RoutingLogic:       &upiv1.RoutingLogic{},
				RouterInput:        &upiv1.RouterInput{},
				RouterOutput: &upiv1.RouterOutput{
					Status:  13,
					Message: "no response from model",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			respCh := make(chan GrpcRouterResponse, 1)
			respCh <- tt.args.routerResp
			close(respCh)
			tt.args.resultLogger.LogTuringRouterRequestSummary(tt.args.reqHeader, tt.args.upiReq, respCh)
			mockLogger, ok := tt.args.resultLogger.upiLogger.(*mockUPILogger)
			assert.True(t, ok, "mockUPILogger not used")
			assert.Equal(t, tt.want, mockLogger.routerLog)
		})
	}
}

func TestUPIResultLogger_LogTuringRouterRequestSummary_logTuringResultLog(t *testing.T) {

	appName := "appName"
	testTime := time.Now()
	predictionID := "123"
	table := &upiv1.Table{
		Name: "abc",
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

	upiReq := &upiv1.PredictValuesRequest{
		PredictionTable: table,
		Metadata: &upiv1.RequestMetadata{
			PredictionId:     predictionID,
			RequestTimestamp: timestamppb.New(testTime),
		},
	}
	upiResponse := &upiv1.PredictValuesResponse{
		PredictionResultTable: table,
	}

	type args struct {
		reqHeader    metadata.MD
		upiReq       *upiv1.PredictValuesRequest
		routerResp   GrpcRouterResponse
		resultLogger *UPIResultLogger
	}
	tests := []struct {
		name string
		args args
		want *turing.TuringResultLogMessage
	}{
		{
			name: "ok request response",
			args: args{
				reqHeader: metadata.Pairs("req", "header"),
				upiReq:    upiReq,
				routerResp: GrpcRouterResponse{
					Key:    ResultLogKeys.Router,
					Header: metadata.Pairs("res", "header", "res2", "header2"),
					Body:   upiResponse,
				},
				resultLogger: &UPIResultLogger{
					turingResultLogger: &ResultLogger{
						trl:     &mockResultLogger{},
						appName: appName,
					},
					loggerType: config.BigqueryLogger,
				},
			},
			want: &turing.TuringResultLogMessage{
				TuringReqId:    "123",
				EventTimestamp: timestamppb.New(testTime),
				RouterVersion:  appName,
				Request: &turing.Request{
					Header: map[string]string{"req": "header"},
					Body:   protoJSONMarshaller.Format(upiReq),
				},
				Router: &turing.Response{
					Header:   map[string]string{"res": "header", "res2": "header2"},
					Response: protoJSONMarshaller.Format(upiResponse),
				},
			},
		},
		{
			name: "error resp",
			args: args{
				reqHeader: metadata.Pairs("req", "header"),
				upiReq:    upiReq,
				routerResp: GrpcRouterResponse{
					Key:     ResultLogKeys.Router,
					Err:     "fail fail",
					ErrCode: 13,
				},
				resultLogger: &UPIResultLogger{
					turingResultLogger: &ResultLogger{
						trl:     &mockResultLogger{},
						appName: appName,
					},
					loggerType: config.BigqueryLogger,
				},
			},
			want: &turing.TuringResultLogMessage{
				TuringReqId:    "123",
				EventTimestamp: timestamppb.New(testTime),
				RouterVersion:  appName,
				Request: &turing.Request{
					Header: map[string]string{"req": "header"},
					Body:   protoJSONMarshaller.Format(upiReq),
				},
				Router: &turing.Response{
					Error: "fail fail",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			respCh := make(chan GrpcRouterResponse, 1)
			respCh <- tt.args.routerResp
			close(respCh)
			tt.args.resultLogger.LogTuringRouterRequestSummary(tt.args.reqHeader, tt.args.upiReq, respCh)
			mockLogger, ok := tt.args.resultLogger.turingResultLogger.trl.(*mockResultLogger)
			assert.True(t, ok, "mockResultLogger not used")
			assert.Equal(t, tt.want, mockLogger.result)
		})
	}
}
