package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Type to represent the autoscaling metrics supported by Knative.
// Ref: https://pkg.go.dev/knative.dev/serving/pkg/apis/autoscaling
//      https://github.com/knative/serving/pull/11668

type AutoscalingMetric string

const (
	AutoscalingMetricConcurrency AutoscalingMetric = "concurrency"
	AutoscalingMetricRPS         AutoscalingMetric = "rps"
	AutoscalingMetricCPU         AutoscalingMetric = "cpu"
	AutoscalingMetricMemory      AutoscalingMetric = "memory"
)

type AutoscalingPolicy struct {
	Metric AutoscalingMetric `json:"metric" validate:"required,oneof=concurrency rps cpu memory"`
	// Target is the target value of the metric that should be reached to add a new replica.
	// It is expected that the autoscaling target is an absolute value for concurrency / rps
	// while it is a % value (of the requested value) for cpu / memory.
	Target string `json:"target" validate:"required,number"`
	// DefaultAutoscalingPolicyPayloadSize indicates the request payload
	DefaultAutoscalingPolicyPayloadSize *resource.Quantity `json:"default_autoscaling_policy_payload_size"`
}

func (a AutoscalingPolicy) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *AutoscalingPolicy) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}
