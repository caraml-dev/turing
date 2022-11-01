package fiberapi

import (
	"context"
	"time"

	"github.com/gojek/fiber"
	"github.com/opentracing/opentracing-go"

	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/tracing"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
)

type ctxKey string

const (
	startTimeKey ctxKey = "startTimeKey"
)

/////////////////////////// TimeLoggingInterceptor ////////////////////////////

// NewTimeLoggingInterceptor is a creator for a TimeLoggingInterceptor
func NewTimeLoggingInterceptor(log log.Logger) fiber.Interceptor {
	return &TimeLoggingInterceptor{
		logger: log,
	}
}

// TimeLoggingInterceptor is the structural interceptor used for logging responses
type TimeLoggingInterceptor struct {
	logger log.Logger
	*fiber.NoopAfterDispatchInterceptor
}

// BeforeDispatch associates the start time to the context
func (i *TimeLoggingInterceptor) BeforeDispatch(
	ctx context.Context,
	req fiber.Request,
) context.Context {
	ctx = context.WithValue(ctx, startTimeKey, time.Now())
	return ctx
}

// AfterCompletion logs the time taken for the component to process the request
// and the response status
func (i *TimeLoggingInterceptor) AfterCompletion(
	ctx context.Context,
	req fiber.Request,
	queue fiber.ResponseQueue,
) {
	cID := ctx.Value(fiber.CtxComponentIDKey)
	turingReqID, _ := turingctx.GetRequestID(ctx)

	if startTime, ok := ctx.Value(startTimeKey).(time.Time); ok {
		i.logger.Debugw("Time Taken", "Component", cID, "Turing Request",
			turingReqID, "Duration", time.Since(startTime).String())
	}
}

////////////////////////// ErrorLoggingInterceptor ////////////////////////////

// NewErrorLoggingInterceptor is a creator for an ErrorLoggingInterceptor
func NewErrorLoggingInterceptor(log log.Logger) fiber.Interceptor {
	return &ErrorLoggingInterceptor{
		logger: log,
	}
}

// ErrorLoggingInterceptor is the structural interceptor used for logging error
// responses from individual routes
type ErrorLoggingInterceptor struct {
	logger log.Logger
	*fiber.NoopBeforeDispatchInterceptor
	*fiber.NoopAfterDispatchInterceptor
}

// AfterCompletion logs the response summary if any route returns an error
func (i *ErrorLoggingInterceptor) AfterCompletion(
	ctx context.Context,
	req fiber.Request,
	queue fiber.ResponseQueue,
) {
	cID := ctx.Value(fiber.CtxComponentIDKey)
	cKind := ctx.Value(fiber.CtxComponentKindKey)

	if cKindCompType, ok := cKind.(fiber.ComponentKind); ok {
		// Only log errors at the the caller
		// (to avoid repeatedly logging the error at higher levels)
		if cKindCompType == fiber.CallerKind {
			turingReqID, _ := turingctx.GetRequestID(ctx)

			// For each response in the queue, if the status is non-success, log warning
			for resp := range queue.Iter() {
				if !resp.IsSuccess() {
					i.logger.Warnw("Route Error", "Component", cID,
						"Turing Request", turingReqID,
						"Status", resp.StatusCode(),
						"Response", string(resp.Payload()))
				}
			}
		}
	}
}

//////////////////////// MetricsInterceptor ///////////////////////////

// NewMetricsInterceptor is a creator for a MetricsInterceptor
func NewMetricsInterceptor() fiber.Interceptor {
	return &MetricsInterceptor{}
}

// MetricsInterceptor is the structural interceptor used for capturing
// run time metrics from the Fiber components
type MetricsInterceptor struct {
	*fiber.NoopAfterDispatchInterceptor
}

// BeforeDispatch associates the start time to the context
func (i *MetricsInterceptor) BeforeDispatch(
	ctx context.Context,
	req fiber.Request,
) context.Context {
	ctx = context.WithValue(ctx, startTimeKey, time.Now())
	return ctx
}

// AfterCompletion logs the time taken for the component to process the request,
// to the metrics collector
func (i *MetricsInterceptor) AfterCompletion(
	ctx context.Context,
	req fiber.Request,
	queue fiber.ResponseQueue,
) {
	cID := ctx.Value(fiber.CtxComponentIDKey)
	cKind := ctx.Value(fiber.CtxComponentKindKey)

	if cKindCompType, ok := cKind.(fiber.ComponentKind); ok {
		// Only measure time taken for Caller
		if cKindCompType == fiber.CallerKind {
			if startTime, ok := ctx.Value(startTimeKey).(time.Time); ok {
				if routeName, ok := cID.(string); ok {
					for resp := range queue.Iter() {
						// Measure the time taken for the route
						labels := map[string]string{
							"status": metrics.GetStatusString(resp.IsSuccess()),
							"route":  routeName,
						}
						metrics.Glob().MeasureDurationMsSince(
							metrics.RouteRequestDurationMs,
							startTime,
							labels,
						)
					}
				}
			}
		}
	}
}

/////////////////////////// TracingInterceptor ////////////////////////////////

// NewTracingInterceptor is a creator for a TracingInterceptor
func NewTracingInterceptor() fiber.Interceptor {
	return &TracingInterceptor{}
}

// TracingInterceptor is the structural interceptor used for capturing
// run time metrics from the Fiber components
type TracingInterceptor struct {
	*fiber.NoopAfterDispatchInterceptor
}

// BeforeDispatch starts a new / child span and associates it with the context
func (i *TracingInterceptor) BeforeDispatch(
	ctx context.Context,
	req fiber.Request,
) context.Context {
	// Get the component id to be used as the operation name
	cID := ctx.Value(fiber.CtxComponentIDKey)
	// Create span and add to context
	_, ctx = tracing.Glob().StartSpanFromContext(ctx, cID.(string))
	return ctx
}

// AfterCompletion retrieves the span from the context, if exists, and finishes the trace
func (i *TracingInterceptor) AfterCompletion(
	ctx context.Context,
	req fiber.Request,
	queue fiber.ResponseQueue,
) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.Finish()
	}
}
