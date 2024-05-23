package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestResourceRequestValue(t *testing.T) {
	resouceReq := ResourceRequest{
		MinReplica:    1,
		MaxReplica:    2,
		CPURequest:    resource.MustParse("500m"),
		MemoryRequest: resource.MustParse("1Gi"),
	}

	// Validate
	value, err := resouceReq.Value()
	// Convert to string for comparison
	byteValue, ok := value.([]byte)
	assert.True(t, ok)
	// Validate
	assert.NoError(t, err)
	assert.JSONEq(t, `
		{
			"min_replica": 1,
			"max_replica": 2,
			"cpu_request": "500m",
			"cpu_limit":"0",
			"memory_request": "1Gi"
		}
	`, string(byteValue))
}

func TestResourceRequestScan(t *testing.T) {
	tests := map[string]struct {
		value    interface{}
		success  bool
		expected ResourceRequest
		err      string
	}{
		"success": {
			value: []byte(`{
				"min_replica": 1,
				"max_replica": 2,
				"cpu_request": "500m",
				"memory_request": "1Gi"
			}`),
			success: true,
			expected: ResourceRequest{
				MinReplica:    1,
				MaxReplica:    2,
				CPURequest:    resource.MustParse("500m"),
				MemoryRequest: resource.MustParse("1Gi"),
			},
		},
		"failure | invalid value": {
			value:   100,
			success: false,
			err:     "type assertion to []byte failed",
		},
		"failure | invalid resource value": {
			value: []byte(`{
				"min_replica": 1,
				"max_replica": 2,
				"cpu_request": "5x",
				"memory_request": "1Gi"
			}`),
			success: false,
			err:     "quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'",
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			var resourceReq ResourceRequest
			err := resourceReq.Scan(data.value)
			if data.success {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, resourceReq)
			} else {
				assert.Error(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}
