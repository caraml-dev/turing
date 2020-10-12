package fiberapi

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
	"github.com/gojek/fiber"
	"github.com/gojek/turing/engines/experiment/common"
	"github.com/gojek/turing/engines/router/missionctl/errors"
	"github.com/gojek/turing/engines/router/missionctl/internal"
	"github.com/gojek/turing/engines/router/missionctl/log"
)

var (
	validation = func() *validator.Validate {
		instance := validator.New()
		_ = instance.RegisterValidation("notBlank", validators.NotBlank)

		return instance
	}()
)

// TrafficSplittingStrategyRule represents one rule of the TrafficSplittingStrategy
// Each rule maps set of conditions to one route configured on fiber router
type TrafficSplittingStrategyRule struct {
	RouteID    string                         `json:"route_id" validate:"required,notBlank"`
	Conditions []*common.TrafficRuleCondition `json:"conditions" validate:"required,notBlank,dive"`
}

// TestRequest checks if the request satisfies all conditions of this rule
func (r *TrafficSplittingStrategyRule) TestRequest(reqHeader http.Header, bodyBytes []byte) (bool, error) {
	safeCh := internal.NewSafeChan(1)
	defer safeCh.Close()

	var wg sync.WaitGroup
	wg.Add(len(r.Conditions))

	// test each condition asynchronously
	for _, condition := range r.Conditions {
		go func(condition *common.TrafficRuleCondition) {
			res, err := condition.TestRequest(reqHeader, bodyBytes)
			if err != nil {
				log.Glob().Infof(
					"Failed to test if request matches traffic-splitting condition: %s", err)
			}

			if !res {
				safeCh.Write(false)
			}

			wg.Done()
		}(condition)
	}

	// wait for all conditions to be tested and write `true` into results channel
	go func() {
		wg.Wait()

		safeCh.Write(true)
	}()

	// return the first value from the channel
	return (<-safeCh.Read()).(bool), nil
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
) (fiber.Component, []fiber.Component, error) {
	doneCh := make(chan interface{}, 1)
	errCh := make(chan error, 1)

	defer close(doneCh)
	defer close(errCh)

	orderedRoutes := []fiber.Component{}
	// array, that holds results of testing the request by each rule configured on the strategy
	// `results[k]` – is `true` if the request satisfies `k`th rule of the strategy, and `false`
	// otherwise
	results := make([]bool, len(s.Rules))

	var wg sync.WaitGroup
	wg.Add(len(s.Rules))

	for idx, rule := range s.Rules {
		// test each rule asynchronously and write results into results array
		go func(rule *TrafficSplittingStrategyRule, idx int) {
			if res, err := rule.TestRequest(req.Header(), req.Payload()); err != nil {
				errCh <- err
			} else {
				results[idx] = res
			}
			wg.Done()
		}(rule, idx)
	}

	go func() {
		wg.Wait()
		doneCh <- true
	}()

	// wait for all rules to be tested or until an error appears in error channel
	select {
	case <-doneCh:
	case err := <-errCh:
		log.WithContext(ctx).Errorf(err.Error())
		return nil, nil, createFiberError(err)
	}

	// select primary route and fallbacks, based on the results of testing
	// given request against traffic-splitting rules
	for i := 0; i < len(results); i++ {
		routeID := s.Rules[i].RouteID
		if results[i] {
			if r, exists := routes[routeID]; exists {
				orderedRoutes = append(orderedRoutes, r)
			} else {
				err := errors.Newf(errors.BadConfig, `route with id "%s" doesn't exist in the router`, routeID)
				log.WithContext(ctx).Errorf(err.Error())
				return nil, nil, createFiberError(err)
			}
		}
	}

	// given request hasn't satisfied any of the rules configured on this routing strategy
	if len(orderedRoutes) == 0 {
		if defaultRoute, exist := routes[s.DefaultRouteID]; exist {
			orderedRoutes = append(orderedRoutes, defaultRoute)
		} else {
			err := errors.Newf(errors.NotFound, "http request didn't match any traffic rule")
			log.WithContext(ctx).Errorf(err.Error())
			return nil, nil, createFiberError(err)
		}
	}

	return orderedRoutes[0], orderedRoutes[1:], nil
}
