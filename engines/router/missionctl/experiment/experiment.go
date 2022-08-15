package experiment

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/caraml-dev/turing/engines/experiment"
	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
)

// NewExperimentRunner returns an instance of the Planner, based on the input engine name
func NewExperimentRunner(name string, cfg map[string]interface{}) (runner.ExperimentRunner, error) {
	factory, err := experiment.NewEngineFactory(name, cfg, log.Glob())
	if err != nil {
		return nil, err
	}

	engine, err := factory.GetExperimentRunner()
	if err != nil {
		return nil, err
	}

	interceptors := []runner.Interceptor{
		NewMetricsInterceptor(),
	}

	return runner.NewInterceptRunner(name, engine, interceptors...), nil

}

// Response holds the experiment configuration / error response,
// also satisfies the missionctl/http Response interface
type Response struct {
	// Success response from the experiment engine, unmodified
	Configuration json.RawMessage `json:"configuration,omitempty"`
	// Error message
	Error string `json:"error,omitempty"`
}

// Body satisfies the Response interface, returning the raw configuration
func (r *Response) Body() []byte {
	return r.Configuration
}

// Header satisfies the Response interface, returns nil
func (r *Response) Header() http.Header {
	return nil
}

// NewResponse is a helper function to create an object of type Response
// that holds the experiment treatment / appropriate error information
func NewResponse(expPlan *runner.Treatment, expPlanErr error) *Response {
	// Create experiment response object
	experimentResponse := &Response{}
	if expPlanErr != nil {
		// Failed retrieving experiment treatment, populate the error field
		experimentResponse.Error = expPlanErr.Error()
	} else {
		experimentResponse.Configuration = expPlan.Config
	}
	return experimentResponse
}

// WithExperimentResponseChannel associates a pointer to a channel of type *Response
// to the given context object
func WithExperimentResponseChannel(ctx context.Context, ch chan *Response) context.Context {
	return context.WithValue(ctx, turingctx.TuringTreatmentChannelKey, ch)
}

// GetExperimentResponseChannel returns a pointer to a channel of type *Response,
// which holds the experiment treatment / error, from the input context
func GetExperimentResponseChannel(ctx context.Context) (chan *Response, error) {
	if ctxValue, ok := ctx.Value(turingctx.TuringTreatmentChannelKey).(chan *Response); ok {
		return ctxValue, nil
	}
	return nil, errors.Newf(errors.Unknown, "Experiment treatment channel not found in the context")
}
