package models

import (
	"testing"

	"github.com/stretchr/testify/assert"

	routerConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"
)

func TestLogConfigValue(t *testing.T) {
	tests := map[string]struct {
		logConfig LogConfig
		expected  string
	}{
		"nop": {
			logConfig: LogConfig{
				LogLevel:             routerConfig.DebugLevel,
				CustomMetricsEnabled: true,
				FiberDebugLogEnabled: false,
				JaegerEnabled:        false,
				ResultLoggerType:     NopLogger,
			},
			expected: string(`{
				"log_level": "DEBUG",
				"custom_metrics_enabled": true,
				"fiber_debug_log_enabled": false,
				"jaeger_enabled": false,
				"result_logger_type": "nop"
			}`),
		},
		"bigquery": {
			logConfig: LogConfig{
				LogLevel:         routerConfig.InfoLevel,
				ResultLoggerType: BigQueryLogger,
				BigQueryConfig: &BigQueryConfig{
					Table:                "test-table",
					ServiceAccountSecret: "svc-acct-secret",
					BatchLoad:            true,
				},
			},
			expected: string(`{
				"log_level": "INFO",
				"custom_metrics_enabled": false,
				"fiber_debug_log_enabled": false,
				"jaeger_enabled": false,
				"result_logger_type": "bigquery",
				"bigquery_config": {
					"table": "test-table",
					"service_account_secret": "svc-acct-secret",
					"batch_load": true
				}
			}`),
		},
		"kafka": {
			logConfig: LogConfig{
				LogLevel:         routerConfig.WarnLevel,
				ResultLoggerType: KafkaLogger,
				KafkaConfig: &KafkaConfig{
					Brokers:             "test-brokers",
					Topic:               "test-topic",
					SerializationFormat: "test-serialization",
				},
			},
			expected: string(`{
				"log_level": "WARN",
				"custom_metrics_enabled": false,
				"fiber_debug_log_enabled": false,
				"jaeger_enabled": false,
				"result_logger_type": "kafka",
				"kafka_config": {
					"brokers": "test-brokers",
					"topic": "test-topic",
					"serialization_format": "test-serialization"
				}
			}`),
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			value, err := data.logConfig.Value()
			// Convert to string for comparison
			byteValue, ok := value.([]byte)
			assert.True(t, ok)
			// Validate
			assert.NoError(t, err)
			assert.JSONEq(t, data.expected, string(byteValue))
		})
	}
}

func TestLogConfigScan(t *testing.T) {
	tests := map[string]struct {
		value    interface{}
		success  bool
		expected LogConfig
		err      string
	}{
		"success": {
			value: []byte(`{
				"log_level": "DEBUG",
				"custom_metrics_enabled": true,
				"fiber_debug_log_enabled": false,
				"jaeger_enabled": false,
				"result_logger_type": "nop"
			}`),
			success: true,
			expected: LogConfig{
				LogLevel:             routerConfig.DebugLevel,
				CustomMetricsEnabled: true,
				FiberDebugLogEnabled: false,
				JaegerEnabled:        false,
				ResultLoggerType:     NopLogger,
			},
		},
		"failure | invalid value": {
			value:   100,
			success: false,
			err:     "type assertion to []byte failed",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			var logConfig LogConfig
			err := logConfig.Scan(data.value)
			if data.success {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, logConfig)
			} else {
				assert.Error(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}
