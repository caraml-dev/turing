package fiberapi

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/experiment"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
	"github.com/gojek/fiber"
	"github.com/gojek/fiber/protocol"
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

	if req.Protocol() != protocol.GRPC {
		// TODO need to handle for grpc
		expPlan, expErr := r.experimentEngine.
			GetTreatmentForRequest(req.Header(), req.Payload().([]byte), options)

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

		// For the DefaultTuringRoutingStrategy, we only expect experimentMappings OR routeNamePath to be configured; we
		// perform a check on both of them and determine the final route response to return
		for _, m := range r.experimentMappings {
			if m.Experiment == expPlan.ExperimentName && m.Treatment == expPlan.Name {
				// Stop matching on first match because only 1 route is required. Don't send in fallbacks,
				// because we do not want to suppress the error from the preferred route.
				return routes[m.Route], []fiber.Component{}, nil
			}
		}

		// Use the route name path to locate the name of the route to be returned as the final response
		if r.routeSelectionPolicy.routeNamePath != "" {
			routeName, err := jsonparser.GetString(
				experimentResponse.Body(),
				strings.Split(r.routeSelectionPolicy.routeNamePath, ".")...,
			)

			if err != nil {
				log.WithContext(ctx).Errorf(err.Error())
				return nil, fallbacks, nil
			}

			if selectedRoute, ok := routes[routeName]; ok {
				return selectedRoute, []fiber.Component{}, nil
			}

			// There are no routes with the route name found in the treatment
			log.WithContext(ctx).Errorf(
				"No route found corresponding to the route name found in the treatment:, %s",
				routeName,
			)
		}
	}

	// primary route will be nil if there are no matching treatments in the mapping
	return nil, fallbacks, nil
}
