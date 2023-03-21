package fiberapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/gojek/fiber"
	fiberHttp "github.com/gojek/fiber/http"
	"github.com/stretchr/testify/assert"

	runnerV1 "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner"
	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/caraml-dev/turing/engines/router/missionctl/experiment"
	tfu "github.com/caraml-dev/turing/engines/router/missionctl/fiberapi/internal/testutils"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
)

type testSuiteInitFanIn struct {
	properties json.RawMessage
	success    bool
	expected   EnsemblingFanIn
	err        string
}

// Create fan in for testing using the mock experiment set up
// and mock responses
var efi = &EnsemblingFanIn{
	&experimentationPolicy{
		experimentEngine: tfu.MockExperimentRunner{
			Treatment: &runner.Treatment{
				ExperimentName: "test_experiment",
				Name:           "treatment-A",
				Config:         json.RawMessage(`{"test_config": "placeholder"}`),
			},
		},
	},
	&routeSelectionPolicy{
		defaultRoute: "control",
	},
}
var efiExpTimeout = &EnsemblingFanIn{
	&experimentationPolicy{
		experimentEngine: tfu.MockExperimentRunner{
			Timeout:     time.Second * 2,
			WantTimeout: true,
		},
	},
	&routeSelectionPolicy{
		defaultRoute: "control",
	},
}

func TestInitializeEnsemblingFanIn(t *testing.T) {
	tests := map[string]testSuiteInitFanIn{
		"success | all properties": {
			properties: json.RawMessage(`{
				"default_route_id":  "route1",
				"experiment_engine": "Test"
			}`),
			success: true,
			expected: EnsemblingFanIn{
				routeSelectionPolicy: &routeSelectionPolicy{
					defaultRoute: "route1",
				},
				experimentationPolicy: &experimentationPolicy{
					experimentEngine: nil,
				},
			},
		},
		"success | missing route policy": {
			properties: json.RawMessage(`{
				"experiment_engine": "Test"
			}`),
			success: true,
			expected: EnsemblingFanIn{
				routeSelectionPolicy:  &routeSelectionPolicy{},
				experimentationPolicy: &experimentationPolicy{},
			},
		},
		"missing experimentation policy": {
			properties: json.RawMessage(`invalid_data`),
			success:    false,
			err:        "Failed initializing experimentation policy on FanIn: Failed to parse experimentation policy",
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Set up
			fanIn := EnsemblingFanIn{}

			// Monkey patch functionality that is external to the current package and run
			monkey.Patch(
				runnerV1.Get,
				func(name string, config json.RawMessage) (runner.ExperimentRunner, error) {
					return nil, nil
				},
			)
			monkey.Patch(
				experiment.NewExperimentRunner,
				func(_ string, _ map[string]interface{}, _ int) (runner.ExperimentRunner, error) {
					return nil, nil
				},
			)
			err := fanIn.Initialize(data.properties)

			monkey.Unpatch(experiment.NewExperimentRunner)
			monkey.Unpatch(runnerV1.Get)

			// Test error and if fanIn is initialised as expected
			assert.Equal(t, data.success, err == nil)
			if data.success {
				assert.Equal(t, data.expected, fanIn)
			} else {
				// Validate error
				tu.FailOnNil(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}

func TestEnsemblingFanInAggregate(t *testing.T) {
	tests := map[string]struct {
		fanIn            *EnsemblingFanIn
		expectedResponse string
	}{
		"success": {
			fanIn: efi,
			expectedResponse: string(`{
				"experiment": {
					"configuration": {
						"test_config": "placeholder"
					}
				},
				"route_responses": [
					{  
						"data": {"value":"treatment-A"},
						"is_default": false,
						"route": "treatment-A"
					},
					{  
						"data": {"value":"treatment-B"},
						"is_default": false,
						"route": "treatment-B"
					}
				]
			}`),
		},
		// Experiment Engine timeout > timeout in ctx passed to Aggregate
		"experiment timeout greater": {
			fanIn: efiExpTimeout,
			expectedResponse: string(`{
				"experiment": {},
				"route_responses": [
					{  
						"data": {"value":"treatment-A"},
						"is_default": false,
						"route": "treatment-A"
					},
					{  
						"data": {"value":"treatment-B"},
						"is_default": false,
						"route": "treatment-B"
					}
				]
			}`),
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Create test responses, add to queue
			respQueue := makeTestResponseQueue("treatment-A", "treatment-B")

			// Create test request
			req := tu.MakeTestRequest(t, tu.NopHTTPRequestModifier)
			fiberReq, err := fiberHttp.NewHTTPRequest(req)
			tu.FailOnError(t, err)

			// Call Aggregate with a timeout, exp result channel, test req and response queue
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			expCh := make(chan *experiment.Response, 1)
			resp := data.fanIn.Aggregate(
				experiment.WithExperimentResponseChannel(ctx, expCh), fiberReq, respQueue)

			// Check response status code is Success
			if status := resp.StatusCode(); status != http.StatusOK {
				t.Errorf("Status code mismatch. Expected %d .\n Got %d instead", http.StatusOK, status)
			}

			// Unmarshal the JSON response, sort the individual responses by treatment name and compare.
			// Expected is already defined in the sorted order. This is done because arrays are order
			// dependent and JSONEq will fail otherwise.
			var f CombinedResponse
			tu.FailOnError(t, json.Unmarshal(resp.Payload(), &f))
			sort.Slice(f.RouteResponses, func(i, j int) bool {
				return f.RouteResponses[i].Route < f.RouteResponses[j].Route
			})
			fBytes, err := json.Marshal(f)
			tu.FailOnError(t, err)

			// Check that the final collected response matches the expected value
			actual := string(fBytes)
			assert.JSONEq(t, data.expectedResponse, actual, "Response body mismatch.")
		})
	}
}

func makeTestResponseQueue(endpointNames ...string) fiber.ResponseQueue {
	// Make response channel
	ch := make(chan fiber.Response, len(endpointNames))

	// Populate fiber responses into the channel
	for _, e := range endpointNames {
		payload := fmt.Sprintf(`{"value": "%s"}`, e)
		body := io.NopCloser(bytes.NewReader([]byte(payload)))
		resp := &http.Response{
			StatusCode: 200,
			Body:       body,
		}
		fiberResp := fiberHttp.NewHTTPResponse(resp).WithBackendName(e)
		ch <- fiberResp
	}

	// Wrap into response queue and return
	return fiber.NewResponseQueue(ch, len(endpointNames))
}
