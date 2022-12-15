package metrics

import (
	"fmt"
	"testing"

	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
	"github.com/stretchr/testify/assert"
)

func TestInitMetricsCollectorPrometheus(t *testing.T) {
	err := InitMetricsCollector(true)
	// Validate
	assert.NoError(t, err)
	if _, ok := metrics.Glob().(*metrics.PrometheusClient); !ok {
		err := fmt.Errorf("Prometheus metrics collector was not initialised")
		tu.FailOnError(t, err)
	}
}
