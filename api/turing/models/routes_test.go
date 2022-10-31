package models

import (
	"testing"
	"time"

	fiberConfig "github.com/gojek/fiber/config"
	fiberProtocol "github.com/gojek/fiber/protocol"
	"github.com/stretchr/testify/assert"
)

var testHTTPRoutes = Routes{
	{
		ID:       "test-id",
		Type:     "PROXY",
		Endpoint: "test-endpoint",
		Annotations: map[string]string{
			"merlin.gojek.com/model-id": "10",
		},
		Timeout: "2s",
	},
}

var testGRPCRoutes = Routes{
	{
		ID:       "test-id",
		Type:     "PROXY",
		Endpoint: "test-endpoint",
		Annotations: map[string]string{
			"merlin.gojek.com/model-id": "10",
		},
		Timeout:       "2s",
		ServiceMethod: "package/method",
	},
}

func TestRoutesValue(t *testing.T) {
	value, err := testGRPCRoutes.Value()
	// Convert to string for comparison
	byteValue, ok := value.([]byte)
	assert.True(t, ok)
	// Validate
	assert.NoError(t, err)
	assert.JSONEq(t, `
		[{
			"id": "test-id",
			"type": "PROXY",
			"endpoint": "test-endpoint",
			"service_method": "package/method",
			"annotations": {
				"merlin.gojek.com/model-id": "10"
			},
			"timeout": "2s"
		}]
	`, string(byteValue))
}

func TestRoutesScan(t *testing.T) {
	tests := map[string]struct {
		value    interface{}
		success  bool
		expected Routes
		err      string
	}{
		"success": {
			value: []byte(`
				[{
					"id": "test-id",
					"type": "PROXY",
					"endpoint": "test-endpoint",
					"service_method": "package/method",
					"annotations": {
						"merlin.gojek.com/model-id": "10"
					},
					"timeout": "2s"
				}]
			`),
			success:  true,
			expected: testGRPCRoutes,
		},
		"failure | invalid value": {
			value:   100,
			success: false,
			err:     "type assertion to []byte failed",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			var routes Routes
			err := routes.Scan(data.value)
			if data.success {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, routes)
			} else {
				assert.Error(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}

func TestRoutesToFiberRoutes(t *testing.T) {
	tests := map[string]struct {
		routes      Routes
		fiberRoutes fiberConfig.Routes
		protocol    fiberProtocol.Protocol
		success     bool
		err         string
	}{
		"success http": {
			routes:   testHTTPRoutes,
			protocol: fiberProtocol.HTTP,
			fiberRoutes: fiberConfig.Routes{
				&fiberConfig.ProxyConfig{
					ComponentConfig: fiberConfig.ComponentConfig{
						ID:   "test-id",
						Type: "PROXY",
					},
					Endpoint: "test-endpoint",
					Timeout:  fiberConfig.Duration(time.Second * 2),
					Protocol: fiberProtocol.HTTP,
				},
			},
			success: true,
		},
		"success grpc": {
			routes:   testGRPCRoutes,
			protocol: fiberProtocol.GRPC,
			fiberRoutes: fiberConfig.Routes{
				&fiberConfig.ProxyConfig{
					ComponentConfig: fiberConfig.ComponentConfig{
						ID:   "test-id",
						Type: "PROXY",
					},
					Endpoint: "test-endpoint",
					Timeout:  fiberConfig.Duration(time.Second * 2),
					Protocol: fiberProtocol.GRPC,
					GrpcConfig: fiberConfig.GrpcConfig{
						ServiceMethod: "package/method",
					},
				},
			},
			success: true,
		},
		"failure | bad timeout": {
			routes: Routes{
				{
					ID:       "test-id",
					Type:     "PROXY",
					Endpoint: "test-endpoint",
					Timeout:  "2t",
				},
			},
			protocol: fiberProtocol.HTTP,
			success:  false,
			err:      "time: unknown unit",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			fiberRoutes, err := data.routes.ToFiberRoutes(data.protocol)
			if data.success {
				assert.NoError(t, err)
				assert.Equal(t, data.fiberRoutes, *fiberRoutes)
			} else {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), data.err)
				}
			}
		})
	}
}
