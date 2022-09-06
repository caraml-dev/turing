package fiberapi

import (
	"time"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/gojek/fiber"
	"github.com/gojek/fiber/config"
	fibererror "github.com/gojek/fiber/errors"
	fiberhttp "github.com/gojek/fiber/http"
	"github.com/gojek/fiber/protocol"
	"github.com/gojek/fiber/types"
)

// CreateFiberRouterFromConfig creates a Fiber router from config
func CreateFiberRouterFromConfig(
	cfgFilePath string,
	fiberDebugLog bool,
) (fiber.Component, error) {

	component, err := createRouterFromConfigFile(cfgFilePath)
	if err != nil {
		return nil, err
	}

	// Create required interceptors
	interceptors := []fiber.Interceptor{
		NewErrorLoggingInterceptor(log.Glob()),
		NewMetricsInterceptor(),
	}

	if fiberDebugLog {
		// Enable debug logs from Fiber - create a new time logging interceptor
		interceptors = append(interceptors, NewTimeLoggingInterceptor(log.Glob()))
	}
	if tracing.Glob().IsEnabled() {
		// Create a tracing interceptor
		interceptors = append(interceptors, NewTracingInterceptor())
	}

	// Add the interceptors to the Fiber component
	component.AddInterceptor(true, interceptors...)

	return component, err
}

// CreateFiberRequestHandler creates a new Fiber component from the given config file,
// associates it with a Fiber HTTP handler and returns it.
func CreateFiberRequestHandler(
	component fiber.Component,
	timeout time.Duration,
) *fiberhttp.Handler {
	return fiberhttp.NewHandler(component, fiberhttp.Options{Timeout: timeout})
}

// createRouterFromConfigFile takes the path to a fiber config file,
// registers the necessary types and initialises the router.
func createRouterFromConfigFile(cfgFilePath string) (fiber.Component, error) {
	err := registerFiberTypes()
	if err != nil {
		return nil, err
	}

	return config.InitComponentFromConfig(cfgFilePath)
}

// createFiberError wraps the input error in a format that is usable by Fiber
func createFiberError(err error, protocol protocol.Protocol) fibererror.FiberError {
	return fibererror.FiberError{
		Code:    errors.GetErrorCode(err, errors.ErrorProtocol(protocol)),
		Message: err.Error(),
	}
}

// registerFiberTypes registers the customer Fiber types defined in the fiberapi module
func registerFiberTypes() error {
	err := types.InstallType("fiber.DefaultTuringRoutingStrategy",
		&DefaultTuringRoutingStrategy{})
	if err != nil {
		return err
	}

	err = types.InstallType("fiber.EnsemblingFanIn", &EnsemblingFanIn{})
	if err != nil {
		return err
	}

	err = types.InstallType("fiber.TrafficSplittingStrategy", &TrafficSplittingStrategy{})
	if err != nil {
		return err
	}
	return nil
}
