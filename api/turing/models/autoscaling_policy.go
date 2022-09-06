package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

type AutoscalingMetric string

const (
	AutoscalingMetricConcurrency AutoscalingMetric = "concurrency"
	AutoscalingMetricRPS         AutoscalingMetric = "rps"
	AutoscalingMetricCPU         AutoscalingMetric = "cpu"
	AutoscalingMetricMemory      AutoscalingMetric = "memory"
)

type AutoscalingPolicy struct {
	Metric AutoscalingMetric `json:"metric" validate:"required,oneof=concurrency rps cpu memory"`
	Target string            `json:"target" validate:"required,number"`
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
