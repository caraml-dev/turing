package fiberapi

import (
	"context"
	"encoding/json"

	"github.com/gojek/fiber"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/experiment"
	"github.com/gojek/turing/engines/router/missionctl/log"
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
	logger := log.WithContext(ctx)
	expPlan, expErr := r.experimentEngine.
		GetTreatmentForRequest(ctx, logger, req.Header(), req.Payload())

	// Create experiment response object
	experimentResponse := experiment.NewResponse(expPlan, expErr)
	// Copy experiment response to the result channel in the context
	expResultCh, expChErr := experiment.GetExperimentResponseChannel(ctx)
	if expChErr == nil {
		expResultCh <- experimentResponse
		close(expResultCh)
	}

	// If error, return it
	if expErr != nil {
		log.WithContext(ctx).Errorf(expErr.Error())
		return nil, fallbacks, createFiberError(expErr)
	}

	// selectedRoute is the route that should be visited first based on the the experiment treatment and mappings (if any)
	// selectedRoute will be nil if the experiment treatment has no corresponding mappings
	var selectedRoute fiber.Component

	for _, m := range r.experimentMappings {
		if m.Experiment == expPlan.ExperimentName && m.Treatment == expPlan.Name {
			selectedRoute = routes[m.Route]
			// stop matching on first match because only 1 selected route is required
			break
		}
	}

	return selectedRoute, fallbacks, nil
}
