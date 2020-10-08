package metrics

import "time"

// NopMetricsCollector implements the Collector interface with a set of
// Nop methods
type NopMetricsCollector struct {
}

func newNopMetricsCollector() Collector {
	return &NopMetricsCollector{}
}

// InitMetrics satisfies the Collector interface
func (NopMetricsCollector) InitMetrics() {}

// MeasureDurationMsSince satisfies the Collector interface
func (NopMetricsCollector) MeasureDurationMsSince(MetricName, time.Time, map[string]string) {}

// MeasureDurationMs satisfies the Collector interface
func (NopMetricsCollector) MeasureDurationMs(MetricName, map[string]func() string) func() {
	return func() {}
}
