package metrics

import (
	"testing"
	"time"
)

// Nop methods return nothing and have no side effects.
// Simply exercise them to check that there are no panics.
func TestNopMethods(_ *testing.T) {
	testMetric := MetricName("TEST_METRIC")
	c := &NopMetricsCollector{}
	c.InitMetrics()
	c.MeasureDurationMs(testMetric, map[string]func() string{})
	c.MeasureDurationMsSince(testMetric, time.Now(), map[string]string{})
}
