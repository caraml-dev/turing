package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	mchttp "github.com/caraml-dev/turing/engines/router/missionctl/server/http"
	"github.com/caraml-dev/turing/engines/router/missionctl/server/http/handlers/compression"
)

type routerResponse struct {
	key    string
	header http.Header
	body   []byte
	err    string
}

// logTuringRouterRequestSummary logs the summary of the request made to the turing router,
// through the configured result logger. It takes as its input the turing request id, the
// request header and body for the original request to the turing router, a response channel
// with responses from each stage of the turing workflow.
func logTuringRouterRequestSummary(
	ctx context.Context,
	logger log.Logger,
	timestamp time.Time,
	reqHeader http.Header,
	reqBody []byte,
	mcRespCh <-chan routerResponse,
) {
	logger.Debugw("logTuringRouterRequestSummary", "reqBody", string(reqBody))
	// Uncompress request data
	uncompressedData, err := uncompressHTTPBody(reqHeader, reqBody)
	if err != nil {
		logger.Errorf("Error occurred when reading request body: %s", err.Error())
	}

	// Create a new TuringResultLogEntry record with the context and request info
	logEntry := resultlog.NewTuringResultLogEntry(ctx, timestamp, reqHeader, string(uncompressedData))

	// Read incoming responses and prepare for logging
	for resp := range mcRespCh {
		// If error exists, add an error record
		if resp.err != "" {
			logEntry.AddResponse(resp.key, "", nil, resp.err)
		} else {
			// Process the response body
			uncompressedData, err := uncompressHTTPBody(resp.header, resp.body)
			if err != nil {
				logger.Errorf("Error occurred when reading %s response body: %s",
					resp.key, err.Error())
				logEntry.AddResponse(resp.key, "", nil, err.Error())
			} else {
				// Format the response header
				responseHeader := resultlog.FormatHeader(resp.header)
				logEntry.AddResponse(resp.key, string(uncompressedData), responseHeader, "")
			}
		}
	}

	// Log the responses. If an error occurs in logging the result to the
	// configured result log destination, log the error.
	if err = resultlog.LogEntry(logEntry); err != nil {
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
	ctx context.Context,
	ch chan<- routerResponse,
	key string,
	r mchttp.Response,
	httpErr *errors.TuringError,
) {
	var data []byte

	// if http error is not nil, use error as response
	if httpErr != nil {
		ch <- routerResponse{
			key: key,
			err: httpErr.Message,
		}
		return
	}

	data = r.Body()
	if data == nil {
		// Error in logging method, doesn't have to be propagated. Simply log the error.
		logger := log.WithContext(ctx)
		defer func() {
			_ = logger.Sync()
		}()
		logger.Errorf("Error occurred when reading data from %s", key)
	}
	// Copy to channel
	ch <- routerResponse{
		key:    key,
		header: r.Header(),
		body:   data,
	}
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
