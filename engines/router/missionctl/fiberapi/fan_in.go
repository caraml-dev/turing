package fiberapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gojek/fiber"
	fiberHttp "github.com/gojek/fiber/http"
	jsoniter "github.com/json-iterator/go"
	"github.com/opentracing/opentracing-go"

	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/experiment"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
)

// FanInID is used to indendify the fan in component when capturing a request span
const FanInID = "fan_in"

// EnsemblingFanIn combines the results from the fanout with the experiment parameters
// and forwards to the configured ensembling endpoint
type EnsemblingFanIn struct {
	*experimentationPolicy
	*routeSelectionPolicy
}

// Initialize is invoked by the Fiber library to initialize a new FanIn.
func (fanIn *EnsemblingFanIn) Initialize(properties json.RawMessage) error {
	var err error
	// Initialize appropriate fields
	fanIn.experimentationPolicy, err = newExperimentationPolicy(properties)
	if err != nil {
		return errors.Wrapf(err, "Failed initializing experimentation policy on FanIn")
	}
	fanIn.routeSelectionPolicy, err = newRouteSelectionPolicy(properties)
	if err != nil {
		return errors.Wrapf(err, "Failed initializing route selection policy on FanIn")
	}
	return nil
}

// Aggregate requests for the treatment parameters from the configured experiment engine,
// collects the results from the fanout, dispatches the combined data to the configured
// ensembling endpoint and returns the result
func (fanIn *EnsemblingFanIn) Aggregate(
	ctx context.Context,
	req fiber.Request,
	queue fiber.ResponseQueue,
) fiber.Response {
	// Monitor for the results
	respCh := queue.Iter()
	// Store the available results and the experiment response
	responses := make(map[string]fiber.Response)
	var experimentResponse *experiment.Response
	// Obtain the experiment response channel from the context, to write the response to
	expCtxCh, expCtxChErr := experiment.GetExperimentResponseChannel(ctx)
	if expCtxChErr == nil {
		defer close(expCtxCh)
	}

	// Request for the experiment treatment and save it to expRespCh asynchronously
	expRespCh := make(chan *experiment.Response, 1)
	go func() {
		turingReqID, _ := turingctx.GetRequestID(ctx)
		options := runner.GetTreatmentOptions{
			TuringRequestID: turingReqID,
		}

		expPlan, expPlanErr := fanIn.experimentEngine.
			GetTreatmentForRequest(req.Header(), req.Payload(), options)
		// Write to channel
		expRespCh <- experiment.NewResponse(expPlan, expPlanErr)
		close(expRespCh)
	}()

	// Wait on the results from the individual routes and experiment engine, for as long as
	// timeout is not reached and we don't have all the results.
	timeout := false
	for (respCh != nil || expRespCh != nil) && !timeout {
		select {
		case resp, ok := <-respCh:
			if ok {
				responses[resp.BackendName()] = resp
			} else {
				// Channel closed, stop reading from respCh
				respCh = nil
			}
		case expResp, ok := <-expRespCh:
			if ok {
				// Update the experiment response to be returned
				experimentResponse = expResp
				// Copy experiment response to the experiment result channel in the context
				if expCtxChErr == nil {
					expCtxCh <- expResp
				}
			} else {
				// Channel closed, stop reading from expRespCh
				expRespCh = nil
			}
		case <-ctx.Done():
			timeout = true
		}
	}

	// Associate span to context to trace response ensembling, if tracing enabled
	if tracing.Glob().IsEnabled() {
		var sp opentracing.Span
		sp, _ = tracing.Glob().StartSpanFromContext(ctx, FanInID)
		if sp != nil {
			defer sp.Finish()
		}
	}
	return fanIn.collectResponses(responses, experimentResponse)
}

// collectResponses collects all responses and treatment in the
// format expected by the ensembler
func (fanIn *EnsemblingFanIn) collectResponses(
	responses map[string]fiber.Response,
	expResponse *experiment.Response,
) fiber.Response {
	result := CombinedResponse{
		RouteResponses: make([]RouteResponse, len(responses)),
	}
	if expResponse != nil {
		result.Experiment = *expResponse
	}

	// Collect all treatment responses
	idx := 0
	for k, v := range responses {
		t := RouteResponse{
			Route:     k,
			Data:      v.Payload(),
			IsDefault: k == fanIn.defaultRoute,
		}
		result.RouteResponses[idx] = t
		idx++
	}

	// Marshal the response, measure time
	var err error
	timer := metrics.Glob().MeasureDurationMs(
		metrics.TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return metrics.GetStatusString(err == nil)
			},
			"component": func() string {
				return "fanin_marshalResponse"
			},
		},
	)
	rBytes, err := jsoniter.Marshal(result)
	timer()
	if err != nil {
		return fiber.NewErrorResponse(err)
	}

	// Return successful response
	resp := http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBuffer(rBytes)),
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}
	return fiberHttp.NewHTTPResponse(&resp)
}

// RouteResponse captures the result of each experiment
type RouteResponse struct {
	Route     string          `json:"route"`
	Data      json.RawMessage `json:"data"`
	IsDefault bool            `json:"is_default"`
}

// CombinedResponse captures the structure of the final response sent back by the fan in
type CombinedResponse struct {
	// List of responses from each treatment
	RouteResponses []RouteResponse `json:"route_responses"`
	// Configuration / Error response from experiment engine
	Experiment experiment.Response `json:"experiment"`
}
