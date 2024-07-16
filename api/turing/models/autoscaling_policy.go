package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
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
	// The 'numeric' type is used to allow decimals in strings to be set as the target value,
	// e.g. "8.88". See https://github.com/go-playground/validator/issues/940 for more details.
	Target string `json:"target" validate:"required,numeric"`
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
