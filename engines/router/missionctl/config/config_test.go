package config

import (
	"os"
	"testing"
	"time"

	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/stretchr/testify/assert"

	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
)

type testSuiteLogLevel struct {
	value   string
	result  LogLevel
	success bool
}

type testSuiteResultLogger struct {
	value   string
	result  ResultLogger
	success bool
}

type testSuiteSerializationFormat struct {
	value   string
	result  SerializationFormat
	success bool
}

var requiredEnvs = map[string]string{
	"PORT":               "8080",
	"ROUTER_CONFIG_FILE": "/var/test.yaml",
	"APP_NAME":           "turing",
	"APP_ENVIRONMENT":    "dev",
}

var optionalEnvs = map[string]string{
	"ENRICHER_ENDPOINT":              "http://localhost:8081",
	"ENRICHER_TIMEOUT":               "5ms",
	"ENSEMBLER_ENDPOINT":             "http://localhost:8082",
	"ENSEMBLER_TIMEOUT":              "2ms",
	"ROUTER_TIMEOUT":                 "10ms",
	"ROUTER_PROTOCOL":                "UPI_V1",
	"APP_LOGLEVEL":                   "DEBUG",
	"APP_FIBER_DEBUG_LOG":            "true",
	"APP_RESULT_LOGGER":              "CONSOLE",
	"APP_GCP_PROJECT":                "gcp-project-id",
	"APP_BQ_DATASET":                 "turing",
	"APP_BQ_TABLE":                   "turing-test",
	"APP_BQ_BATCH_LOAD":              "true",
	"APP_CUSTOM_METRICS":             "true",
	"APP_FLUENTD_HOST":               "localhost",
	"APP_FLUENTD_PORT":               "24224",
	"APP_FLUENTD_TAG":                "response.log",
	"APP_KAFKA_BROKERS":              "localhost:9000",
	"APP_KAFKA_TOPIC":                "kafka_topic",
	"APP_KAFKA_SERIALIZATION_FORMAT": "json",
	"APP_JAEGER_ENABLED":             "true",
	"APP_JAEGER_COLLECTOR_ENDPOINT":  "http://localhost:5000",
	"APP_JAEGER_REPORTER_HOST":       "localhost",
	"APP_JAEGER_REPORTER_PORT":       "5001",
	"SENTRY_ENABLED":                 "true",
	"SENTRY_DSN":                     "test:dsn",
	"SENTRY_LABELS":                  "sentry_key1:value1,sentry_key2:value2",
}

func TestMissingRequiredEnvs(t *testing.T) {
	// Setup
	setupNewEnv()
	_, err := InitConfigEnv()
	if err == nil {
		t.Error("Expected init config to fail, but it succeeded.")
	}
}

func TestInitConfigDefaultEnvs(t *testing.T) {
	expected := Config{
		Port: 8080,
		EnrichmentConfig: &EnrichmentConfig{
			Endpoint: "",
			Timeout:  15 * time.Millisecond,
		},
		RouterConfig: &RouterConfig{
			ConfigFile: "/var/test.yaml",
			Timeout:    20 * time.Millisecond,
			Protocol:   HTTP,
		},
		EnsemblerConfig: &EnsemblerConfig{
			Endpoint: "",
			Timeout:  10 * time.Millisecond,
		},
		AppConfig: &AppConfig{
			Name:          "turing",
			Environment:   "dev",
			LogLevel:      "INFO",
			FiberDebugLog: false,
			ResultLogger:  "NOP",
			BigQuery: &BQConfig{
				Project:   "",
				Dataset:   "",
				Table:     "",
				BatchLoad: false,
			},
			Fluentd: &FluentdConfig{
				Host: "",
				Port: 0,
				Tag:  "",
			},
			Kafka: &KafkaConfig{
				Brokers:             "",
				Topic:               "",
				SerializationFormat: SerializationFormat(""),
				MaxMessageBytes:     1048588,
				CompressionType:     "none",
			},
			CustomMetrics: false,
			Jaeger: &JaegerConfig{
				Enabled:           false,
				CollectorEndpoint: "",
				ReporterAgentHost: "",
				ReporterAgentPort: 0,
			},
			Sentry: sentry.Config{
				Enabled: false,
				DSN:     "",
				Labels:  nil,
			},
		},
	}

	// Setup
	setupNewEnv(requiredEnvs)
	cfg, err := InitConfigEnv()
	tu.FailOnError(t, err)

	// Test
	tu.FailOnError(t, tu.CompareObjects(cfg, &expected))
}

func TestInitConfigEnv(t *testing.T) {
	expected := Config{
		Port: 8080,
		EnrichmentConfig: &EnrichmentConfig{
			Endpoint: "http://localhost:8081",
			Timeout:  5 * time.Millisecond,
		},
		RouterConfig: &RouterConfig{
			ConfigFile: "/var/test.yaml",
			Timeout:    10 * time.Millisecond,
			Protocol:   UPI,
		},
		EnsemblerConfig: &EnsemblerConfig{
			Endpoint: "http://localhost:8082",
			Timeout:  2 * time.Millisecond,
		},
		AppConfig: &AppConfig{
			Name:          "turing",
			Environment:   "dev",
			LogLevel:      "DEBUG",
			FiberDebugLog: true,
			ResultLogger:  "CONSOLE",
			BigQuery: &BQConfig{
				Project:   "gcp-project-id",
				Dataset:   "turing",
				Table:     "turing-test",
				BatchLoad: true,
			},
			Fluentd: &FluentdConfig{
				Host: "localhost",
				Port: 24224,
				Tag:  "response.log",
			},
			Kafka: &KafkaConfig{
				Brokers:             "localhost:9000",
				Topic:               "kafka_topic",
				SerializationFormat: JSONSerializationFormat,
				MaxMessageBytes:     1048588,
				CompressionType:     "none",
			},
			CustomMetrics: true,
			Jaeger: &JaegerConfig{
				Enabled:           true,
				CollectorEndpoint: "http://localhost:5000",
				ReporterAgentHost: "localhost",
				ReporterAgentPort: 5001,
			},
			Sentry: sentry.Config{
				Enabled: true,
				DSN:     "test:dsn",
				Labels: map[string]string{
					"sentry_key1": "value1",
					"sentry_key2": "value2",
				},
			},
		},
	}

	// Setup
	setupNewEnv(requiredEnvs, optionalEnvs)
	cfg, err := InitConfigEnv()
	tu.FailOnError(t, err)

	// Test
	tu.FailOnError(t, tu.CompareObjects(cfg, &expected))
}

func TestGetters(t *testing.T) {
	cfg := Config{
		Port:             8080,
		EnrichmentConfig: &EnrichmentConfig{},
		RouterConfig: &RouterConfig{
			ConfigFile: "/var/test.yaml",
		},
		EnsemblerConfig: &EnsemblerConfig{},
	}

	// Test Getters
	assert.Equal(t, ":8080", cfg.ListenAddress())
	assert.Equal(t, "/var/test.yaml", cfg.RouterConfigFile())
}

func TestLogLevelDecode(t *testing.T) {
	// Make test cases
	tests := map[string]testSuiteLogLevel{
		"debug": {
			value:   "debug",
			result:  DebugLevel,
			success: true,
		},
		"info": {
			value:   "INFO",
			result:  InfoLevel,
			success: true,
		},
		"warn": {
			value:   "Warn",
			result:  WarnLevel,
			success: true,
		},
		"error": {
			value:   "error",
			result:  ErrorLevel,
			success: true,
		},
		"unknown_panic": {
			value:   "panic",
			result:  LogLevel(""),
			success: false,
		},
		"unknown_fatal": {
			value:   "fatal",
			result:  LogLevel(""),
			success: false,
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			var logLvl LogLevel
			err := logLvl.Decode(data.value)

			// Validate
			assert.Equal(t, data.result, logLvl)
			assert.Equal(t, data.success, err == nil)
		})
	}
}

func TestResultLoggerDecode(t *testing.T) {
	// Make test cases
	tests := map[string]testSuiteResultLogger{
		"bigquery": {
			value:   "BigQuery",
			result:  BigqueryLogger,
			success: true,
		},
		"console": {
			value:   "CONSOLE",
			result:  ConsoleLogger,
			success: true,
		},
		"nop": {
			value:   "nop",
			result:  NopLogger,
			success: true,
		},
		"unknown_resultlogger": {
			value:   "resultlogger",
			result:  ResultLogger(""),
			success: false,
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			var resLogger ResultLogger
			err := resLogger.Decode(data.value)

			// Validate
			assert.Equal(t, data.result, resLogger)
			assert.Equal(t, data.success, err == nil)
		})
	}
}

func TestSerializationFormatDecode(t *testing.T) {
	// Make test cases
	tests := map[string]testSuiteSerializationFormat{
		"json": {
			value:   "json",
			result:  JSONSerializationFormat,
			success: true,
		},
		"protobuf": {
			value:   "protobuf",
			result:  ProtobufSerializationFormat,
			success: true,
		},
		"unknown_serialization": {
			value:   "serialization",
			result:  SerializationFormat(""),
			success: false,
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			var serialization SerializationFormat
			err := serialization.Decode(data.value)

			// Validate
			assert.Equal(t, data.result, serialization)
			assert.Equal(t, data.success, err == nil)
		})
	}
}

func setupNewEnv(envMaps ...map[string]string) {
	os.Clearenv()

	for _, envMap := range envMaps {
		for key, val := range envMap {
			os.Setenv(key, val)
		}
	}
}
