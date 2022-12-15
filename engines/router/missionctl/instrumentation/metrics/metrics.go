package metrics

import (
	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation"
	"github.com/caraml-dev/turing/engines/router/missionctl/log"
	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
)

// InitMetricsCollector is used to select the appropriate metrics collector and
// set up the required values for instrumenting.
func InitMetricsCollector(enabled bool) error {
	if enabled {
		log.Glob().Info("Initializing Prometheus Metrics Collector")
		// Use the Prometheus Instrumentation Client
		err := metrics.InitPrometheusMetricsCollector(
			map[metrics.MetricName]metrics.PrometheusGaugeVec{},
			instrumentation.GetHistogramMap(),
			map[metrics.MetricName]metrics.PrometheusCounterVec{},
		)
		if err != nil {
			return err
		}
	}
	return nil
}
