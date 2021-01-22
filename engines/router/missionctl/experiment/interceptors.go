package experiment

import (
	"context"
	"time"

	"github.com/gojek/turing/engines/experiment/runner"
	"github.com/gojek/turing/engines/router/missionctl/instrumentation/metrics"
)

type ctxKey string

const (
	startTimeKey ctxKey = "startTimeKey"
)

// MetricsInterceptor is the structural interceptor used for capturing metrics
// from experiment runs
type MetricsInterceptor struct{}

// NewMetricsInterceptor is a creator for a MetricsInterceptor
func NewMetricsInterceptor() runner.Interceptor {
	return &MetricsInterceptor{}
}

// BeforeDispatch associates the start time to the context
func (i *MetricsInterceptor) BeforeDispatch(
	ctx context.Context,
) context.Context {
	return context.WithValue(ctx, startTimeKey, time.Now())
}

// AfterCompletion logs the time taken for the component to process the request,
// to the metrics collector
func (i *MetricsInterceptor) AfterCompletion(
	ctx context.Context,
	err error,
) {
	labels := map[string]string{
		"status":     metrics.GetStatusString(err == nil),
		"engine":     "",
		"experiment": "",
		"treatment":  "",
	}

	if engine, ok := ctx.Value(runner.ExperimentEngineKey).(string); ok {
		labels["engine"] = engine
	}
	if experiment, ok := ctx.Value(runner.ExperimentNameKey).(string); ok {
		labels["experiment"] = experiment
	}
	if treatment, ok := ctx.Value(runner.TreatmentNameKey).(string); ok {
		labels["treatment"] = treatment
	}

	// Get start time
	if startTime, ok := ctx.Value(startTimeKey).(time.Time); ok {
		// Measure the time taken for the experiment run
		metrics.Glob().MeasureDurationMsSince(
			metrics.ExperimentEngineRequestMs,
			startTime,
			labels,
		)
	}
}
