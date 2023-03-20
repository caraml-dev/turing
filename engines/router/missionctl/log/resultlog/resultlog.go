package resultlog

import (
	"context"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog/proto/turing"
	mchttp "github.com/caraml-dev/turing/engines/router/missionctl/server/http"
	"github.com/caraml-dev/turing/engines/router/missionctl/server/http/handlers/compression"
)

// ResultLogger holds the logic how the TuringResultLogMessage is being constructed,
// and writes to the destination using the logger middleware
type ResultLogger struct {
	trl TuringResultLogger
	// appName stores the configured app name, to be applied to each log entry
	// This corresponds to the name and version of the router deployed from the Turing app and
	// will be logged as RouterVersion in TuringResultLog.proto
	// Format: {router_name}-{router_version}.{project_name}
	appName string
}

// TuringResultLogger is an abstraction for the underlying result logger for TuringResultLogMessage
type TuringResultLogger interface {
	write(message *turing.TuringResultLogMessage) error
}

// RouterResponse is the struct of expected to pass into response channel to be logged as TuringResultLogMessage later
type RouterResponse struct {
	key    string
	header http.Header
	body   []byte
	err    string
}

// ResultLogKeys defines the individual components for which the result log must be created
var ResultLogKeys = struct {
	Experiment string
	Enricher   string
	Router     string
	Ensembler  string
}{
	Experiment: "experiment",
	Enricher:   "enricher",
	Router:     "router",
	Ensembler:  "ensembler",
}

var protoJSONMarshaller = protojson.MarshalOptions{UseProtoNames: true}

// LogTuringRouterRequestSummary logs the summary of the request made to the turing router,
// through the configured result logger. It takes as its input the turing request id, the
// request header and body for the original request to the turing router, a response channel
// with responses from each stage of the turing workflow.
func (rl *ResultLogger) LogTuringRouterRequestSummary(
	predictionID string,
	logger log.Logger,
	timestamp time.Time,
	reqHeader http.Header,
	reqBody []byte,
	mcRespCh <-chan RouterResponse) {

	logger.Debugw("Logging request", "reqBody", string(reqBody))
	// Uncompress request data
	uncompressedData, err := uncompressHTTPBody(reqHeader, reqBody)
	if err != nil {
		logger.Errorf("Error occurred when reading request body: %s", err.Error())
	}

	// Create a new TuringResultLogEntry record with the context and request info
	logEntry := NewTuringResultLog(predictionID, timestamp, reqHeader, string(uncompressedData))

	// Read incoming responses and prepare for logging
	for resp := range mcRespCh {
		logger.Debugw("Received data in response channel")
		// If error exists, add an error record
		if resp.err != "" {
			AddResponse(logEntry, resp.key, "", nil, resp.err)
		} else {
			// Process the response body
			uncompressedData, err := uncompressHTTPBody(resp.header, resp.body)
			if err != nil {
				logger.Errorf("Error occurred when reading %s response body: %s",
					resp.key, err.Error())
				AddResponse(logEntry, resp.key, "", nil, err.Error())
			} else {
				logger.Debugw("Logging response", "respBody", string(uncompressedData))
				// Format the response header
				responseHeader := FormatHeader(resp.header)
				AddResponse(logEntry, resp.key, string(uncompressedData), responseHeader, "")
			}
		}
	}

	logger.Debugw("Received all response from mcRespCh")
	// Log the responses. If an error occurs in logging the result to the
	// configured result log destination, log the error.
	if err = rl.logEntry(logEntry); err != nil {
		logger.Errorf("Result Logging Error: %s", err.Error())
	}
}

// LogTuringRouterRequestError logs the given turing request id and the error data
func (rl *ResultLogger) LogTuringRouterRequestError(ctx context.Context, err *errors.TuringError) {
	logger := log.WithContext(ctx)
	logger.Errorw("Turing Request Error",
		"error", err.Message,
		"status", err.Code,
	)
}

// SendResponseToLogChannel copies the response from the turing router to the given channel
// as a RouterResponse object
func (rl *ResultLogger) SendResponseToLogChannel(
	ctx context.Context,
	ch chan<- RouterResponse,
	key string,
	r mchttp.Response,
	httpErr *errors.TuringError,
) {
	var data []byte

	// if http error is not nil, use error as response
	if httpErr != nil {
		ch <- RouterResponse{
			key: key,
			err: httpErr.Message,
		}
		return
	}

	data = r.Body()
	if data == nil {
		// Error in logging method, doesn't have to be propagated. Simply log the error.
		logger := log.WithContext(ctx)
		logger.Errorf("Error occurred when reading data from %s", key)
	}
	// Copy to channel
	ch <- RouterResponse{
		key:    key,
		header: r.Header(),
		body:   data,
	}
}

func (rl *ResultLogger) logEntry(log *turing.TuringResultLogMessage) error {
	log.RouterVersion = rl.appName
	return rl.trl.write(log)
}

// NewTuringResultLog returns a new TuringResultLogMessage object with the given context
// and request
func NewTuringResultLog[h http.Header | metadata.MD](
	predictionID string,
	timestamp time.Time,
	header h,
	body string,
) *turing.TuringResultLogMessage {

	// Format Request Header
	reqHeader := FormatHeader(header)

	return &turing.TuringResultLogMessage{
		TuringReqId:    predictionID,
		EventTimestamp: timestamppb.New(timestamp),
		Request: &turing.Request{
			Header: reqHeader,
			Body:   body,
		},
	}
}

// InitTuringResultLogger initializes the result with supplied logger for
// logging TuringResultLogMessage. appName stores the configured app name,
// Format: {router_name}-{router_version}.{project_name}
func InitTuringResultLogger(appName string, logger TuringResultLogger) *ResultLogger {
	return &ResultLogger{
		trl:     logger,
		appName: appName,
	}
}

// FormatHeader formats the header which by concatenating the string values corresponding to each header into a
// single comma-delimited string
func FormatHeader[h http.Header | metadata.MD](header h) map[string]string {
	formattedHeader := map[string]string{}
	for k, v := range header {
		formattedHeader[k] = strings.Join(v, ",")
	}
	return formattedHeader
}

// uncompressHTTPBody uses the content encoding from the header and handles the
// uncompressing of request/response body accordingly
func uncompressHTTPBody(header http.Header, body []byte) ([]byte, error) {
	var result []byte

	if header == nil {
		return body, nil
	}

	switch header.Get("Content-Encoding") {
	case "lz4":
		lz := compression.LZ4Compressor{}
		return lz.Uncompress(body)
	default:
		// Use the input data as it is
		result = body
	}
	return result, nil
}

// AddResponse adds the per-component response/error info to the TuringResultLogEntry
func AddResponse(rl *turing.TuringResultLogMessage, key string, body string, header map[string]string, err string) {
	responseRecord := &turing.Response{
		Header:   header,
		Response: body,
		Error:    err,
	}
	switch key {
	case ResultLogKeys.Experiment:
		rl.Experiment = responseRecord
	case ResultLogKeys.Enricher:
		rl.Enricher = responseRecord
	case ResultLogKeys.Router:
		rl.Router = responseRecord
	case ResultLogKeys.Ensembler:
		rl.Ensembler = responseRecord
	}
}
