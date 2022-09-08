package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/caraml-dev/turing/engines/router/missionctl"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/experiment"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/log/resultlog"
	mchttp "github.com/caraml-dev/turing/engines/router/missionctl/server/http"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

const turingReqIDHeaderKey = "Turing-Req-ID"
const httpHandlerID = "http_handler"

// NewHTTPHandler creates an instance of the Mission Control's prediction request handler
func NewHTTPHandler(mc missionctl.MissionControl) http.Handler {
	return &httpHandler{mc}
}

// httpHandler is the Mission Control's prediction request handler
type httpHandler struct {
	missionctl.MissionControl
}

func (h *httpHandler) error(
	ctx context.Context,
	rw http.ResponseWriter,
	err *errors.TuringError,
) {
	// Get the turing request id from the context
	turingReqID, _ := turingctx.GetRequestID(ctx)
	// Add the turing request id to the error response header
	rw.Header().Set(turingReqIDHeaderKey, turingReqID)
	http.Error(rw, err.Message, err.Code)
	// Log the error
	logTuringRouterRequestError(ctx, err)
}

// enableTracingSpan associates span to context, if applicable
func (h *httpHandler) enableTracingSpan(ctx context.Context,
	req *http.Request,
	httpHandlerID string) (context.Context, opentracing.Span) {
	var sp opentracing.Span
	sp, ctx = tracing.Glob().StartSpanFromRequestHeader(ctx, httpHandlerID, req.Header)
	return ctx, sp
}

// getPrediction takes in a request and owns the flow of the request - enrich, route, ensemble
func (h *httpHandler) getPrediction(
	ctx context.Context,
	req *http.Request,
	ctxLogger *zap.SugaredLogger,
	requestBody []byte,
) (mchttp.Response, *errors.TuringError) {
	// Create response channel to store the response from each step. Allocate buffer size = 4
	// (max responses possible, from enricher, experiment engine, router and ensembler respectively).
	respCh := make(chan routerResponse, 4)
	defer close(respCh)

	// Defer logging request summary
	defer func() {
		go logTuringRouterRequestSummary(ctx, ctxLogger, time.Now(), req.Header, requestBody, respCh)
	}()

	// Enrich
	var resp mchttp.Response
	// Creates a new map to represent the merged headers from the original request headers + enricher response headers
	postEnrichmentResponseHeader := req.Header.Clone()

	payload := requestBody
	if h.IsEnricherEnabled() {
		resp, httpErr := h.Enrich(ctx, req.Header, payload)
		// Send enricher response/error for logging
		copyResponseToLogChannel(ctx, respCh, resultlog.ResultLogKeys.Enricher, resp, httpErr)
		// Check error
		if httpErr != nil {
			return nil, httpErr
		}
		// Merge Enricher response headers with original request headers
		for key := range resp.Header() {
			postEnrichmentResponseHeader.Set(key, resp.Header().Get(key))
		}
		// No error, copy response body
		payload = resp.Body()
	}

	// Route
	var expResp *experiment.Response
	expResp, resp, httpErr := h.Route(ctx, postEnrichmentResponseHeader, payload)
	if expResp != nil {
		var expErr *errors.TuringError
		if expResp.Error != "" {
			expErr = errors.NewTuringError(fmt.Errorf(expResp.Error), errors.HTTP)
		}
		if expResp.Configuration != nil || expErr != nil {
			copyResponseToLogChannel(ctx, respCh, resultlog.ResultLogKeys.Experiment, expResp, expErr)
		}
	}
	copyResponseToLogChannel(ctx, respCh, resultlog.ResultLogKeys.Router, resp, httpErr)
	if httpErr != nil {
		return nil, httpErr
	}
	payload = resp.Body()

	// Ensemble
	if h.IsEnsemblerEnabled() {
		resp, httpErr = h.Ensemble(ctx, postEnrichmentResponseHeader, requestBody, payload)
		copyResponseToLogChannel(ctx, respCh, resultlog.ResultLogKeys.Ensembler, resp, httpErr)
		if httpErr != nil {
			return nil, httpErr
		}
	}
	return resp, nil
}

func (h *httpHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var httpErr *errors.TuringError
	defer metrics.GetMeasureDurationFunc(httpErr, httpHandlerID)()

	// Create context from the request context
	ctx := turingctx.NewTuringContext(req.Context())
	// Create context logger
	ctxLogger := log.WithContext(ctx)
	defer func() {
		_ = ctxLogger.Sync()
	}()

	// Get the unique turing request id from the context
	turingReqID, err := turingctx.GetRequestID(ctx)
	if err != nil {
		ctxLogger.Errorf("Could not retrieve Turing Request ID from context: %v",
			err.Error())
	}
	ctxLogger.Debugf("Received request for %v", turingReqID)

	// Sets the turing request id in the header of the original request so that it gets sent to the enricher,
	// the experiment engine, the model routes, and the ensembler
	req.Header.Set(turingReqIDHeaderKey, turingReqID)

	if tracing.Glob().IsEnabled() {
		var sp opentracing.Span
		ctx, sp = h.enableTracingSpan(ctx, req, httpHandlerID)
		if sp != nil {
			defer sp.Finish()
		}
	}

	// Read the request body
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		h.error(ctx, rw, errors.NewTuringError(err, errors.HTTP))
		return
	}

	resp, httpErr := h.getPrediction(ctx, req, ctxLogger, requestBody)
	if httpErr != nil {
		h.error(ctx, rw, httpErr)
		return
	}
	payload := resp.Body()

	// Write the json response to the writer
	for key := range resp.Header() {
		rw.Header().Set(key, resp.Header().Get(key))
	}
	rw.Header().Set(turingReqIDHeaderKey, turingReqID)
	rw.WriteHeader(http.StatusOK)
	contentLength, err := rw.Write(payload)
	if err != nil {
		ctxLogger.Errorf("Error occurred when copying content: %v", err.Error())
	}
	ctxLogger.Debugf("Written %d bytes", contentLength)
}
