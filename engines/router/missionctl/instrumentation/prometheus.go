package instrumentation

import (
	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
	"github.com/golang-collections/collections/set"
	"github.com/prometheus/client_golang/prometheus"
)

//////////////////////////// Metrics Definitions //////////////////////////////

const (
	// Namespace is the Prometheus Namespace in all metrics published by the turing app
	Namespace string = "mlp"
	// Subsystem is the Prometheus Subsystem in all metrics published by the turing app
	Subsystem string = "turing"
)

// Define all metric names for the Turing App
const (
	// ExperimentEngineRequestMs is the key to measure requests for fetching a treatment from the experiment-engine
	ExperimentEngineRequestMs metrics.MetricName = "exp_engine_request_duration_ms"
	// RouteRequestDurationMs is the key to measure http requests to individual Fiber routes
	RouteRequestDurationMs metrics.MetricName = "route_request_duration_ms"
	// TuringComponentRequestDurationMs is the key to measure time taken at each Turing Component
	TuringComponentRequestDurationMs metrics.MetricName = "turing_comp_request_duration_ms"
)

// requestLatencyBuckets defines the buckets used in the custom Histogram metrics defined by Turing
var requestLatencyBuckets = []float64{
	2, 4, 6, 8, 10, 15, 20, 30, 40, 50, 60, 70, 80, 90, 100, 120, 140, 160, 180, 200,
	250, 300, 350, 400, 450, 500, 550, 600, 650, 700, 750, 800, 850, 900, 950, 1000,
	2000, 5000, 10000, 20000, 50000, 100000,
}

// additionalRegisteredMetricNames is a set containing all registered experiment engine metric names to prevent
// re-registrations if the RegisterMetrics method is called multiple times on the same metrics (happens when there are
// multiple fiber routes using the same experimentation policy)
var additionalRegisteredMetricNames = set.New(nil)

func GetHistogramMap() map[metrics.MetricName]metrics.PrometheusHistogramVec {
	// histogramMap maintains a mapping between the metric name and the corresponding histogram vector
	var histogramMap = map[metrics.MetricName]metrics.PrometheusHistogramVec{
		ExperimentEngineRequestMs: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      string(ExperimentEngineRequestMs),
			Help:      "Histogram for the runtime (in milliseconds) of Experiment Engine requests.",
			Buckets:   requestLatencyBuckets,
		},
			[]string{"status", "engine"},
		),
		RouteRequestDurationMs: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      string(RouteRequestDurationMs),
			Help:      "Histogram for the runtime (in milliseconds) of Fiber route requests.",
			Buckets:   requestLatencyBuckets,
		},
			[]string{"status", "route", "traffic_rule"},
		),
		TuringComponentRequestDurationMs: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: Namespace,
			Subsystem: Subsystem,
			Name:      string(TuringComponentRequestDurationMs),
			Help:      "Histogram for time spent (in milliseconds) at each Turing component.",
			Buckets:   requestLatencyBuckets,
		},
			[]string{"status", "component", "traffic_rule"},
		),
	}

	return histogramMap
}

//////////////////////////// MetricsRegistrationHelper Definitions //////////////////////////////

type MetricType string

const (
	GaugeMetricType     MetricType = "gauge"
	HistogramMetricType MetricType = "histogram"
	CounterMetricType   MetricType = "counter"
)

type Metric struct {
	Name        string
	Type        MetricType
	Description string
	Labels      []string
	Buckets     []float64
}

type MetricsRegistrationHelper struct{}

func (MetricsRegistrationHelper) Register(additionalMetrics []Metric) error {
	gaugeMap := map[metrics.MetricName]metrics.PrometheusGaugeVec{}
	histogramMap := map[metrics.MetricName]metrics.PrometheusHistogramVec{}
	counterMap := map[metrics.MetricName]metrics.PrometheusCounterVec{}

	for _, metric := range additionalMetrics {
		if !additionalRegisteredMetricNames.Has(metric.Name) {
			switch metric.Type {
			case GaugeMetricType:
				gaugeMap[metrics.MetricName(metric.Name)] = prometheus.NewGaugeVec(prometheus.GaugeOpts{
					Namespace: Namespace,
					Subsystem: Subsystem,
					Help:      metric.Description,
					Name:      metric.Name,
				},
					metric.Labels,
				)
			case HistogramMetricType:
				buckets := requestLatencyBuckets
				if metric.Buckets != nil {
					buckets = metric.Buckets
				}
				histogramMap[metrics.MetricName(metric.Name)] = prometheus.NewHistogramVec(prometheus.HistogramOpts{
					Namespace: Namespace,
					Subsystem: Subsystem,
					Help:      metric.Description,
					Name:      metric.Name,
					Buckets:   buckets,
				},
					metric.Labels,
				)
			case CounterMetricType:
				counterMap[metrics.MetricName(metric.Name)] = prometheus.NewCounterVec(prometheus.CounterOpts{
					Namespace: Namespace,
					Subsystem: Subsystem,
					Help:      metric.Description,
					Name:      metric.Name,
				},
					metric.Labels,
				)
			}
			additionalRegisteredMetricNames.Insert(metric.Name)
		}
	}

	err := metrics.Glob().(*metrics.PrometheusClient).RegisterMetrics(
		gaugeMap,
		histogramMap,
		counterMap,
	)
	if err != nil {
		return err
	}
	return nil
}
