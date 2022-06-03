package fiberapi

import (
	"context"
	"encoding/json"

	"github.com/gojek/fiber"
	"github.com/gojek/turing/engines/experiment/runner"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/experiment"
	"github.com/gojek/turing/engines/router/missionctl/log"
	"github.com/gojek/turing/engines/router/missionctl/turingctx"
)

// DefaultTuringRoutingStrategy selects the route that matches experiment treatment for a
// given unit and picks the route marked default, if any, as fallback.
type DefaultTuringRoutingStrategy struct {
	*experimentationPolicy
	*routeSelectionPolicy
}

// Initialize is invoked by the Fiber library to initialize a new RoutingStrategy.
func (r *DefaultTuringRoutingStrategy) Initialize(properties json.RawMessage) error {
	var err error
	// Initialize appropriate fields
	r.experimentationPolicy, err = newExperimentationPolicy(properties)
	if err != nil {
		return errors.Wrapf(err, "Failed initializing experimentation policy on routing strategy")
	}
	r.routeSelectionPolicy, err = newRouteSelectionPolicy(properties)
	if err != nil {
		return errors.Wrapf(err, "Failed initializing route selection policy on routing strategy")
	}
	// Check that the default route is not empty
	if r.routeSelectionPolicy.defaultRoute == "" {
		return errors.Newf(errors.BadConfig, "No default route defined")
	}
	return nil
}

// SelectRoute decides the priority order of the routes for the unit in the given request,
// according to the treatment returned by the configured experiment engine.
func (r *DefaultTuringRoutingStrategy) SelectRoute(
	ctx context.Context,
	req fiber.Request,
	routes map[string]fiber.Component,
) (fiber.Component, []fiber.Component, error) {
	// Get fallback
	fallbacks := []fiber.Component{}
	if defRoute, ok := routes[r.defaultRoute]; ok {
		fallbacks = append(fallbacks, defRoute)
	}

	// Get the experiment treatment
	turingReqID, _ := turingctx.GetRequestID(ctx)
	options := runner.GetTreatmentOptions{
		TuringRequestID: turingReqID,
	}
	expPlan, expErr := r.experimentEngine.
		GetTreatmentForRequest(req.Header(), req.Payload(), options)

	// Create experiment response object
	experimentResponse := experiment.NewResponse(expPlan, expErr)
	// Copy experiment response to the result channel in the context
	expResultCh, expChErr := experiment.GetExperimentResponseChannel(ctx)
	if expChErr == nil {
		expResultCh <- experimentResponse
		close(expResultCh)
	}

	// If error, log it and return the fallback(s)
	if expErr != nil {
		log.WithContext(ctx).Errorf(expErr.Error())
		return nil, fallbacks, nil
	}

	for _, m := range r.experimentMappings {
		if m.Experiment == expPlan.ExperimentName && m.Treatment == expPlan.Name {
			// Stop matching on first match because only 1 route is required. Don't send in fallbacks,
			// because we do not want to suppress the error from the preferred route.
			return routes[m.Route], []fiber.Component{}, nil
		}
	}

	// primary route will be nil if there are no matching treatments in the mapping
	return nil, fallbacks, nil
}
