package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	mlpcluster "github.com/caraml-dev/mlp/api/pkg/cluster"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/newrelic"
	"github.com/caraml-dev/mlp/api/pkg/instrumentation/sentry"
	"github.com/caraml-dev/mlp/api/pkg/webhooks"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"

	openapi "github.com/caraml-dev/turing/api/turing/generated"
	"github.com/caraml-dev/turing/api/turing/utils"

	// Using a maintained fork of https://github.com/spf13/viper mainly so that viper.AllSettings()
	// always returns map[string]interface{}. Without this, config for experiment cannot be
	// easily marshalled into JSON, which is currently the format required for experiment config.
	"github.com/ory/viper"
	"k8s.io/apimachinery/pkg/api/resource"
	sigyaml "sigs.k8s.io/yaml"
)

// Quantity is an alias for resource.Quantity
type Quantity resource.Quantity

// Decode parses the Quantity config and checks that it is non-empty and can be converted into
// a valid resource.Quantity.
func (qty *Quantity) Decode(value string) (err error) {
	if value == "" {
		return errors.New("value is empty")
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

// MarshalJSON implements the json.Marshaller interface so that we will get the expected
// JSON string representation when marshalling a quantity object.
func (qty *Quantity) MarshalJSON() ([]byte, error) {
	q := resource.Quantity(*qty)
	return q.MarshalJSON()
}

type EngineConfig map[string]interface{}

// EnvironmentConfig is an abridged version of
// https://github.dev/caraml-dev/merlin/blob/98ada0d3aa8de30d73e441d3fd1000fe5d5ac266/api/config/environment.go#L26
// Only requires Name and K8sConfig
// This struct should be removed when Environments API is moved from Merlin to MLP
type EnvironmentConfig struct {
	Name      string                `yaml:"name" validate:"required"`
	K8sConfig *mlpcluster.K8sConfig `yaml:"k8s_config" validate:"required"`
}

// Config is used to parse and store the environment configs
type Config struct {
	Port                          int `validate:"required"`
	LogLevel                      string
	AllowedOrigins                []string
	BatchEnsemblingConfig         BatchEnsemblingConfig         `validate:"required"`
	EnsemblerServiceBuilderConfig EnsemblerServiceBuilderConfig `validate:"required"`
	DbConfig                      *DatabaseConfig               `validate:"required"`
	DeployConfig                  *DeploymentConfig             `validate:"required"`
	SparkAppConfig                *SparkAppConfig               `validate:"required"`
	RouterDefaults                *RouterDefaults               `validate:"required"`
	KubernetesLabelConfigs        *KubernetesLabelConfigs       `validate:"required"`
	KnativeServiceDefaults        *KnativeServiceDefaults
	NewRelicConfig                newrelic.Config
	Sentry                        sentry.Config
	ClusterConfig                 ClusterConfig `validate:"required"`
	TuringEncryptionKey           string        `validate:"required"`
	AlertConfig                   *AlertConfig
	MLPConfig                     *MLPConfig `validate:"required"`
	TuringUIConfig                *SinglePageApplicationConfig
	OpenapiConfig                 *OpenapiConfig
	MlflowConfig                  *MlflowConfig
	WebhooksConfig                webhooks.Config

	// Experiment specifies the JSON configuration to set up experiment managers and runners.
	//
	// The configuration follows the following format to support different experiment engines
	// and to allow custom configuration for different engines:
	// { "<experiment_engine>": <experiment_engine_config>, ... }
	//
	// For example:
	// { "experiment_engine_a": {"client": "foo"}, "experiment_engine_b": {"apikey": 12} }
	Experiment map[string]EngineConfig
}

// ListenAddress returns the Turing Api app's port
func (c *Config) ListenAddress() string {
	return fmt.Sprintf(":%d", c.Port)
}

func (c *Config) Validate() error {
	validate, err := NewConfigValidator()
	if err != nil {
		return err
	}
	return validate.Struct(c)
}

// BatchEnsemblingConfig captures the config related to the running of batch runners
type BatchEnsemblingConfig struct {
	// Unfortunately if Enabled is false and user sets JobConfig/RunnerConfig/ImageBuildingConfig wrongly
	// it will still error out.
	Enabled             bool
	JobConfig           *JobConfig           `validate:"required_if=Enabled True"`
	RunnerConfig        *RunnerConfig        `validate:"required_if=Enabled True"`
	ImageBuildingConfig *ImageBuildingConfig `validate:"required_if=Enabled True"`
	LoggingURLFormat    *string
	MonitoringURLFormat *string
}

// EnsemblerServiceConfig captures the config related to the build and running of ensembler services (real-time)
type EnsemblerServiceBuilderConfig struct {
	ClusterName         string               `validate:"required"`
	ImageBuildingConfig *ImageBuildingConfig `validate:"required"`
}

// JobConfig captures the config related to the ensembling batch jobs.
type JobConfig struct {
	// DefaultEnvironment is the environment used for image building and running the batch ensemblers.
	DefaultEnvironment string `validate:"required"`
	// DefaultConfigurations contains the default configurations applied to the ensembling job.
	// The user (the person who calls the API) is free to override/append the default values.
	DefaultConfigurations DefaultEnsemblingJobConfigurations `validate:"required"`
}

// DefaultEnsemblingJobConfigurations contains the default configurations applied to the ensembling job.
type DefaultEnsemblingJobConfigurations struct {
	// BatchEnsemblingJobResources contains the resources delared to run the ensembling job.
	BatchEnsemblingJobResources openapi.EnsemblingResources
	// SparkConfigAnnotations contains the Spark configurations
	SparkConfigAnnotations map[string]string
}

// RunnerConfig contains the batch runner configurations
type RunnerConfig struct {
	// TimeInterval is the interval between job firings
	TimeInterval time.Duration `validate:"required"`
	// RecordsToProcessInOneIteration dictates the number of batch ensembling jobs to be queried at once.
	RecordsToProcessInOneIteration int `validate:"required"`
	// MaxRetryCount is the number of retries the batch ensembler runner should try before giving up.
	MaxRetryCount int `validate:"required"`
}

// ImageBuildingConfig contains the information regarding the image builder and the image buildee.
type ImageBuildingConfig struct {
	// BuildNamespace contains the Kubernetes namespace it should be built in.
	BuildNamespace string `validate:"required"`
	// BuildTimeoutDuration is the Kubernetes Job timeout duration.
	BuildTimeoutDuration time.Duration `validate:"required"`
	// DestinationRegistry is the registry of the newly built ensembler image.
	DestinationRegistry string `validate:"required"`
	// KanikoConfig contains the configuration related to the kaniko executor image builder.
	KanikoConfig KanikoConfig `validate:"required"`
	// TolerationName allow the scheduler to schedule image building jobs with the matching name
	TolerationName *string
	// NodeSelector restricts the running of image building jobs to nodes with the specified labels.
	NodeSelector map[string]string
	// Value for cluster-autoscaler.kubernetes.io/safe-to-evict annotation
	SafeToEvict bool
	// BaseImageRef is the image name of the base ensembler image built from the engines/pyfunc-ensembler-*/Dockerfile
	BaseImage string
}

// Resource contains the Kubernetes resource request and limits
type Resource struct {
	CPU    string `validate:"required"`
	Memory string `validate:"required"`
}

// ResourceRequestsLimits contains the Kubernetes resource request and limits for kaniko
type ResourceRequestsLimits struct {
	Requests Resource `validate:"required"`
	Limits   Resource `validate:"required"`
}

// KanikoConfig provides the configuration used for the Kaniko image.
type KanikoConfig struct {
	// BuildContextURI contains the image build context, which should be engines/batch-ensembler/
	// The forms supported are listed here https://github.com/GoogleContainerTools/kaniko#kaniko-build-contexts
	BuildContextURI string `validate:"required"`
	// DockerfileFilePath contains where the Dockerfile is
	DockerfileFilePath string `validate:"required"`
	// Image is the Kaniko image
	Image string `validate:"required"`
	// ImageVersion is the version tag of the Kaniko image
	ImageVersion string `validate:"required"`
	// AdditionalArgs allows platform-level additional arguments to be configured for Kaniko jobs
	AdditionalArgs []string
	// APIServerEnvVars allows extra API-server environment variables to be passed to Kaniko jobs
	APIServerEnvVars []string
	// Kaniko kubernetes service account
	ServiceAccount string
	// ResourceRequestsLimits is the resources required by Kaniko executor.
	ResourceRequestsLimits ResourceRequestsLimits `validate:"required"`
	// Kaniko push registry type
	PushRegistryType string `validate:"required,oneof=docker gcr"`
	// Kaniko docker credential secret name for pushing to docker registries
	DockerCredentialSecretName string
}

// SparkAppConfig contains the infra configurations that is unique to the user's Kubernetes
type SparkAppConfig struct {
	NodeSelector                   map[string]string
	APIServerEnvVars               []string
	CorePerCPURequest              float64 `validate:"required"`
	CPURequestToCPULimit           float64 `validate:"required"`
	SparkVersion                   string  `validate:"required"`
	TolerationName                 *string
	SubmissionFailureRetries       int32  `validate:"required"`
	SubmissionFailureRetryInterval int64  `validate:"required"`
	FailureRetries                 int32  `validate:"required"`
	FailureRetryInterval           int64  `validate:"required"`
	PythonVersion                  string `validate:"required"`
	TTLSecond                      int64  `validate:"required"`
}

// DeploymentConfig captures the config related to the deployment of the turing routers
type DeploymentConfig struct {
	EnvironmentType           string        `validate:"required"`
	Timeout                   time.Duration `validate:"required"`
	DeletionTimeout           time.Duration `validate:"required"`
	MaxCPU                    Quantity      `validate:"required"`
	MaxMemory                 Quantity      `validate:"required"`
	MaxAllowedReplica         int           `validate:"required"`
	TopologySpreadConstraints []corev1.TopologySpreadConstraint
	PodDisruptionBudget       PodDisruptionBudgetConfig
}

// PodDisruptionBudgetConfig are the configuration for PodDisruptionBudgetConfig for
// Turing services.
type PodDisruptionBudgetConfig struct {
	Enabled bool
	// Can specify only one of maxUnavailable and minAvailable
	MaxUnavailablePercentage *int
	MinAvailablePercentage   *int
}

// KubernetesLabelConfigs are the configurations for labeling
type KubernetesLabelConfigs struct {
	// LabelPrefix is the prefix used for tagging kubernetes components.
	// Default is an empty string which means your tags will look something like this:
	//   team: teen-titans
	//   stream: nile
	//   environment: dev
	//   orchestrator: turing
	//   app: my-model-app
	LabelPrefix string
	// NamespaceLabelPrefix is the prefix used for tagging kubernetes namespace.
	NamespaceLabelPrefix string
	// Environment is the value for the environment label
	Environment string `validate:"required"`
}

// KnativeServiceDefaults captures some of the configurable defaults specific to
// Knative services
type KnativeServiceDefaults struct {
	QueueProxyResourcePercentage          int
	UserContainerCPULimitRequestFactor    float64
	UserContainerMemoryLimitRequestFactor float64
	DefaultEnvVarsWithoutCPULimits        []corev1.EnvVar
}

// SinglePageApplicationConfig holds configuration required for serving SPAs
type SinglePageApplicationConfig struct {
	// Specifies the directory, that contains static files that will be served as an SPA
	ServingDirectory string
	// Defines the relative path under which the application will be accessible.
	ServingPath string
}

// DatabaseConfig config captures the Turing database config
type DatabaseConfig struct {
	Host     string `validate:"required"`
	Port     int    `validate:"required"`
	User     string `validate:"required"`
	Password string `validate:"required"`
	Database string `validate:"required"`

	MigrationsFolder string `validate:"required"`

	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
	MaxIdleConns    int
	MaxOpenConns    int
}

type ExperimentEngineConfig struct {
	PluginConfig              *ExperimentEnginePluginConfig
	ServiceAccountKeyFilePath *string
}

type ExperimentEnginePluginConfig struct {
	Image                 string `json:"image" validate:"required"`
	LivenessPeriodSeconds int    `json:"liveness_period_seconds" validate:"required"`
}

// RouterDefaults contains default configuration for routers deployed
// by this instance of the Turing API.
type RouterDefaults struct {
	// Turing router image, in the format registry/repository:version.
	Image string `validate:"required"`
	// Enable Fiber debug logging
	FiberDebugLogEnabled bool
	// Enable router custom metrics
	CustomMetricsEnabled bool
	// Enable Jaeger Tracing
	JaegerEnabled bool
	// Jaeger collector endpoint. If JaegerEnabled is true, this value
	// must be set.
	JaegerCollectorEndpoint string
	// Router log level
	LogLevel string `validate:"required"`
	// Fluentd config for the router
	FluentdConfig       *FluentdConfig
	MonitoringURLFormat *string
	// Configuration of experiment engine plugins, that consists of experiment engine name
	// and the image that contains the plugin implementation.
	//
	// Example:
	// ExperimentEnginePlugins:
	// 	red-exp-engine:
	//	  Image: ghcr.io/myproject/red-exp-engine-plugin:v0.0.1
	// 	blue-exp-engine:
	//	  Image: ghcr.io/myproject/blue-exp-engine-plugin:v0.0.1
	ExperimentEnginePlugins map[string]*ExperimentEngineConfig `validate:"dive"`
	// Kafka Configuration. If result logging is using Kafka
	KafkaConfig *KafkaConfig
	// UPIConfig config for UPI routers
	UPIConfig *UPIConfig
}

// FluentdConfig captures the defaults used by the Turing Router when Fluentd is enabled
type FluentdConfig struct {
	// Image to use for fluentd deployments, in the format registry/repository:version.
	Image string
	// Fluentd tag for logs
	Tag string
	// Flush interval seconds - value determined by load job frequency to BQ
	FlushIntervalSeconds int
	// No. of fluentd workers to launch for utilizing multiple CPU, useful to tune for high traffic
	WorkerCount int
}

// KafkaConfig captures the defaults used by Turing Router when result logger is set to kafka
type KafkaConfig struct {
	// Producer Config - Max message byte to send to broker
	MaxMessageBytes int
	// Producer Config - Compression Type of message
	CompressionType string
}

// UPIConfig captures the defaults used by UPI Router
type UPIConfig struct {
	// KafkaBrokers is broker which all Router will write to when UPI logging is enabled
	KafkaBrokers string
}

// ClusterConfig contains the cluster controller information.
// Supported features are in cluster configuration and Kubernetes client CA certificates.
type ClusterConfig struct {
	// InClusterConfig is a flag if the service account is provided in Kubernetes
	// and has the relevant credentials to handle all cluster operations.
	InClusterConfig bool

	// EnvironmentConfigPath refers to a path that contains EnvironmentConfigs
	EnvironmentConfigPath      string                `validate:"required_without=InClusterConfig"`
	EnsemblingServiceK8sConfig *mlpcluster.K8sConfig `validate:"required_without=InClusterConfig"`
	EnvironmentConfigs         []*EnvironmentConfig
}

// ProcessEnvConfigs reads the env configs from a file and unmarshalls them
func (c *ClusterConfig) ProcessEnvConfigs() error {
	if c.InClusterConfig {
		return nil
	}
	envConfig, err := os.ReadFile(c.EnvironmentConfigPath)
	if err != nil {
		return err
	}
	var envs []*EnvironmentConfig
	if err = yaml.Unmarshal(envConfig, &envs); err != nil {
		return err
	}
	for _, env := range envs {
		if env.K8sConfig == nil {
			return fmt.Errorf("Error, k8sConfig for %s is nil", env.Name)
		}
	}
	c.EnvironmentConfigs = envs
	return nil
}

type AlertConfig struct {
	Enabled bool
	GitLab  *GitlabConfig
	// PlaybookURL is the URL that contains documentation on how to resolve triggered alerts
	PlaybookURL string
	// DashboardURLTemplate is a template for grafana dashboard URL that shows router metrics.
	// The template accepts go-template format and will be executed with dashboardURLValue which has
	// the following fields: Environment, Cluster, Project, Router, Version.
	DashboardURLTemplate string
}

type GitlabConfig struct {
	BaseURL    string
	Token      string
	ProjectID  string
	Branch     string
	PathPrefix string
}

// MLPConfig captures the configuration used to connect to the Merlin/MLP API servers
type MLPConfig struct {
	MerlinURL string `validate:"required"`
	MLPURL    string `validate:"required"`
}

type MlflowConfig struct {
	TrackingURL string `validate:"required_if=ArtifactServiceType gcs ArtifactServiceType s3"`
	// Note that the Kaniko image builder needs to be configured correctly to have the necessary credentials to download
	// the artifacts from the blob storage tool depending on the artifact service type selected (gcs/s3). For gcs, the
	// credentials can be provided via a k8s service account or a secret but for s3, the credentials can be provided via
	// 1) additional arguments in the config KanikoConfig.AdditionalArgs e.g.
	// --build-arg=[AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY/AWS_DEFAULT_REGION/AWS_ENDPOINT_URL]=xxx
	// OR
	// 2) additional arguments in the config KanikoConfig.APIServerEnvVars, which will pass the specified environment
	// variables PRESENT within the Turing API server's container to the image builder as build arguments
	ArtifactServiceType string `validate:"required,oneof=nop gcs s3"`
}

// OpenapiConfig contains the settings for the OpenAPI specs used for validation and Swagger UI
type OpenapiConfig struct {
	// ValidationEnabled specifies whether to use OpenAPI validation middleware,
	// which validates HTTP requests against the spec.
	ValidationEnabled bool
	// SpecFile specifies the file path containing OpenAPI v3 spec
	SpecFile string
	// Config file to be used for serving swagger.
	SwaggerUIConfig *SinglePageApplicationConfig
	// Where the merged spec file yaml should be saved.
	MergedSpecFile string
	// Optional. Overrides the file before running the Swagger UI.
	SpecOverrideFile *string
}

// SpecData returns a byte slice of the combined yaml files
func (c *OpenapiConfig) SpecData() ([]byte, error) {
	return utils.MergeTwoYamls(c.SpecFile, c.SpecOverrideFile)
}

// GenerateSpecFile gets the spec data and writes to a file
func (c *OpenapiConfig) GenerateSpecFile() error {
	b, err := c.SpecData()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(c.MergedSpecFile), 0o755)
	if err != nil {
		return err
	}

	return utils.WriteYAMLFile(b, c.MergedSpecFile)
}

// Load creates a Config object from default config values, config files and environment variables.
// Load accepts config files as the argument. JSON and YAML format are both supported.
//
// If multiple config files are provided, the subsequent config files will override the config
// values from the config files loaded earlier.
//
// These config files will override the default config values (refer to setDefaultValues function)
// and can be overridden by the values from environment variables. Nested keys in the config
// can be set from environment variable name separed by "_". For instance the config value for
// "DbConfig.Port" can be overridden by environment variable name "DBCONFIG_PORT". Note that
// all environment variable names must be upper case.
//
// If no config file is provided, only the default config values and config values from environment
// varibales will be loaded.
//
// Refer to example.yaml for an example of config file.
func Load(filepaths ...string) (*Config, error) {
	v := viper.NewWithOptions(viper.KeyDelimiter("::"))

	// Load default config values
	setDefaultValues(v)

	// Load config values from the provided config files
	for _, f := range filepaths {
		v.SetConfigFile(f)
		err := v.MergeInConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to read config from file '%s': %s", f, err)
		}
	}

	// Load config values from environment variables.
	// Nested keys in the config is represented by variable name separated by '_'.
	// For example, DbConfig.Host can be set from environment variable DBCONFIG_HOST.
	v.SetEnvKeyReplacer(strings.NewReplacer("::", "_"))
	v.AutomaticEnv()

	config := &Config{}

	// Unmarshal config values into the config object.
	// Add StringToQuantityHookFunc() to the default DecodeHook in order to parse quantity string
	// into quantity object. Refs:
	// https://github.com/spf13/viper/blob/493643fd5e4b44796124c05d59ee04ba5f809e19/viper.go#L1003-L1005
	// https://github.com/mitchellh/mapstructure/blob/9e1e4717f8567d7ead72d070d064ad17d444a67e/decode_hooks_test.go#L128
	err := v.Unmarshal(config, func(c *mapstructure.DecoderConfig) {
		c.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
			StringToQuantityHookFunc(),
		)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config values: %s", err)
	}
	config, err = loadEnsemblingSvcConfig(config, v.AllSettings())
	if err != nil {
		return nil, fmt.Errorf("Failed to load ensemblingservicek8sconfig, err %s", err)
	}

	return config, nil
}

func loadEnsemblingSvcConfig(config *Config, v map[string]interface{}) (*Config, error) {
	// NOTE: This section is added to parse any fields in EnsemblingServiceK8sConfig that does not
	// have yaml tags.
	// For example `certificate-authority-data` is not unmarshalled
	// by vipers unmarshal method.

	clusterConfig, ok := v["clusterconfig"]
	if !ok {
		return config, nil
	}
	contents := clusterConfig.(map[string]interface{})
	ensemblingSvcK8sCfg, ok := contents["ensemblingservicek8sconfig"]
	if !ok {
		return config, nil
	}
	// convert back to byte string
	var byteForm []byte
	byteForm, err := yaml.Marshal(ensemblingSvcK8sCfg)
	if err != nil {
		return nil, err
	}
	// use sigyaml.Unmarshal to convert to json object then unmarshal

	k8sConfig := mlpcluster.K8sConfig{}
	if err := sigyaml.Unmarshal(byteForm, &k8sConfig); err != nil {
		return nil, err
	}
	config.ClusterConfig.EnsemblingServiceK8sConfig = &k8sConfig
	return config, nil
}

// setDefaultValues for all keys in Viper config. We need to set values for all keys so that
// we can always use environment variables to override the config keys. In Viper v1, if the
// keys do not have default values, and the key does not appear in the config file, it cannot
// be overridden by environment variables, unless each key is called with BindEnv.
// https://github.com/spf13/viper/issues/188
// https://github.com/spf13/viper/issues/761
func setDefaultValues(v *viper.Viper) {
	v.SetDefault("Port", "8080")

	v.SetDefault("AllowedOrigins", "*")

	v.SetDefault("DbConfig::Host", "localhost")
	v.SetDefault("DbConfig::Port", "5432")
	v.SetDefault("DbConfig::User", "")
	v.SetDefault("DbConfig::Password", "")
	v.SetDefault("DbConfig::Database", "turing")
	v.SetDefault("DbConfig::MigrationsFolder", "db-migrations/")

	v.SetDefault("DeployConfig::EnvironmentType", "")
	v.SetDefault("DeployConfig::Timeout", "3m")
	v.SetDefault("DeployConfig::DeletionTimeout", "1m")
	v.SetDefault("DeployConfig::MaxCPU", "4")
	v.SetDefault("DeployConfig::MaxMemory", "8Gi")
	v.SetDefault("DeployConfig::MaxAllowedReplica", "20")

	v.SetDefault("KnativeServiceDefaults::QueueProxyResourcePercentage", "30")
	v.SetDefault("KnativeServiceDefaults::UserContainerCPULimitRequestFactor", "0")
	v.SetDefault("KnativeServiceDefaults::UserContainerMemoryLimitRequestFactor", "1")

	v.SetDefault("RouterDefaults::Image", "")
	v.SetDefault("RouterDefaults::FiberDebugLogEnabled", "false")
	v.SetDefault("RouterDefaults::CustomMetricsEnabled", "false")
	v.SetDefault("RouterDefaults::JaegerEnabled", "false")
	v.SetDefault("RouterDefaults::JaegerCollectorEndpoint", "")
	v.SetDefault("RouterDefaults::LogLevel", "INFO")
	v.SetDefault("RouterDefaults::FluentdConfig::Image", "")
	v.SetDefault("RouterDefaults::FluentdConfig::Tag", "turing-result.log")
	v.SetDefault("RouterDefaults::FluentdConfig::FlushIntervalSeconds", "90")
	v.SetDefault("RouterDefaults::FluentdConfig::WorkerCount", "1")
	v.SetDefault("RouterDefaults::Experiment", map[string]interface{}{})
	v.SetDefault("RouterDefaults::KafkaConfig::MaxMessageBytes", "1048588")
	v.SetDefault("RouterDefaults::KafkaConfig::CompressionType", "none")

	v.SetDefault("Sentry::Enabled", "false")
	v.SetDefault("Sentry::DSN", "")

	v.SetDefault("TuringEncryptionKey", "")

	v.SetDefault("AlertConfig::Enabled", "false")
	v.SetDefault("AlertConfig::GitLab::BaseURL", "https://gitlab.com")
	v.SetDefault("AlertConfig::GitLab::Token", "")
	v.SetDefault("AlertConfig::GitLab::ProjectID", "")
	v.SetDefault("AlertConfig::GitLab::Branch", "master")
	v.SetDefault("AlertConfig::GitLab::PathPrefix", "turing")

	v.SetDefault("MLPConfig::MerlinURL", "")
	v.SetDefault("MLPConfig::MLPURL", "")
	v.SetDefault("MLPConfig::MLPEncryptionKey", "")

	v.SetDefault("TuringUIConfig::ServingDirectory", "")
	v.SetDefault("TuringUIConfig::ServingPath", "/turing")

	v.SetDefault("OpenapiConfig::SwaggerUIConfig::ServingDirectory", "")
	v.SetDefault("OpenapiConfig::SwaggerUIConfig::ServingPath", "/api-docs/")
	v.SetDefault("OpenapiConfig::ValidationEnabled", "true")
	v.SetDefault("OpenapiConfig::SpecFile", "api/openapi.bundle.yaml")
	v.SetDefault("OpenapiConfig::MergedSpecFile", "api/swagger-ui-dist/openapi.bundle.yaml")

	v.SetDefault("MlflowConfig::TrackingURL", "")
	v.SetDefault("MlflowConfig::ArtifactServiceType", "nop")

	v.SetDefault("Experiment", map[string]interface{}{})
}

func NewConfigValidator() (*validator.Validate, error) {
	v := validator.New()

	// Use struct level validation for PodDisruptionBudgetConfig
	v.RegisterStructValidation(func(sl validator.StructLevel) {
		field := sl.Current().Interface().(PodDisruptionBudgetConfig)
		// If PDB is enabled, one of max unavailable or min available shall be set
		if field.Enabled &&
			(field.MaxUnavailablePercentage == nil && field.MinAvailablePercentage == nil) ||
			(field.MaxUnavailablePercentage != nil && field.MinAvailablePercentage != nil) {
			sl.ReportError(field.MaxUnavailablePercentage, "max_unavailable_percentage", "int",
				"choose_one[max_unavailable_percentage,min_available_percentage]", "")
			sl.ReportError(field.MinAvailablePercentage, "min_available_percentage", "int",
				"choose_one[max_unavailable_percentage,min_available_percentage]", "")
		}
	}, PodDisruptionBudgetConfig{})

	return v, nil
}

// StringToQuantityHookFunc converts string to quantity type. This function is required since
// viper uses mapstructure to unmarshal values.
// https://github.com/spf13/viper#unmarshaling
// https://pkg.go.dev/github.com/mitchellh/mapstructure#DecodeHookFunc
func StringToQuantityHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		if t != reflect.TypeOf(Quantity{}) {
			return data, nil
		}

		// Convert it by parsing
		q, err := resource.ParseQuantity(data.(string))
		if err != nil {
			return nil, err
		}

		return Quantity(q), nil
	}
}
