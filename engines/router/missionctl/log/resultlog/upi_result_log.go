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
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

// UPIResultLogger holds the logic how the RouterLog is being constructed,
type UPIResultLogger struct {
	// UPILogger holds the detail of how the RouterLog is being written to the configured sink,
	// currently expected to work with KafkaLogger only
	logger UPILogger
	// This corresponds to the name and version of the router deployed from the Turing app
	// Format: {router_name}-{router_version}.{project_name}
	routerName    string
	routerVersion string
	projectName   string
}

// GrpcRouterResponse is sent to the result logger to construct RouterLog
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
	writeUPIRouterLog(routerLog *upiv1.RouterLog) error
}

var loggingErrorTemplate = "logging error. unable to convert table to struct for %s : %s"

var convertorTableSchema uint32 = converter.TableSchemaV1

var routerNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-]+-\d+.[a-zA-Z0-9\-]+$`)

func InitUPIResultLogger(appName string, kafkaCfg *config.KafkaConfig) (*UPIResultLogger, error) {

	resultLogger := &UPIResultLogger{}
	if !routerNameRegex.MatchString(appName) {
		return nil, fmt.Errorf("invalid router name")
	}
	s := strings.Split(appName, ".")
	routerNameWithVersion := s[0]
	resultLogger.projectName = s[1]
	i := strings.LastIndex(routerNameWithVersion, "-")
	resultLogger.routerName = routerNameWithVersion[:i]
	// do not include '-'
	resultLogger.routerVersion = routerNameWithVersion[i+1:]

	kafkaLogger, err := newKafkaLogger(kafkaCfg)
	if err != nil {
		return nil, err
	}
	return &UPIResultLogger{
		logger: kafkaLogger,
	}, nil
}

func (ul *UPIResultLogger) LogEntry(routerLog *upiv1.RouterLog) error {
	return ul.logger.writeUPIRouterLog(routerLog)
}

func (ul *UPIResultLogger) LogTuringRouterRequestSummary(
	header metadata.MD,
	upiReq *upiv1.PredictValuesRequest,
	mcRespCh <-chan GrpcRouterResponse,
) {

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

	if grpcResp.Err != "" {
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
	if err := ul.LogEntry(routerLog); err != nil {
		log.Glob().Errorf("Result Logging Error: %s", err.Error())
	}
}

// LogTuringRouterRequestError logs the given turing request id and the error data
func (ul *UPIResultLogger) LogTuringRouterRequestError(ctx context.Context, err *errors.TuringError) {
	logger := log.WithContext(ctx)
	defer func() {
		_ = logger.Sync()
	}()
	logger.Errorw("Turing Request Error",
		"error", err.Message,
		"status", err.Code,
	)
}

// CopyResponseToLogChannel copies the response from the turing router to the given channel
// as a routerResponse object
func (ul *UPIResultLogger) CopyResponseToLogChannel(
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
