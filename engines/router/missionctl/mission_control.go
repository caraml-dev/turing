package missionctl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	fiberHttp "github.com/gojek/fiber/http"
	fiberProtocol "github.com/gojek/fiber/protocol"
	jsoniter "github.com/json-iterator/go"

	"github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/experiment"
	"github.com/caraml-dev/turing/engines/router/missionctl/fiberapi"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	mchttp "github.com/caraml-dev/turing/engines/router/missionctl/server/http"
)

// MissionControl is the base interface for the Turing Mission Control
type MissionControl interface {
	Enrich(
		ctx context.Context,
		header http.Header,
		body []byte,
	) (mchttp.Response, *errors.TuringError)
	Route(
		ctx context.Context,
		header http.Header,
		body []byte,
	) (*experiment.Response, mchttp.Response, *errors.TuringError)
	Ensemble(
		ctx context.Context,
		header http.Header,
		requestBody []byte,
		routerResponse []byte,
	) (mchttp.Response, *errors.TuringError)
	IsEnricherEnabled() bool
	IsEnsemblerEnabled() bool
}

// NewMissionControl creates new instance of the MissingControl,
// based on the http.Client and configuration passed into it
func NewMissionControl(
	client *http.Client,
	enrichmentCfg *config.EnrichmentConfig,
	routerCfg *config.RouterConfig,
	ensemblerCfg *config.EnsemblerConfig,
	appCfg *config.AppConfig,
) (MissionControl, error) {
	fiberRouter, err := fiberapi.CreateFiberRouterFromConfig(routerCfg.ConfigFile, appCfg.FiberDebugLog)
	if err != nil {
		return nil, err
	}

	if client == nil {
		client = http.DefaultClient
	}
	fiberHandler := fiberHttp.NewHandler(fiberRouter, fiberHttp.Options{Timeout: routerCfg.Timeout})

	return &missionControl{
		httpClient:        client,
		fiberHandler:      fiberHandler,
		enricherEndpoint:  enrichmentCfg.Endpoint,
		enricherTimeout:   enrichmentCfg.Timeout,
		routerTimeout:     routerCfg.Timeout,
		ensemblerEndpoint: ensemblerCfg.Endpoint,
		ensemblerTimeout:  ensemblerCfg.Timeout,
	}, nil
}

type missionControl struct {
	httpClient   *http.Client
	fiberHandler *fiberHttp.Handler

	enricherEndpoint string
	enricherTimeout  time.Duration

	routerTimeout time.Duration

	ensemblerEndpoint string
	ensemblerTimeout  time.Duration
}

func createNewHTTPRequest(
	ctx context.Context,
	httpMethod string,
	url string,
	header http.Header,
	body []byte,
) (*http.Request, error) {
	// Create new http request with the input body, ctx, url and method
	req, err := http.NewRequestWithContext(ctx, httpMethod, url,
		io.NopCloser(bytes.NewReader(body)))
	if err != nil {
		return nil, err
	}
	// Copy request headers from input
	for k, v := range header {
		for _, s := range v {
			req.Header.Set(k, s)
		}
	}
	return req, err
}

func (mc *missionControl) doPost(
	ctx context.Context,
	url string,
	header http.Header,
	body []byte,
	timeout time.Duration,
	componentLabel string,
) (mchttp.Response, *errors.TuringError) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := createNewHTTPRequest(ctx, http.MethodPost, url, header, body)
	if err != nil {
		return nil, errors.NewTuringError(err, fiberProtocol.HTTP)
	}

	// Make HTTP request and measure duration
	stopTimer := metrics.Glob().MeasureDurationMs(
		metrics.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(err == nil)
			},
			"component": func() string {
				return fmt.Sprint(componentLabel, "_makeRequest")
			},
		},
	)
	resp, err := mc.httpClient.Do(req)
	stopTimer()

	if err != nil {
		return nil, errors.NewTuringError(err, fiberProtocol.HTTP)
	}

	// Defer close non-nil response body
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NewTuringError(errors.Newf(errors.BadResponse,
			"Error response received: status – [%d]", resp.StatusCode), fiberProtocol.HTTP)
	}

	// No error, convert to mission control response and return
	mcResp, err := mchttp.NewCachedResponseFromHTTP(resp)
	if err != nil {
		return nil, errors.NewTuringError(err, fiberProtocol.HTTP)
	}
	return mcResp, nil
}

// Enrich method calls the configured enricher endpoint with the request body
// and returns the received response
func (mc *missionControl) Enrich(
	ctx context.Context,
	header http.Header,
	body []byte,
) (mchttp.Response, *errors.TuringError) {
	var httpErr *errors.TuringError
	// Measure execution time
	defer metrics.Glob().MeasureDurationMs(
		metrics.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(httpErr == nil)
			},
			"component": func() string {
				return "enrich"
			},
		},
	)()
	// Make HTTP request
	resp, httpErr := mc.doPost(ctx, mc.enricherEndpoint,
		header, body, mc.enricherTimeout, "enrich")
	return resp, httpErr
}

// Route dispatches the request the Fiber handler
func (mc *missionControl) Route(
	ctx context.Context,
	header http.Header,
	body []byte,
) (*experiment.Response, mchttp.Response, *errors.TuringError) {
	var routerErr *errors.TuringError
	// Measure execution time
	defer metrics.Glob().MeasureDurationMs(
		metrics.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(routerErr == nil)
			},
			"component": func() string {
				return "route"
			},
		},
	)()

	// Create a channel for experiment treatment response and add to context
	ch := make(chan *experiment.Response, 1)
	ctx = experiment.WithExperimentResponseChannel(ctx, ch)

	// Create a new POST request with the input body and header
	httpReq, err := createNewHTTPRequest(ctx, http.MethodPost, "", header, body)
	if err != nil {
		routerErr = errors.NewTuringError(err, fiberProtocol.HTTP)
		return nil, nil, routerErr
	}

	// Pass the request to the Fiber Handler and process the response
	var routerResp mchttp.Response
	fiberResponse, fiberError := mc.fiberHandler.DoRequest(httpReq)
	if fiberError != nil {
		routerResp, routerErr = nil, errors.NewTuringError(fiberError, fiberProtocol.HTTP, fiberError.Code)
	} else if fiberResponse == nil {
		routerResp, routerErr = nil, errors.NewTuringError(errors.Newf(errors.BadResponse,
			"Did not get back a valid response from the fiberHandler"), fiberProtocol.HTTP)
	} else if !fiberResponse.IsSuccess() {
		routerResp, routerErr = nil, errors.NewTuringError(errors.Newf(errors.BadResponse,
			"Error response received: status – [%d]", fiberResponse.StatusCode()), fiberProtocol.HTTP)
	} else {
		httpResp := fiberResponse.(*fiberHttp.Response)
		routerResp, routerErr = mchttp.NewCachedResponse(httpResp.Payload(), httpResp.Header()), nil
	}

	// Get the experiment treatment channel from the request context, read result
	var experimentResponse *experiment.Response
	expResultCh, err := experiment.GetExperimentResponseChannel(ctx)
	if err == nil {
		select {
		case experimentResponse = <-expResultCh:
		default:
			break
		}
	}

	return experimentResponse, routerResp, routerErr
}

// Ensemble dispatches the request to the configured ensembler endpoint
func (mc *missionControl) Ensemble(
	ctx context.Context,
	header http.Header,
	requestBody []byte,
	routerResponse []byte,
) (mchttp.Response, *errors.TuringError) {
	var httpErr *errors.TuringError
	// Measure execution time for Ensemble
	defer metrics.Glob().MeasureDurationMs(
		metrics.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(httpErr == nil)
			},
			"component": func() string {
				return "ensemble"
			},
		},
	)()

	// Combine the request body with the router response to make ensembler payload
	var err error
	// Measure execution time for creating the combined payload
	timer := metrics.Glob().MeasureDurationMs(
		metrics.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(err == nil)
			},
			"component": func() string {
				return "ensemble_makePayload"
			},
		},
	)
	payload, err := makeEnsemblerPayload(requestBody, routerResponse)
	timer()
	if err != nil {
		httpErr = errors.NewTuringError(err, fiberProtocol.HTTP)
		return nil, httpErr
	}

	// Make HTTP request
	resp, httpErr := mc.doPost(ctx, mc.ensemblerEndpoint,
		header, payload, mc.ensemblerTimeout, "ensemble")
	return resp, httpErr
}

func (mc *missionControl) IsEnricherEnabled() bool {
	return mc.enricherEndpoint != ""
}

func (mc *missionControl) IsEnsemblerEnabled() bool {
	return mc.ensemblerEndpoint != ""
}

func makeEnsemblerPayload(reqBody []byte, routerResp []byte) ([]byte, error) {
	payload := ensemblerPayload{
		Request:  reqBody,
		Response: routerResp,
	}
	return jsoniter.Marshal(payload)
}

type ensemblerPayload struct {
	// Original request payload
	Request json.RawMessage `json:"request"`
	// Response from the Fiber router
	Response json.RawMessage `json:"response"`
}
