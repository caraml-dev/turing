package fiberapi

import (
	"context"
	"encoding/json"
	"fmt"

	upiv1 "github.com/caraml-dev/universal-prediction-interface/gen/go/grpc/caraml/upi/v1"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
	"github.com/gojek/fiber"
	grpcFiber "github.com/gojek/fiber/grpc"
	fiberProtocol "github.com/gojek/fiber/protocol"

	"github.com/caraml-dev/turing/engines/router"
	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
)

var (
	validation = func() *validator.Validate {
		instance := validator.New()
		_ = instance.RegisterValidation("notBlank", validators.NotBlank)

		return instance
	}()
	TrafficRuleLabel string = "traffic-rule"
)

// TrafficSplittingStrategyRule represents one rule of the TrafficSplittingStrategy
// Each rule maps set of conditions to one route configured on fiber router
type TrafficSplittingStrategyRule struct {
	RouteID    string                         `json:"route_id" validate:"required,notBlank"`
	Conditions []*router.TrafficRuleCondition `json:"conditions" validate:"required,dive"`
}

// TestRequest checks if the request satisfies all conditions of this rule
func (r *TrafficSplittingStrategyRule) TestRequest(req fiber.Request) (bool, error) {
	switch req.Protocol() {
	case fiberProtocol.HTTP:
		// test all condition and return immediately if one condition is not satisfied
		for _, condition := range r.Conditions {
			res, err := condition.TestRequest(req)
			if err != nil {
				log.Glob().Infof(
					"Failed to test if request matches traffic-splitting condition: %s", err)
			}

			if !res {
				// short circuit
				return false, nil
			}
		}
	case fiberProtocol.GRPC:
		grpcFiberReq, ok := req.(*grpcFiber.Request)
		if !ok {
			err := fmt.Errorf("failed to convert into grpc fiber request")
			log.Glob().Error(err.Error())
			return false, err
		}

		upiReq, ok := grpcFiberReq.ProtoMessage().(*upiv1.PredictValuesRequest)
		if !ok {
			err := fmt.Errorf("failed to convert into upi request")
			log.Glob().Error(err.Error())
			return false, err
		}

		// test all condition and return immediately if one condition is not satisfied
		for _, condition := range r.Conditions {
			res, err := condition.TestUPIRequest(upiReq, req.Header())
			if err != nil {
				log.Glob().Infof(
					"Failed to test if request matches traffic-splitting condition: %s", err)
			}

			if !res {
				// short circuit
				return false, nil
			}
		}
	}
	// return the first value from the channel
	return true, nil
}

// TrafficSplittingStrategy selects the route based on the traffic splitting
// conditions, configured on this strategy
type TrafficSplittingStrategy struct {
	DefaultRouteID string                          `json:"default_route_id"`
	Rules          []*TrafficSplittingStrategyRule `json:"rules" validate:"required,notBlank,dive"`
}

// Initialize is invoked by the Fiber library to initialize this strategy
// with the configuration
func (s *TrafficSplittingStrategy) Initialize(properties json.RawMessage) error {
	if err := json.Unmarshal(properties, s); err != nil {
		return errors.Wrapf(err, "Failed initializing traffic splitting strategy")
	}
	if err := validation.Struct(s); err != nil {
		return errors.Wrapf(err, "Failed initializing traffic splitting strategy")
	}

	return nil
}

// SelectRoute picks primary and fallback routes based the
// traffic-splitting rules configured on this strategy
func (s *TrafficSplittingStrategy) SelectRoute(
	ctx context.Context,
	req fiber.Request,
	routes map[string]fiber.Component,
) (fiber.Component, []fiber.Component, fiber.Labels, error) {
	labels := fiber.NewLabelsMap()

	for _, rule := range s.Rules {
		res, err := rule.TestRequest(req)
		if err != nil {
			return nil, nil, labels, createFiberError(err, req.Protocol())
		} else if res {
			routeID := rule.RouteID
			if r, exists := routes[routeID]; exists {
				return r, []fiber.Component{}, labels.WithLabel(TrafficRuleLabel, r.ID()), nil
			}
			// This is unexpected, terminate with error.
			err := errors.Newf(errors.BadConfig, `route with id "%s" doesn't exist in the router`, routeID)
			log.WithContext(ctx).Errorf(err.Error())
			return nil, nil, labels, createFiberError(err, req.Protocol())
		}
	}

	// Given request hasn't satisfied any of the rules configured on this routing strategy;
	// check if default route exists.
	if defaultRoute, exist := routes[s.DefaultRouteID]; exist {
		return defaultRoute, []fiber.Component{}, labels.WithLabel(TrafficRuleLabel, defaultRoute.ID()), nil
	}

	// No matches whatsoever.
	err := errors.Newf(errors.NotFound, "http request didn't match any traffic rule")
	log.WithContext(ctx).Errorf(err.Error())
	return nil, nil, labels, createFiberError(err, req.Protocol())
}
