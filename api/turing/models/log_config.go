package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	routercfg "github.com/gojek/turing/engines/router/missionctl/config"
)

// ResultLogger is the type used to capture the supported response
// logging destinations
type ResultLogger string

const (
	// BigQueryLogger logs the responses to BigQuery
	BigQueryLogger ResultLogger = "bigquery"
	// ConsoleLogger logs the responses to console
	ConsoleLogger ResultLogger = "console"
	// KafkaLogger logs the responses to kafka
	KafkaLogger ResultLogger = "kafka"
	// NopLogger disables response logging
	NopLogger ResultLogger = "nop"
)

// BigQueryConfig contains the configuration to log results to  BigQuery.
type BigQueryConfig struct {
	// BigQuery table to write to, as a fully qualified BQ Table string.
	// e.g. project.dataset.table
	Table string `json:"table"`
	// Service account secret name (correct to merlin) for writing to BQ.
	ServiceAccountSecret string `json:"service_account_secret"`
	// Whether to perform batch or streaming writes.
	BatchLoad bool `json:"batch_load"`
}

type KafkaConfig struct {
	// List of brokers for the kafka to write logs to
	Brokers string `json:"brokers"`
	// Topic to write logs to
	Topic string `json:"topic"`
}

// LogConfig contains all log configuration necessary for a deployment
// of the Turing Router.
type LogConfig struct {
	// LogLevel of the router.
	LogLevel routercfg.LogLevel `json:"log_level"`
	// Enable custom metrics for the router. Defaults to false.
	CustomMetricsEnabled bool `json:"custom_metrics_enabled"`
	// Enable debug logs for Fiber. Defaults to false.
	FiberDebugLogEnabled bool `json:"fiber_debug_log_enabled"`
	// Enable Jaeger tracing.
	JaegerEnabled bool `json:"jaeger_enabled"`
	// Result Logger type. The associated config must not be null.
	ResultLoggerType ResultLogger `json:"result_logger_type"`
	// Configuration necessary to log results to BigQuery. Cannot be empty if
	// ResultLoggerType is set to "bigquery".
	BigQueryConfig *BigQueryConfig `json:"bigquery_config,omitempty"`
	// Configuration necessary to log results to kafka. Cannot be empty if
	// ResultLoggerType is set to "kafka"
	KafkaConfig *KafkaConfig `json:"kafka_config,omitempty"`
}

func (l LogConfig) Value() (driver.Value, error) {
	return json.Marshal(l)
}

func (l *LogConfig) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &l)
}
