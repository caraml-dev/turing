package experiment

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/turing/engines/router/missionctl/instrumentation/metrics"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
)

// mockMetricsCollector satisfies the metrics.Collector interface
type mockMetricsCollector struct {
	mock.Mock
}

func (*mockMetricsCollector) InitMetrics() {}
func (c *mockMetricsCollector) MeasureDurationMsSince(
	key metrics.MetricName,
	starttime time.Time,
	labels map[string]string,
) {
	c.Called(key, starttime, labels)
}
func (*mockMetricsCollector) MeasureDurationMs(
	key metrics.MetricName,
	labels map[string]func() string,
) func() {
	return func() {}
}

func TestMetricsInterceptorBeforeDispatch(t *testing.T) {
	i := MetricsInterceptor{}
	ctx := i.BeforeDispatch(context.Background())
	// Verify that the start time has been set
	if _, ok := ctx.Value(startTimeKey).(time.Time); !ok {
		tu.FailOnError(t, fmt.Errorf("Start time has not been set in the context"))
	}
}

func TestMetricsInterceptorAfterCompletion(t *testing.T) {
	// Make instrumentation interceptor
	i := MetricsInterceptor{}

	// Make test ctx
	starttime := time.Now()
	ctx := context.WithValue(context.Background(), startTimeKey, starttime)

	// Patch global metrics collector and run AfterCompletion
	mc := &mockMetricsCollector{}
	mc.On("MeasureDurationMsSince",
		mock.Anything, mock.Anything, mock.Anything,
	).Return(nil)
	globMC := metrics.Glob()
	metrics.SetGlobMetricsCollector(mc)
	i.AfterCompletion(ctx, nil)
	metrics.SetGlobMetricsCollector(globMC)

	// Verify invocation of MeasureDurationMsSince
	mc.AssertCalled(t,
		"MeasureDurationMsSince",
		mock.Anything, mock.Anything, mock.Anything,
	)
}
