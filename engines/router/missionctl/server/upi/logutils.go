package upi

import (
	"context"
	"strings"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/caraml-dev/universal-prediction-interface/pkg/converter"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

type grpcRouterResponse struct {
	key     string
	header  metadata.MD
	body    *upiv1.PredictValuesResponse
	err     string
	errCode int
}

var loggingErrorTemplate = "logging error. unable to convert table to struct for %s : %s"

var convertorTableSchema uint32 = converter.TableSchemaV1

func logTuringRouterRequestSummary(
	header metadata.MD,
	upiReq *upiv1.PredictValuesRequest,
	mcRespCh <-chan grpcRouterResponse,
) {

	grpcResp := <-mcRespCh
	upiResp := grpcResp.body

	predictionTable, err := converter.TableToStruct(upiReq.GetPredictionTable(), convertorTableSchema)
	if err != nil {
		log.Glob().Errorf(loggingErrorTemplate, "input prediction_table", err.Error())
		return
	}
	transformerTable, err := convertTransformerTable(upiReq.GetTransformerInput().GetTables())
	if err != nil {
		log.Glob().Errorf(loggingErrorTemplate, "input transformer_table", err.Error())
		return
	}

	// traffic-rule is returned in fiber response with `traffic-rule` key
	routerLog := &upiv1.RouterLog{
		PredictionId: upiReq.GetMetadata().GetPredictionId(),
		TargetName:   upiReq.GetTargetName(),
		RoutingLogic: &upiv1.RoutingLogic{
			Models:         upiResp.GetMetadata().GetModels(),
			TrafficRule:    strings.Join(grpcResp.header.Get("traffic-rule"), ""), //TODO
			ExperimentName: upiResp.GetMetadata().GetExperimentName(),
			TreatmentName:  upiResp.GetMetadata().GetTreatmentName(),
		},
		RouterInput: &upiv1.RouterInput{
			PredictionTable:      predictionTable,
			TransformerTables:    transformerTable,
			TransformerVariables: upiReq.GetTransformerInput().GetVariables(),
			PredictionContext:    upiReq.GetPredictionContext(),
			Headers:              convertToUPIHeader(header),
		},
		RouterOutput: &upiv1.RouterOutput{
			PredictionContext: upiResp.GetPredictionContext(),
			Headers:           convertToUPIHeader(grpcResp.header),
		},
		RequestTimestamp:   upiReq.GetMetadata().GetRequestTimestamp(),
		TableSchemaVersion: 1,
	}

	if grpcResp.err != "" {
		routerLog.RouterOutput.Message = grpcResp.err
		routerLog.RouterOutput.Status = uint32(grpcResp.errCode)
	} else {
		predictionTable, err := converter.TableToStruct(upiReq.GetPredictionTable(), converter.TableSchemaV1)
		if err != nil {
			log.Glob().Errorf(loggingErrorTemplate, "output prediction_table", err.Error())
			return
		}
		routerLog.RouterOutput.Status = 0
		routerLog.RouterOutput.PredictionResultsTable = predictionTable
	}

	// Log the responses. If an error occurs in logging the result to the
	// configured result log destination, log the error.
	if err := resultlog.LogUPIEntry(routerLog); err != nil {
		log.Glob().Errorf("Result Logging Error: %s", err.Error())
	}
}

// logTuringRouterRequestError logs the given turing request id and the error data
func logTuringRouterRequestError(ctx context.Context, err *errors.TuringError) {
	logger := log.WithContext(ctx)
	defer func() {
		_ = logger.Sync()
	}()
	logger.Errorw("Turing Request Error",
		"error", err.Message,
		"status", err.Code,
	)
}

// copyResponseToLogChannel copies the response from the turing router to the given channel
// as a routerResponse object
func copyResponseToLogChannel(
	ch chan<- grpcRouterResponse,
	key string,
	md metadata.MD,
	r *upiv1.PredictValuesResponse,
	err *errors.TuringError) {
	// if error is not nil, use error as response
	if err != nil {
		ch <- grpcRouterResponse{
			key:     key,
			err:     err.Message,
			errCode: err.Code,
		}
		return
	}

	// Copy to channel
	ch <- grpcRouterResponse{
		key:    key,
		header: md,
		body:   r,
	}
}

func convertToUPIHeader(md metadata.MD) []*upiv1.Header {
	var headers []*upiv1.Header
	for key, val := range md {
		headers = append(headers, &upiv1.Header{
			Key:   key,
			Value: strings.Join(val, ","),
		})
	}
	return headers
}

func convertTransformerTable(tables []*upiv1.Table) ([]*structpb.Struct, error) {
	var t []*structpb.Struct
	for _, table := range tables {
		s, err := converter.TableToStruct(table, convertorTableSchema)
		if err != nil {
			return nil, err
		}
		t = append(t, s)
	}
	return t, nil
}
