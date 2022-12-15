package metrics

import (
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
)

//////////////////// Define Metric Name, label constants //////////////////////

// MetricName is a type used to define the names of the various metrics collected in the
// Turing App.
type MetricName string

// Define all metric names for the Turing App
const (
	// ExperimentEngineRequestMs is the key to measure requests for fetching a treatment from the experiment-engine
	ExperimentEngineRequestMs metrics.MetricName = "exp_engine_request_duration_ms"
	// RouteRequestDurationMs is the key to measure http requests to individual Fiber routes
	RouteRequestDurationMs metrics.MetricName = "route_request_duration_ms"
	// TuringComponentRequestDurationMs is the key to measure time taken at each Turing Component
	TuringComponentRequestDurationMs metrics.MetricName = "turing_comp_request_duration_ms"
)

var statusLabels = struct {
	Success string
	Failure string
}{
	Success: "success",
	Failure: "failure",
}

// globalMetricsCollector is initialised to a Nop metrics collector. Calling
// InitMetricsCollector can update this value.
var globalMetricsCollector metrics.Collector = nil

// Glob returns the global metrics collector
func Glob() metrics.Collector {
	return globalMetricsCollector
}

// SetGlobMetricsCollector is used to update the global metrics collector instance with the input
func SetGlobMetricsCollector(c metrics.Collector) {
	globalMetricsCollector = c
}

// InitMetricsCollector is used to select the appropriate metrics collector and
// set up the required values for instrumenting.
func InitMetricsCollector(enabled bool) error {
	if enabled {
		log.Glob().Info("Initializing Prometheus Metrics Collector")
		// Use the Prometheus Instrumentation Client
		err := metrics.InitPrometheusMetricsCollector(
			map[metrics.MetricName]metrics.PrometheusGaugeVec{},
			histogramMap,
			map[metrics.MetricName]metrics.PrometheusCounterVec{},
		)
		if err != nil {
			return err
		}
		SetGlobMetricsCollector(metrics.Glob())
	}
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
