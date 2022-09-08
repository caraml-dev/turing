package metrics

import (
	"time"

	"github.com/caraml-dev/turing/engines/router/missionctl/log"
)

//////////////////// Define Metric Name, label constants //////////////////////

// MetricName is a type used to define the names of the various metrics collected in the
// Turing App.
type MetricName string

// Define all metric names for the Turing App
const (
	// ExperimentEngineRequestMs is the key to measure requests for fetching a treatment from the experiment-engine
	ExperimentEngineRequestMs MetricName = "exp_engine_request_duration_ms"
	// RouteRequestDurationMs is the key to measure http requests to individual Fiber routes
	RouteRequestDurationMs MetricName = "route_request_duration_ms"
	// TuringComponentRequestDurationMs is the key to measure time taken at each Turing Component
	TuringComponentRequestDurationMs MetricName = "turing_comp_request_duration_ms"
)

var statusLabels = struct {
	Success string
	Failure string
}{
	Success: "success",
	Failure: "failure",
}

////////////////////////// Collector ///////////////////////////////////

// Collector defines the common interface for all metrics collection engines
type Collector interface {
	InitMetrics()
	MeasureDurationMsSince(key MetricName, starttime time.Time, labels map[string]string)
	// MeasureDurationMs is a deferrable version of MeasureDurationMsSince which evaluates labels
	// at the time of logging
	MeasureDurationMs(key MetricName, labels map[string]func() string) func()
}

// globalMetricsCollector is initialised to a Nop metrics collector. Calling
// InitMetricsCollector can update this value.
var globalMetricsCollector = newNopMetricsCollector()

// Glob returns the global metrics collector
func Glob() Collector {
	return globalMetricsCollector
}

// SetGlobMetricsCollector is used to update the global metrics collector instance with the input
func SetGlobMetricsCollector(c Collector) {
	globalMetricsCollector = c
}

// InitMetricsCollector is used to select the appropriate metrics collector and
// set up the required values for instrumenting.
func InitMetricsCollector(enabled bool) error {
	if enabled {
		log.Glob().Info("Initializing Prometheus Metrics Collector")
		// Use the Prometheus Instrumentation Client
		SetGlobMetricsCollector(&PrometheusClient{})
	} else {
		// Use the Nop Metrics collector
		log.Glob().Info("Initializing Nop Metrics Collector")
		SetGlobMetricsCollector(newNopMetricsCollector())
	}
	// Initialize
	globalMetricsCollector.InitMetrics()
	return nil
}

// GetStatusString returns a classification string (success / failure) based on the
// input boolean
func GetStatusString(status bool) string {
	if status {
		return statusLabels.Success
	}
	return statusLabels.Failure
}

// GetMeasureDurationFunc return the func that measures the duration of the request
func GetMeasureDurationFunc(err error, componentID string) func() {
	return Glob().MeasureDurationMs(
		TuringComponentRequestDurationMs,
		map[string]func() string{
			"status": func() string {
				return GetStatusString(err == nil)
			},
			"component": func() string {
				return componentID
			},
		},
	)
}
