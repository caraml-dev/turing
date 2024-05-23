package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"k8s.io/apimachinery/pkg/api/resource"
)

type ResourceRequest struct {
	// Minimum number of replica of inference service
	MinReplica int `json:"min_replica"`
	// Maximum number of replica of inference service
	MaxReplica int `json:"max_replica"`

	// CPU request of inference service
	CPURequest resource.Quantity `json:"cpu_request"`
	// CPU limit of inference service
	CPULimit resource.Quantity `json:"cpu_limit"`
	// Memory request of inference service
	MemoryRequest resource.Quantity `json:"memory_request"`
}

func (r ResourceRequest) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *ResourceRequest) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &r)
}
