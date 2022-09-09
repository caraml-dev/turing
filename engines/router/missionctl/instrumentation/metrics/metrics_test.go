package metrics

import (
	"fmt"
	"testing"

	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
	promtestutil "github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStatusString(t *testing.T) {
	assert.Equal(t, "success", GetStatusString(true))
	assert.Equal(t, "failure", GetStatusString(false))
}

func TestInitMetricsCollectorNop(t *testing.T) {
	err := InitMetricsCollector(false)
	// Validate
	assert.NoError(t, err)
	if _, ok := Glob().(*NopMetricsCollector); !ok {
		err := fmt.Errorf("Nop metrics collector was not initialised")
		tu.FailOnError(t, err)
	}
}

func TestInitMetricsCollectorPrometheus(t *testing.T) {
	err := InitMetricsCollector(true)
	// Validate
	assert.NoError(t, err)
	if _, ok := Glob().(*PrometheusClient); !ok {
		err := fmt.Errorf("Prometheus metrics collector was not initialised")
		tu.FailOnError(t, err)
	}
}

func TestGetMeasureDurationFunc(t *testing.T) {
	histoVec := histogramMap[TuringComponentRequestDurationMs]
	require.Zero(t, promtestutil.CollectAndCount(histoVec))

	err := InitMetricsCollector(true)
	assert.NoError(t, err)
	GetMeasureDurationFunc(err, "test")()
	require.Equal(t, 1, promtestutil.CollectAndCount(histoVec))
}
