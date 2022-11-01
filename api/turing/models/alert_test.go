package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/caraml-dev/turing/api/turing/internal/alertschema"
)

func TestValidateAlert(t *testing.T) {
	tests := map[string]struct {
		alert   Alert
		success bool
		err     string
	}{
		"success": {
			alert: Alert{
				Environment: "test-env",
				Team:        "test-team",
				Service:     "test-svc",
				Metric:      "throughput",
				Duration:    "1m",
			},
			success: true,
		},
		"failure | missing required field": {
			alert: Alert{
				Team:     "test-team",
				Service:  "test-svc",
				Metric:   "latency95p",
				Duration: "1m",
			},
			success: false,
			err:     "Key: 'Alert.Environment' Error:Field validation for 'Environment' failed on the 'required' tag",
		},
		"failure | invalid duration": {
			alert: Alert{
				Environment: "test-env",
				Team:        "test-team",
				Service:     "test-svc",
				Metric:      "error_rate",
				Duration:    "1tt",
			},
			success: false,
			err:     "time: unknown unit",
		},
		"failure | invalid metric": {
			alert: Alert{
				Environment: "test-env",
				Team:        "test-team",
				Service:     "test-svc",
				Metric:      "test-metric",
				Duration:    "1s",
			},
			success: false,
			err:     "Key: 'Alert.Metric' Error:Field validation for 'Metric' failed on the 'oneof' tag",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			err := data.alert.Validate()
			if data.success {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), data.err)
				}
			}
		})
	}
}

func TestAlertGroup(t *testing.T) {
	alert := Alert{
		Environment:       "test-env",
		Team:              "test-team",
		Service:           "test-svc",
		Metric:            "cpu_util",
		Duration:          "1m",
		WarningThreshold:  0.95,
		CriticalThreshold: 0.90,
	}

	expected := schema.AlertGroup{
		Name: "test-env_test-team_test-svc_cpu_util",
		Alerts: []schema.Alert{
			{
				AlertName: "test-svc_cpu_util_violation_test-env",
				Expression: strings.Join([]string{"sum(rate(container_cpu_usage_seconds_total{",
					"  environment=\"test-env\",",
					"  pod=~\"test-svc-[0-9]*.*\"\n}[1m]))\n/",
					"sum(kube_pod_container_resource_requests{",
					"  resource=\"cpu\",",
					"  environment=\"test-env\",",
					"  pod=~\"test-svc-[0-9]*.*\"\n}) * 100.0 > 0.950000\n"}, "\n"),
				For: "1m",
				Labels: map[string]string{
					"owner":        "test-team",
					"service_name": "test-svc",
					"severity":     "warning",
				},
				Annotations: map[string]string{
					"dashboard":   "https://example.com/dashboard",
					"description": "cpu_util for the past 1m: {{ $value }}%",
					"playbook":    "https://example.com",
					"summary":     "cpu_util is higher than the threshold: 1%",
				},
			},
			{
				AlertName: "test-svc_cpu_util_violation_test-env",
				Expression: strings.Join([]string{"sum(rate(container_cpu_usage_seconds_total{",
					"  environment=\"test-env\",",
					"  pod=~\"test-svc-[0-9]*.*\"\n}[1m]))\n/",
					"sum(kube_pod_container_resource_requests{",
					"  resource=\"cpu\",",
					"  environment=\"test-env\",",
					"  pod=~\"test-svc-[0-9]*.*\"\n}) * 100.0 > 0.900000\n"}, "\n"),
				For: "1m",
				Labels: map[string]string{
					"owner":        "test-team",
					"service_name": "test-svc",
					"severity":     "critical",
				},
				Annotations: map[string]string{
					"dashboard":   "https://example.com/dashboard",
					"description": "cpu_util for the past 1m: {{ $value }}%",
					"playbook":    "https://example.com",
					"summary":     "cpu_util is higher than the threshold: 1%",
				},
			},
		},
	}

	assert.Equal(t, expected, alert.Group("https://example.com", "https://example.com/dashboard"))
}

func TestGetAlertExpr(t *testing.T) {
	tests := map[string]struct {
		metric       Metric
		env          string
		rev          string
		threshold    float64
		expectedExpr string
	}{
		"throughput": {
			metric:    MetricThroughput,
			env:       "test-env",
			rev:       "test-rev",
			threshold: 0.90,
			expectedExpr: strings.Join([]string{"sum(rate(revision_request_count{",
				"  environment=\"test-env\",",
				"  service_name=~\"test-rev-[0-9]*\"", "}[1m])) < 0.900000\n",
			}, "\n"),
		},
		"latency": {
			metric:    MetricLatency95p,
			env:       "test-env",
			rev:       "test-rev",
			threshold: 0.95,
			expectedExpr: strings.Join([]string{
				"histogram_quantile(0.95, sum(rate(revision_request_latencies_bucket{",
				"  environment=\"test-env\",",
				"  service_name=~\"test-rev-[0-9]*\"",
				"}[1m])) by (le)) > 0.950000\n",
			}, "\n"),
		},
		"error_rate": {
			metric:    MetricErrorRate,
			env:       "test-env",
			rev:       "test-rev",
			threshold: 0.80,
			expectedExpr: strings.Join([]string{
				"sum(rate(revision_request_count{",
				"  environment=\"test-env\",",
				"  service_name=~\"test-rev-[0-9]*\",",
				"  response_code_class!=\"2xx\"",
				"}[1m]))\n/\nsum(rate(revision_request_count{",
				"  environment=\"test-env\",",
				"  service_name=~\"test-rev-[0-9]*\"",
				"}[1m])) * 100.0 > 0.800000\n",
			}, "\n"),
		},
		"cpu_util": {
			metric:    MetricCPUUtil,
			env:       "test-env",
			rev:       "test-rev",
			threshold: 0.5,
			expectedExpr: strings.Join([]string{
				"sum(rate(container_cpu_usage_seconds_total{",
				"  environment=\"test-env\",",
				"  pod=~\"test-rev-[0-9]*.*\"",
				"}[1m]))\n/\nsum(kube_pod_container_resource_requests{",
				"  resource=\"cpu\",",
				"  environment=\"test-env\",",
				"  pod=~\"test-rev-[0-9]*.*\"",
				"}) * 100.0 > 0.500000\n",
			}, "\n"),
		},
		"memory_util": {
			metric:    MetricMemoryUtil,
			env:       "test-env",
			rev:       "test-rev",
			threshold: 0.55,
			expectedExpr: strings.Join([]string{
				"sum(container_memory_usage_bytes{",
				"  environment=\"test-env\",",
				"  pod=~\"test-rev-[0-9]*.*\",",
				"  image!=\"\"",
				"})\n/\nsum(kube_pod_container_resource_requests{",
				"  resource=\"memory\",",
				"  environment=\"test-env\",",
				"  pod=~\"test-rev-[0-9]*.*\"",
				"}) * 100.0 > 0.550000\n",
			}, "\n"),
		},
		"default": {
			metric:    Metric("test"),
			env:       "test-env",
			rev:       "test-rev",
			threshold: 0.90,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			expr := getAlertExpr(data.metric, data.env, data.rev, data.threshold)
			assert.Equal(t, data.expectedExpr, expr)
		})
	}
}

func TestGetAlertUnit(t *testing.T) {
	assert.Equal(t, "%", getAlertUnit(MetricErrorRate))
	assert.Equal(t, "%", getAlertUnit(MetricCPUUtil))
	assert.Equal(t, "%", getAlertUnit(MetricMemoryUtil))
	assert.Equal(t, "rps", getAlertUnit(MetricThroughput))
	assert.Equal(t, "ms", getAlertUnit(MetricLatency95p))
	assert.Equal(t, "", getAlertUnit(Metric("")))
}

func TestGetAlertOperatorText(t *testing.T) {
	assert.Equal(t, "lower", getAlertOperatorText(MetricThroughput))
	assert.Equal(t, "higher", getAlertOperatorText(Metric("")))
}
