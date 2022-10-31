package models

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"

	schema "github.com/caraml-dev/turing/api/turing/internal/alertschema"
)

// Alert contains policy that determines the "type of notification" and
// the "metric condition" that will trigger notifications for a
// particular service.
//
// There are 5 provided metrics that can be used as conditions for triggering the alert:
//
// - throughput: when current request per second is lower than the threshold
// - latency95p: when the 95-th percentile millisecond latency is higher than the threshold
// - error_rate: when the error percentage of all requests is higher than the threshold
// - cpu_util: when the percentage of cpu utilization is higher than the threshold
// - memory_util: when the percentage of memory utilization is higher than the threshold
//
// There are "warning" and "critical" thresholds that can be specified. A value of 0 or less
// will deactivate that particular type of alert. "warning" and "critical" will usually
// send alerts to different notifications channels.
//
// Environment, Team, Service, Metric and Duration are required in order to generate the correct
// query for the metric values and to direct the alert to the correct recipients.
type Alert struct {
	Model

	Environment       string  `json:"environment" validate:"required"`
	Team              string  `json:"team" validate:"required"`
	Service           string  `json:"service"`
	Metric            Metric  `json:"metric" validate:"oneof=throughput latency95p error_rate cpu_util memory_util"`
	WarningThreshold  float64 `json:"warning_threshold"`
	CriticalThreshold float64 `json:"critical_threshold"`
	// Duration to wait after the threshold is violated before firing the alert.
	// The duration format is a sequence of decimal numbers followed by a time unit suffix.
	// For instance: 5m
	Duration string `json:"duration" validate:"required"`
}

type Metric string

const (
	MetricThroughput Metric = "throughput"
	MetricLatency95p Metric = "latency95p"
	MetricErrorRate  Metric = "error_rate"
	MetricCPUUtil    Metric = "cpu_util"
	MetricMemoryUtil Metric = "memory_util"
)

func (alert Alert) Validate() error {
	validate := validator.New()
	if err := validate.Struct(alert); err != nil {
		return err
	}
	if _, err := time.ParseDuration(alert.Duration); err != nil {
		return err
	}
	return nil
}

// Group creates an AlertGroup from the Alert specification. An alert group follows the alerting rule format
// in prometheus: https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/
func (alert Alert) Group(playbookURL string, dashboardURL string) schema.AlertGroup {
	groupName := fmt.Sprintf("%s_%s_%s_%s", alert.Environment, alert.Team, alert.Service, alert.Metric)
	alerts := make([]schema.Alert, 0)

	if alert.WarningThreshold > 0 {
		alerts = append(alerts, schema.Alert{
			AlertName:  fmt.Sprintf("%s_%s_violation_%s", alert.Service, alert.Metric, alert.Environment),
			Expression: getAlertExpr(alert.Metric, alert.Environment, alert.Service, alert.WarningThreshold),
			For:        alert.Duration,
			Labels: map[string]string{
				"owner":        alert.Team,
				"service_name": alert.Service,
				"severity":     "warning",
			},
			Annotations: map[string]string{
				"summary": fmt.Sprintf("%s is %s than the threshold: %.0f%s",
					alert.Metric,
					getAlertOperatorText(alert.Metric),
					alert.WarningThreshold,
					getAlertUnit(alert.Metric),
				),
				"description": fmt.Sprintf("%s for the past %s: {{ $value }}%s",
					alert.Metric,
					alert.Duration,
					getAlertUnit(alert.Metric),
				),
				"dashboard": dashboardURL,
				"playbook":  playbookURL,
			},
		})
	}

	if alert.CriticalThreshold > 0 {
		alerts = append(alerts, schema.Alert{
			AlertName:  fmt.Sprintf("%s_%s_violation_%s", alert.Service, alert.Metric, alert.Environment),
			Expression: getAlertExpr(alert.Metric, alert.Environment, alert.Service, alert.CriticalThreshold),
			For:        alert.Duration,
			Labels: map[string]string{
				"owner":        alert.Team,
				"service_name": alert.Service,
				"severity":     "critical",
			},
			Annotations: map[string]string{
				"summary": fmt.Sprintf("%s is %s than the threshold: %.0f%s",
					alert.Metric,
					getAlertOperatorText(alert.Metric),
					alert.CriticalThreshold,
					getAlertUnit(alert.Metric),
				),
				"description": fmt.Sprintf("%s for the past %s: {{ $value }}%s",
					alert.Metric,
					alert.Duration,
					getAlertUnit(alert.Metric),
				),
				"dashboard": dashboardURL,
				"playbook":  playbookURL,
			},
		})
	}

	return schema.AlertGroup{
		Name:   groupName,
		Alerts: alerts,
	}
}

func getAlertExpr(metric Metric, env string, rev string, threshold float64) string {
	switch metric {
	case MetricThroughput:
		return fmt.Sprintf(`sum(rate(revision_request_count{
  environment="%s",
  service_name=~"%s-[0-9]*"
}[1m])) %s %f
`, env, rev, getAlertOperator(metric), threshold)

	case MetricLatency95p:
		return fmt.Sprintf(`histogram_quantile(0.95, sum(rate(revision_request_latencies_bucket{
  environment="%s",
  service_name=~"%s-[0-9]*"
}[1m])) by (le)) %s %f
`, env, rev, getAlertOperator(metric), threshold)

	case MetricErrorRate:
		return fmt.Sprintf(`sum(rate(revision_request_count{
  environment="%s",
  service_name=~"%s-[0-9]*",
  response_code_class!="2xx"
}[1m]))
/
sum(rate(revision_request_count{
  environment="%s",
  service_name=~"%s-[0-9]*"
}[1m])) * 100.0 %s %f
`, env, rev, env, rev, getAlertOperator(metric), threshold)

	case MetricCPUUtil:
		return fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{
  environment="%s",
  pod=~"%s-[0-9]*.*"
}[1m]))
/
sum(kube_pod_container_resource_requests{
  resource="cpu",
  environment="%s",
  pod=~"%s-[0-9]*.*"
}) * 100.0 %s %f
`, env, rev, env, rev, getAlertOperator(metric), threshold)

	case MetricMemoryUtil:
		return fmt.Sprintf(`sum(container_memory_usage_bytes{
  environment="%s",
  pod=~"%s-[0-9]*.*",
  image!=""
})
/
sum(kube_pod_container_resource_requests{
  resource="memory",
  environment="%s",
  pod=~"%s-[0-9]*.*"
}) * 100.0 %s %f
`, env, rev, env, rev, getAlertOperator(metric), threshold)

	default:
		return ""
	}
}

func getAlertOperator(metric Metric) string {
	if metric == MetricThroughput {
		return "<"
	}
	return ">"
}

func getAlertOperatorText(metric Metric) string {
	if metric == MetricThroughput {
		return "lower"
	}
	return "higher"
}

func getAlertUnit(metric Metric) string {
	switch metric {
	case MetricThroughput:
		return "rps"
	case MetricLatency95p:
		return "ms"
	case MetricErrorRate:
		fallthrough
	case MetricCPUUtil:
		fallthrough
	case MetricMemoryUtil:
		return "%"
	}
	return ""
}
