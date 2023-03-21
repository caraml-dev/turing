package request

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/engines/experiment/manager"
	routerConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"
)

// CreateOrUpdateRouterRequest structure defines the format of the request payload
// when creating or updating routers
type CreateOrUpdateRouterRequest struct {
	Environment string        `json:"environment_name" validate:"required"`
	Name        string        `json:"name" validate:"required"`
	Config      *RouterConfig `json:"config" validate:"required,dive"`
}

// RouterConfig defines the properties of the specific router version
type RouterConfig struct {
	Routes             models.Routes              `json:"routes" validate:"required"`
	DefaultRouteID     *string                    `json:"default_route_id"`
	DefaultTrafficRule *models.DefaultTrafficRule `json:"default_traffic_rule,omitempty"`
	TrafficRules       models.TrafficRules        `json:"rules" validate:"unique=Name,dive"`
	ExperimentEngine   *ExperimentEngineConfig    `json:"experiment_engine" validate:"required,dive"`
	ResourceRequest    *models.ResourceRequest    `json:"resource_request"`
	AutoscalingPolicy  *models.AutoscalingPolicy  `json:"autoscaling_policy" validate:"omitempty,dive"`
	Timeout            string                     `json:"timeout" validate:"required"`
	Protocol           *routerConfig.Protocol     `json:"protocol"`

	LogConfig *LogConfig `json:"log_config" validate:"required"`

	Enricher  *EnricherEnsemblerConfig `json:"enricher,omitempty" validate:"omitempty,dive"`
	Ensembler *models.Ensembler        `json:"ensembler,omitempty" validate:"omitempty,dive"`
}

// ExperimentEngineConfig defines the experiment engine config
type ExperimentEngineConfig struct {
	Type   string          `json:"type" validate:"required"`
	Config json.RawMessage `json:"config,omitempty" validate:"-"` // Skip validate to invoke custom validation
}

// LogConfig defines the logging configs
type LogConfig struct {
	ResultLoggerType models.ResultLogger `json:"result_logger_type"`
	BigQueryConfig   *BigQueryConfig     `json:"bigquery_config,omitempty"`
	KafkaConfig      *KafkaConfig        `json:"kafka_config,omitempty"`
}

// BigQueryConfig defines the configs for logging to BQ
type BigQueryConfig struct {
	Table                string `json:"table"`
	ServiceAccountSecret string `json:"service_account_secret"`
}

// KafkaConfig defines the configs for logging to Kafka
type KafkaConfig struct {
	Brokers             string                     `json:"brokers"`
	Topic               string                     `json:"topic"`
	SerializationFormat models.SerializationFormat `json:"serialization_format"`
}

// EnricherEnsemblerConfig defines the configs for the enricher / ensembler,
// used by the specific router config
type EnricherEnsemblerConfig struct {
	// Fully qualified docker image string used by the enricher, in the
	// format registry/repository:version.
	Image string `json:"image" validate:"required"`
	// Resource requests  for the deployment of the enricher.
	ResourceRequest *models.ResourceRequest `json:"resource_request" validate:"required"`
	// Autoscaling policy for the enricher / ensembler.
	AutoscalingPolicy *models.AutoscalingPolicy `json:"autoscaling_policy" validate:"omitempty,dive"`
	// Endpoint to query.
	Endpoint string `json:"endpoint" validate:"required"`
	// Request timeout as a valid quantity string.
	Timeout string `json:"timeout" validate:"required"`
	// Port to query.
	Port int `json:"port" validate:"required"`
	// Environment variables to inject into the pod.
	Env models.EnvVars `json:"env" validate:"required"`
	// ServiceAccount specifies the name of the secret registered in the MLP project containing the service account.
	// The service account will be mounted into the user-container and the environment variable
	// GOOGLE_APPLICATION_CREDENTIALS will reference the service account file.
	ServiceAccount string `json:"service_account"`
}

// BuildEnricher builds the enricher model from the enricher config
func (cfg EnricherEnsemblerConfig) BuildEnricher() *models.Enricher {
	return &models.Enricher{
		Image:             cfg.Image,
		ResourceRequest:   cfg.ResourceRequest,
		AutoscalingPolicy: getAutoscalingPolicyOrDefault(cfg.AutoscalingPolicy),
		Endpoint:          cfg.Endpoint,
		Timeout:           cfg.Timeout,
		Port:              cfg.Port,
		Env:               cfg.Env,
		ServiceAccount:    cfg.ServiceAccount,
	}
}

// BuildRouter builds the router model from the entire request payload
func (r CreateOrUpdateRouterRequest) BuildRouter(projectID models.ID) *models.Router {
	return &models.Router{
		ProjectID:       projectID,
		EnvironmentName: r.Environment,
		Name:            r.Name,
		Status:          models.RouterStatusPending,
	}
}

// BuildRouterVersion builds the router version model from the entire request payload
func (r RouterConfig) BuildRouterVersion(
	projectName string,
	router *models.Router,
	defaults *config.RouterDefaults,
	cryptoSvc service.CryptoService,
	expSvc service.ExperimentsService,
	ensemblersSvc service.EnsemblersService,
) (rv *models.RouterVersion, err error) {
	var defaultRouteID string
	if r.DefaultRouteID != nil {
		defaultRouteID = *r.DefaultRouteID
	}

	// Set default to http
	routerProtocol := routerConfig.HTTP
	if r.Protocol != nil {
		routerProtocol = *r.Protocol
	}
	if routerProtocol != routerConfig.UPI && routerProtocol != routerConfig.HTTP {
		return nil, errors.New("invalid router protocol")
	}

	rv = &models.RouterVersion{
		RouterID:           router.ID,
		Router:             router,
		Image:              defaults.Image,
		Status:             models.RouterVersionStatusPending,
		Routes:             r.Routes,
		DefaultRouteID:     defaultRouteID,
		DefaultTrafficRule: r.DefaultTrafficRule,
		TrafficRules:       r.TrafficRules,
		ExperimentEngine: &models.ExperimentEngine{
			Type: r.ExperimentEngine.Type,
		},
		ResourceRequest:   r.ResourceRequest,
		AutoscalingPolicy: getAutoscalingPolicyOrDefault(r.AutoscalingPolicy),
		Timeout:           r.Timeout,
		Protocol:          routerProtocol,
		LogConfig: &models.LogConfig{
			LogLevel:             routerConfig.LogLevel(defaults.LogLevel),
			CustomMetricsEnabled: defaults.CustomMetricsEnabled,
			FiberDebugLogEnabled: defaults.FiberDebugLogEnabled,
			JaegerEnabled:        defaults.JaegerEnabled,
			ResultLoggerType:     r.LogConfig.ResultLoggerType,
		},
	}
	if r.Enricher != nil {
		rv.Enricher = r.Enricher.BuildEnricher()
	}
	if r.Ensembler != nil {
		// Ensure ensembler config is set based on the ensembler type
		if r.Ensembler.Type == models.EnsemblerDockerType {
			if r.Ensembler.DockerConfig == nil {
				return nil, errors.New("missing ensembler docker config")
			}
			r.Ensembler.DockerConfig.AutoscalingPolicy = getAutoscalingPolicyOrDefault(
				r.Ensembler.DockerConfig.AutoscalingPolicy)
		}
		if r.Ensembler.Type == models.EnsemblerStandardType && r.Ensembler.StandardConfig == nil {
			return nil, errors.New("missing ensembler standard config")
		}
		if r.Ensembler.Type == models.EnsemblerPyFuncType {
			if r.Ensembler.PyfuncConfig == nil {
				return nil, errors.New("missing ensembler pyfunc config")
			}

			r.Ensembler.PyfuncConfig.AutoscalingPolicy = getAutoscalingPolicyOrDefault(
				r.Ensembler.PyfuncConfig.AutoscalingPolicy)

			// Verify if the ensembler given by its ProjectID and EnsemblerID exist
			ensembler, err := ensemblersSvc.FindByID(
				*r.Ensembler.PyfuncConfig.EnsemblerID,
				service.EnsemblersFindByIDOptions{
					ProjectID: r.Ensembler.PyfuncConfig.ProjectID,
				})
			if err != nil {
				return nil, fmt.Errorf("failed to find specified ensembler: %w", err)
			}

			// Check if retrieved ensembler as a pyfunc ensembler
			switch v := ensembler.(type) {
			case *models.PyFuncEnsembler:
				break
			default:
				return nil, fmt.Errorf("only pyfunc ensemblers allowed; ensembler type given: %T", v)
			}
		}
		rv.Ensembler = r.Ensembler
	}
	switch rv.LogConfig.ResultLoggerType {
	case models.BigQueryLogger:
		rv.LogConfig.BigQueryConfig = &models.BigQueryConfig{
			Table:                r.LogConfig.BigQueryConfig.Table,
			ServiceAccountSecret: r.LogConfig.BigQueryConfig.ServiceAccountSecret,
			BatchLoad:            true, // default for now
		}
	case models.KafkaLogger:
		rv.LogConfig.KafkaConfig = &models.KafkaConfig{
			Brokers:             r.LogConfig.KafkaConfig.Brokers,
			Topic:               r.LogConfig.KafkaConfig.Topic,
			SerializationFormat: r.LogConfig.KafkaConfig.SerializationFormat,
		}
	case models.UPILogger:
		rv.LogConfig.KafkaConfig = &models.KafkaConfig{
			Brokers:             defaults.UPIConfig.KafkaBrokers,
			Topic:               fmt.Sprintf("caraml-%s-%s-router-log", projectName, router.Name),
			SerializationFormat: models.ProtobufSerializationFormat,
		}
	}
	if rv.ExperimentEngine.Type != models.ExperimentEngineTypeNop {
		if experimentEnginePlugin, ok := defaults.ExperimentEnginePlugins[rv.ExperimentEngine.Type]; ok {
			rv.ExperimentEngine.PluginConfig = experimentEnginePlugin.PluginConfig
			if experimentEnginePlugin.ServiceAccountKeyFilePath != nil {
				rv.ExperimentEngine.ServiceAccountKeyFilePath = experimentEnginePlugin.ServiceAccountKeyFilePath
			}
		}

		rv.ExperimentEngine.Config, err = r.BuildExperimentEngineConfig(router, cryptoSvc, expSvc)
		if err != nil {
			return nil, err
		}
	}

	return rv, nil
}

// BuildExperimentEngineConfig creates the Experiment config from the given input properties
func (r RouterConfig) BuildExperimentEngineConfig(
	router *models.Router,
	cryptoSvc service.CryptoService,
	expSvc service.ExperimentsService,
) (json.RawMessage, error) {
	rawExpConfig := r.ExperimentEngine.Config
	// Handle missing passkey / encrypt it, if Standard experiment config using client selection
	isClientSelectionEnabled, err := expSvc.IsClientSelectionEnabled(r.ExperimentEngine.Type)
	if err != nil {
		return nil, err
	}
	if isClientSelectionEnabled {
		// Convert the new config to the standard type
		expConfig, err := manager.ParseStandardExperimentConfig(rawExpConfig)
		if err != nil {
			return nil, fmt.Errorf("Cannot parse standard experiment config: %v", err)
		}

		if expConfig.Client.Passkey == "" {
			// Extract existing router version config
			if router.CurrRouterVersion != nil &&
				router.CurrRouterVersion.ExperimentEngine.Type == r.ExperimentEngine.Type {
				currVerExpConfig, err := manager.ParseStandardExperimentConfig(router.CurrRouterVersion.ExperimentEngine.Config)
				if err != nil {
					return nil, fmt.Errorf("Error parsing existing experiment config: %v", err)
				}
				if expConfig.Client.Username == currVerExpConfig.Client.Username {
					// Copy the passkey
					expConfig.Client.Passkey = currVerExpConfig.Client.Passkey
				}
			}
			// If the passkey is still empty, we cannot proceed
			if expConfig.Client.Passkey == "" {
				return nil, errors.New("Passkey must be configured")
			}
		} else {
			// Passkey has been supplied, encrypt it
			var err error
			expConfig.Client.Passkey, err = cryptoSvc.Encrypt(expConfig.Client.Passkey)
			if err != nil {
				return nil, fmt.Errorf("Passkey could not be encrypted: %s", err.Error())
			}
		}

		// Marshal Experiment engine config
		return json.Marshal(expConfig)
	}

	// Custom experiment manager config, return as is.
	return rawExpConfig, nil
}

// getAutoscalingPolicyOrDefault applies the default autoscaling policy that has been used all along, prior to
// user configurations. Thus, this function also ensures backward-compatibility with requests originating from
// older SDK versions.
// TODO: Remove the default values from the API and make the autoscaling parameters required.
func getAutoscalingPolicyOrDefault(policy *models.AutoscalingPolicy) *models.AutoscalingPolicy {
	if policy == nil {
		return &models.AutoscalingPolicy{
			Metric: models.AutoscalingMetricConcurrency,
			Target: "1",
		}
	}
	return policy
}
