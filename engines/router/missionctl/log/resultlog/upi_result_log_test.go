package resultlog

import (
	"fmt"
	"testing"

	"bou.ke/monkey"
	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/caraml-dev/universal-prediction-interface/pkg/converter"
	fiberProtocol "github.com/gojek/fiber/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_InitUPIResultLogger(t *testing.T) {

	kafkaConfig := config.KafkaConfig{
		Brokers:             "localhost:9001",
		Topic:               "kafka_topic",
		SerializationFormat: config.ProtobufSerializationFormat,
	}
	mockProducer := &mockKafkaProducer{}
	monkey.Patch(newKafkaProducer, func(cfg *config.KafkaConfig) (kafkaProducer, error) {
		return mockProducer, nil
	})
	defer monkey.Unpatch(newKafkaProducer)
	// Set up GetMetadata on the mock producer
	mockProducer.On("GetMetadata", kafkaConfig.Topic, false, 1000).Return(nil, nil)

	type args struct {
		appName     string
		kafkaConfig *config.KafkaConfig
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
				appName:     "abc-asdaf-1",
				kafkaConfig: &kafkaConfig,
			},
			errMsg: "invalid router name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := InitUPIResultLogger(tt.args.appName, tt.args.kafkaConfig)
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

			ul.CopyResponseToLogChannel(respCh, tt.args.key, tt.args.md, tt.args.r, tt.args.err)
			close(respCh)
			data := <-respCh
			assert.Equal(t, tt.want, data)
		})
	}
}

type mockLogger struct{}

var logResult *upiv1.RouterLog

func (ml *mockLogger) writeUPIRouterLog(routerLog *upiv1.RouterLog) error {
	logResult = routerLog
	return nil
}

// mockLogger is injected to the logger, so that the mapping logic of this function can be
// tested and verified
func TestUPIResultLogger_LogTuringRouterRequestSummary(t *testing.T) {

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
					logger:        &mockLogger{},
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
				resultLogger: &UPIResultLogger{logger: &mockLogger{}},
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
				resultLogger: &UPIResultLogger{logger: &mockLogger{}},
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
				resultLogger: &UPIResultLogger{logger: &mockLogger{}},
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
			assert.Equal(t, tt.want, logResult)
		})
	}
}
