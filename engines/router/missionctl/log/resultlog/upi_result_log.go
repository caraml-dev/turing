package resultlog

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/caraml-dev/universal-prediction-interface/pkg/converter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

// UPIResultLogger holds the logic how the RouterLog is being constructed,
// the server will always provide GrpcRouterResponse and UPIResultLogger
// will be responsible to construct it into TuringResultLogMessage or
// RouterLog and writes to the destination with the underlying logger middleware
type UPIResultLogger struct {
	// loggerType is the ResultLoggerType configured in env var
	loggerType config.ResultLogger
	// UPILogger holds the detail of how the RouterLog is being written to the configured sink,
	// currently expected to work with KafkaLogger only
	upiLogger UPILogger
	// turingResultLogger is the ResultLogger for TuringResultLogMessage proto
	turingResultLogger *ResultLogger
	// This corresponds to the name and version of the router deployed from the Turing app
	// Format: {router_name}-{router_version}.{project_name}
	routerName    string
	routerVersion string
	projectName   string
}

// GrpcRouterResponse is sent to the result logger to construct RouterLog or TuringResultLog
type GrpcRouterResponse struct {
	// Key field is not used for now as there is only router,
	// it can be used to differentiate enricher/ensembler/experiment engine response
	Key     string
	Header  metadata.MD
	Body    *upiv1.PredictValuesResponse
	Err     string
	ErrCode int
}

type UPILogger interface {
	write(routerLog *upiv1.RouterLog) error
}

var loggingErrorTemplate = "logging error. unable to convert table to struct for %s : %s"

var convertorTableSchema uint32 = converter.TableSchemaV1

var routerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+-\d+.[a-zA-Z0-9\-]+$`)

func InitUPIResultLogger(
	appName string,
	loggerType config.ResultLogger,
	upiLogger UPILogger,
	resultLogger *ResultLogger) (*UPIResultLogger, error) {

	upiResultLogger := &UPIResultLogger{
		upiLogger:          upiLogger,
		turingResultLogger: resultLogger,
		loggerType:         loggerType,
	}
	if !routerNameRegex.MatchString(appName) {
		return nil, fmt.Errorf("invalid router name")
	}
	s := strings.Split(appName, ".")
	routerNameWithVersion := s[0]
	upiResultLogger.projectName = s[1]

	i := strings.LastIndex(routerNameWithVersion, "-")
	upiResultLogger.routerName = routerNameWithVersion[:i]
	// do not include '-'
	upiResultLogger.routerVersion = routerNameWithVersion[i+1:]

	return upiResultLogger, nil
}

func (ul *UPIResultLogger) LogTuringRouterRequestSummary(
	header metadata.MD,
	upiReq *upiv1.PredictValuesRequest,
	mcRespCh <-chan GrpcRouterResponse,
) {
	if ul.loggerType == config.UPILogger {
		logRouterLog(header, upiReq, mcRespCh, ul)
	} else {
		logTuringResultLog(header, upiReq, mcRespCh, ul)
	}
}

func (ul *UPIResultLogger) logEntry(log *upiv1.RouterLog) error {
	return ul.upiLogger.write(log)
}

// LogTuringRouterRequestError logs the given turing request id and the error data
func (ul *UPIResultLogger) LogTuringRouterRequestError(ctx context.Context, err *errors.TuringError) {
	logger := log.WithContext(ctx)
	logger.Errorw("Turing Request Error",
		"error", err.Message,
		"status", err.Code,
	)
}

// SendResponseToLogChannel send the response from the turing router to the given channel
// as a RouterResponse object
func (ul *UPIResultLogger) SendResponseToLogChannel(
	ch chan<- GrpcRouterResponse,
	key string,
	md metadata.MD,
	r *upiv1.PredictValuesResponse,
	err *errors.TuringError) {

	if err != nil {
		ch <- GrpcRouterResponse{
			Key:     key,
			Header:  md,
			Err:     err.Message,
			ErrCode: err.Code,
		}
		return
	}

	ch <- GrpcRouterResponse{
		Key:    key,
		Header: md,
		Body:   r,
	}
}

func logTuringResultLog(
	header metadata.MD,
	req *upiv1.PredictValuesRequest,
	mcRespCh <-chan GrpcRouterResponse,
	ul *UPIResultLogger) {

	// Create a new TuringResultLogEntry record with the context and request info
	// send proto as json string
	reqStr := protoJSONMarshaller.Format(req)
	logEntry := NewTuringResultLog(
		req.GetMetadata().GetPredictionId(),
		req.GetMetadata().GetRequestTimestamp().AsTime(),
		header,
		reqStr)

	// Read incoming responses and prepare for logging
	for resp := range mcRespCh {
		// If error exists, add an error record
		if resp.Err != "" {
			AddResponse(logEntry, resp.Key, "", nil, resp.Err)
		} else {
			upiResp := protoJSONMarshaller.Format(resp.Body)
			AddResponse(logEntry, resp.Key, upiResp, FormatHeader(resp.Header), "")
		}
	}

	// Log the responses. If an error occurs in logging the result to the
	// configured result log destination, log the error.
	if err := ul.turingResultLogger.logEntry(logEntry); err != nil {
		log.Glob().Errorf("Result Logging Error: %s", err.Error())
	}
}

func logRouterLog(header metadata.MD,
	upiReq *upiv1.PredictValuesRequest,
	mcRespCh <-chan GrpcRouterResponse,
	ul *UPIResultLogger) {
	grpcResp := <-mcRespCh
	upiResp := grpcResp.Body

	var predictionTable *structpb.Struct
	var err error
	if upiReq.GetPredictionTable() != nil {
		predictionTable, err = converter.TableToStruct(upiReq.GetPredictionTable(), convertorTableSchema)
		if err != nil {
			log.Glob().Errorf(loggingErrorTemplate, "input prediction_table", err.Error())
			return
		}
	}
	transformerTable, err := convertTransformerTable(upiReq.GetTransformerInput().GetTables())
	if err != nil {
		log.Glob().Errorf(loggingErrorTemplate, "input transformer_table", err.Error())
		return
	}

	// traffic-rule is returned in fiber response with `traffic-rule` key
	routerLog := &upiv1.RouterLog{
		PredictionId:  upiReq.GetMetadata().GetPredictionId(),
		TargetName:    upiReq.GetTargetName(),
		RouterName:    ul.routerName,
		RouterVersion: ul.routerVersion,
		ProjectName:   ul.projectName,
		RoutingLogic: &upiv1.RoutingLogic{
			Models:         upiResp.GetMetadata().GetModels(),
			TrafficRule:    strings.Join(grpcResp.Header.Get("traffic-rule"), ""),
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
			Headers:           convertToUPIHeader(grpcResp.Header),
		},
		RequestTimestamp:   upiReq.GetMetadata().GetRequestTimestamp(),
		TableSchemaVersion: convertorTableSchema,
	}

	if grpcResp.ErrCode != int(codes.OK) {
		routerLog.RouterOutput.Message = grpcResp.Err
		routerLog.RouterOutput.Status = uint32(grpcResp.ErrCode)
	} else {
		var predictionResultTable *structpb.Struct
		if upiResp.GetPredictionResultTable() != nil {
			predictionResultTable, err = converter.TableToStruct(upiResp.GetPredictionResultTable(), converter.TableSchemaV1)
			if err != nil {
				log.Glob().Errorf(loggingErrorTemplate, "output prediction_table", err.Error())
				return
			}
		}
		routerLog.RouterOutput.Status = 0
		routerLog.RouterOutput.PredictionResultsTable = predictionResultTable
	}

	// Log the responses. If an error occurs in logging the result to the
	// configured result log destination, log the error.
	if err := ul.logEntry(routerLog); err != nil {
		log.Glob().Errorf("Result Logging Error: %s, Router log: %s", err.Error(), routerLog.String())
	}
}

// consider moving these function out if there are other places of usage
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
