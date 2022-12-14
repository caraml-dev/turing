package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"
	"github.com/kelseyhightower/envconfig"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
)

// LogLevel type is used to capture the supported logging levels
type LogLevel string

const (
	// DebugLevel is used for verbose logs at debug level
	DebugLevel LogLevel = "DEBUG"
	// InfoLevel is used for logs that are info level and higher
	InfoLevel LogLevel = "INFO"
	// WarnLevel is used for logs that are warning level and higher
	WarnLevel LogLevel = "WARN"
	// ErrorLevel is used for logs that are error level and higher
	ErrorLevel LogLevel = "ERROR"
)

// ResultLogger is the type used to capture the supported response
// logging destinations
type ResultLogger string

const (
	// BigqueryLogger logs the responses to BigQuery
	BigqueryLogger ResultLogger = "BIGQUERY"
	// ConsoleLogger logs the responses to console
	ConsoleLogger ResultLogger = "CONSOLE"
	// KafkaLogger logs the response to a Kafka topic
	KafkaLogger ResultLogger = "KAFKA"
	// NopLogger disables response logging
	NopLogger ResultLogger = "NOP"
)

// Protocol is used for router config to indicate spinning up of respective router mission control
type Protocol string

const (
	UPI  Protocol = "UPI_V1"
	HTTP Protocol = "HTTP_JSON"
)

// SerializationFormat represents the message serialization format to be used by the ResultLogger
type SerializationFormat string

const (
	// JSONSerializationFormat formats the message as json, for logging
	JSONSerializationFormat SerializationFormat = "json"
	// ProtobufSerializationFormat formats the message using protobuf, for logging
	ProtobufSerializationFormat SerializationFormat = "protobuf"
)

// Config is the structure used to parse the environment configs
type Config struct {
	Port int `envconfig:"PORT" required:"true"`

	EnrichmentConfig *EnrichmentConfig `envconfig:"ENRICHER"`
	RouterConfig     *RouterConfig     `envconfig:"ROUTER"`
	EnsemblerConfig  *EnsemblerConfig  `envconfig:"ENSEMBLER"`

	AppConfig *AppConfig `envconfig:"APP"`
}

// ListenAddress retrieves the Mission Control's configured port
func (c *Config) ListenAddress() string {
	return fmt.Sprintf(":%d", c.Port)
}

// RouterConfigFile retrieves the router's config file path
func (c *Config) RouterConfigFile() string {
	return c.RouterConfig.ConfigFile
}

// BQConfig is used to parse the APP config env vars that hold the BQ connection
// information. These env vars are optional, i.e., only required to be set for
// ResultLogger type BigqueryLogger. If they are not set for BQLogger type, init will
// fail. If BatchLoad is true, the Fluentd config is also expected to be set.
type BQConfig struct {
	Project   string `envconfig:"APP_GCP_PROJECT"`
	Dataset   string
	Table     string
	BatchLoad bool `split_words:"true" default:"false"`
}

// FluentdConfig is used to parse the APP config env var(s) that hold the Fluentd
// backend information
type FluentdConfig struct {
	Host string
	Port int
	Tag  string
}

// KafkaConfig captures the minimal configuration for writing result logs to
// Kafka topics
type KafkaConfig struct {
	Brokers             string
	Topic               string
	SerializationFormat SerializationFormat `split_words:"true"`
	MaxMessageBytes     int                 `split_words:"true" default:"1048588"`
	CompressionType     string              `split_words:"true" default:"none"`
}

// JaegerConfig captures the settings for tracing using Jaeger client
// Ref: https://pkg.go.dev/github.com/uber/jaeger-client-go/config
type JaegerConfig struct {
	Enabled           bool
	CollectorEndpoint string `split_words:"true"`
	ReporterAgentHost string `envconfig:"REPORTER_HOST" split_words:"true"`
	ReporterAgentPort int    `envconfig:"REPORTER_PORT" split_words:"true"`
}

// AppConfig is the structure used to the parse the environment configs that correspond
// to application behavior such as logging, instrumentation, etc.
type AppConfig struct {
	Name          string       `required:"true"`
	Environment   string       `required:"true"`
	LogLevel      LogLevel     `split_words:"false" default:"INFO"`
	CustomMetrics bool         `split_words:"true" default:"false"`
	FiberDebugLog bool         `split_words:"true" default:"false"`
	ResultLogger  ResultLogger `split_words:"true" default:"NOP"`
	BigQuery      *BQConfig    `envconfig:"BQ"`
	Fluentd       *FluentdConfig
	Kafka         *KafkaConfig
	Jaeger        *JaegerConfig
	Sentry        sentry.Config
}

// Decode parses the LogLevel config defined and validates if it is one of the supported
// values.
func (logLvl *LogLevel) Decode(value string) error {
	value = strings.ToUpper(value)
	switch LogLevel(value) {
	case DebugLevel,
		InfoLevel,
		WarnLevel,
		ErrorLevel:
		*logLvl = LogLevel(value)
		return nil
	}
	return errors.Newf(errors.BadConfig, "Log level value %s not supported", value)
}

// Decode parses the ResultLogger config defined and validates if it is one of the
// supported values.
func (resLogger *ResultLogger) Decode(value string) error {
	value = strings.ToUpper(value)
	switch ResultLogger(value) {
	case BigqueryLogger,
		ConsoleLogger,
		KafkaLogger,
		NopLogger:
		*resLogger = ResultLogger(value)
		return nil
	}
	return errors.Newf(errors.BadConfig, "Response logger value %s not supported", value)
}

// Decode parses the SerializationFormat config and validates if it is one of the
// supported values.
func (serialization *SerializationFormat) Decode(value string) error {
	value = strings.ToLower(value)
	switch SerializationFormat(value) {
	case JSONSerializationFormat,
		ProtobufSerializationFormat:
		*serialization = SerializationFormat(value)
		return nil
	}
	return errors.Newf(errors.BadConfig, "Serialization format value %s not supported", value)
}

// EnrichmentConfig is the structure used to parse the Enricher's environment configs
type EnrichmentConfig struct {
	Endpoint string
	Timeout  time.Duration `default:"15ms"`
}

// RouterConfig is the structure used to parse the Router's environment configs
type RouterConfig struct {
	ConfigFile string        `split_words:"true" required:"true"`
	Timeout    time.Duration `default:"20ms"`
	Protocol   Protocol      `default:"HTTP_JSON"`
}

// EnsemblerConfig is the structure used to parse the Ensembler's environment configs
type EnsemblerConfig struct {
	Endpoint string
	Timeout  time.Duration `default:"10ms"`
}

// InitConfigEnv initialises configuration from the environment.
func InitConfigEnv() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
