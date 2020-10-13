package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gojek/mlp/pkg/instrumentation/sentry"
	"github.com/kelseyhightower/envconfig"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Quantity is an alias for resource.Quantity
type Quantity resource.Quantity

// Decode parses the Quantity config and checks that it is non-empty and can be converted into
// a valid resource.Quantity.
func (qty *Quantity) Decode(value string) (err error) {
	if value == "" {
		return errors.New("Value is empty")
	}

	// MustParse panics if the supplied value cannot be parsed
	defer func() {
		if panicErr := recover(); panicErr != nil {
			err = panicErr.(error)
		}
	}()
	*qty = Quantity(resource.MustParse(value))

	return nil
}

// Config is used to parse and store the environment configs
type Config struct {
	Port                int                  `envconfig:"turing_port" default:"8080"`
	AuthConfig          *AuthorizationConfig `envconfig:"authorization" validate:"required"`
	DbConfig            *DatabaseConfig      `envconfig:"turing_database"`
	DeployConfig        *DeploymentConfig    `envconfig:"deployment"`
	RouterDefaults      *RouterDefaults      `envconfig:"turing_router"`
	Sentry              sentry.Config        `split_words:"false" envconfig:"SENTRY"`
	VaultConfig         *VaultConfig         `envconfig:"vault"`
	TuringEncryptionKey string               `split_words:"true" required:"true"`
	AlertConfig         *AlertConfig         `envconfig:"alert"`
	MLPConfig           *MLPConfig
	TuringUIConfig      *TuringUIConfig `envconfig:"turing_ui"`
	// SwaggerFile specifies the file path containing OpenAPI spec. This file will be used to configure
	// OpenAPI validation middleware, which validates HTTP requests against the spec.
	SwaggerFile string `envconfig:"swagger_file" default:"swagger.yaml"`
}

// ListenAddress returns the Turing Api app's port
func (c *Config) ListenAddress() string {
	return fmt.Sprintf(":%d", c.Port)
}

// DeploymentConfig captures the config related to the deployment of the turing routers
type DeploymentConfig struct {
	EnvironmentType string        `envconfig:"environment_type" required:"true"`
	GcpProject      string        `envconfig:"gcp_project" required:"true"`
	Timeout         time.Duration `required:"true"`
	DeletionTimeout time.Duration `envconfig:"deletion_timeout" required:"true"`
	MaxCPU          Quantity      `envconfig:"max_cpu" required:"true"`
	MaxMemory       Quantity      `envconfig:"max_memory" required:"true"`
}

// TuringUIConfig captures config related to serving Turing UI files
type TuringUIConfig struct {
	// Optional. If configured, turing-api will serve static files of the turing-ui React app
	AppDirectory string `envconfig:"app_directory"`
	// Optional. Defines the relative path under which the app will be accessible.
	// This should match `homepage` value from the `package.json` file of the CRA app
	Homepage string `default:"/turing"`
}

// DatabaseConfig config captures the Turing database config
type DatabaseConfig struct {
	Host     string `required:"true"`
	Port     int    `default:"5432"`
	User     string `required:"true"`
	Password string `required:"true"`
	Database string `envconfig:"name" required:"true"`
}

// RouterDefaults contains default configuration for routers deployed
// by this isntance of the Turing API.
type RouterDefaults struct {
	// Turing router image, in the format registry/repository:version.
	Image string `required:"true"`
	// Enable Fiber debug logging
	FiberDebugLogEnabled bool `split_words:"true" default:"true"`
	// Enable router custom metrics
	CustomMetricsEnabled bool `split_words:"true" default:"true"`
	// Enable Jaeger Tracing
	JaegerEnabled bool `split_words:"true" default:"true"`
	// Jaeger collector endpoint. If JaegerEnabled is true, this value
	// must be set.
	JaegerCollectorEndpoint string `split_words:"true" required:"true"`
	// Router log level
	LogLevel string `split_words:"true" default:"INFO"`
	// Fluentd config for the router
	FluentdConfig *FluentdConfig `envconfig:"fluentd"`
}

// FluentdConfig captures the defaults used by the Turing Router when Fluentd is enabled
type FluentdConfig struct {
	// Image to use for fluentd deployments, in the format registry/repository:version.
	Image string `split_words:"true" required:"true"`
	// Fluentd tag for logs
	Tag string `split_words:"true" default:"turing-result.log"`
	// Flush interval seconds - value determined by load job frequency to BQ
	FlushIntervalSeconds int `split_words:"true" default:"90"`
}

// AuthorizationConfig captures the config for auth using mlp authz
type AuthorizationConfig struct {
	Enabled bool `default:"true"`
	URL     string
}

// VaultConfig captures the config for connecting to the Vault server
type VaultConfig struct {
	Address string `required:"true"`
	Token   string `required:"true"`
}

type AlertConfig struct {
	Enabled bool          `default:"false"`
	GitLab  *GitlabConfig `envconfig:"gitlab"`
}

type GitlabConfig struct {
	BaseURL    string
	Token      string
	ProjectID  string
	Branch     string `default:"master"`
	PathPrefix string `default:"turing"`
}

// MLPConfig captures the configuration used to connect to the Merlin/MLP API servers
type MLPConfig struct {
	MerlinURL        string `envconfig:"merlin_url" required:"true"`
	MLPURL           string `envconfig:"mlp_url" required:"true"`
	MLPEncryptionKey string `envconfig:"mlp_encryption_key" required:"true"`
}

// FromEnv loads the configs from the supplied environment
func FromEnv() (*Config, error) {
	var cfg Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}

	// Run config validator
	validate, err := newConfigValidator()
	if err != nil {
		return nil, fmt.Errorf("Unable to init Config validator: %v", err)
	}
	err = validate.Struct(cfg)
	if err != nil {
		return nil, fmt.Errorf("Config validation failed: %v", err)
	}

	return &cfg, nil
}

func newConfigValidator() (*validator.Validate, error) {
	v := validator.New()
	// Use struct level validation for AuthorizationConfig
	v.RegisterStructValidation(func(sl validator.StructLevel) {
		field := sl.Current().Interface().(AuthorizationConfig)
		// If auth is enabled, URL should be set
		if field.Enabled && field.URL == "" {
			sl.ReportError(field.URL, "authorization_url", "URL", "url-set", "")
		}
	}, AuthorizationConfig{})
	return v, nil
}
