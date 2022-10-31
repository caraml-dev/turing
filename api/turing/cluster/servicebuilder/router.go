package servicebuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	fiberConfig "github.com/gojek/fiber/config"
	fiberProtocol "github.com/gojek/fiber/protocol"
	mlp "github.com/gojek/mlp/api/client"
	corev1 "k8s.io/api/core/v1"

	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/utils"
	"github.com/caraml-dev/turing/engines/router"
	routeConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"
	"github.com/caraml-dev/turing/engines/router/missionctl/fiberapi"
)

// Define env var names for the router
const (
	envAppName                      = "APP_NAME"
	envAppEnvironment               = "APP_ENVIRONMENT"
	envRouterTimeout                = "ROUTER_TIMEOUT"
	envEnricherEndpoint             = "ENRICHER_ENDPOINT"
	envEnricherTimeout              = "ENRICHER_TIMEOUT"
	envEnsemblerEndpoint            = "ENSEMBLER_ENDPOINT"
	envEnsemblerTimeout             = "ENSEMBLER_TIMEOUT"
	envLogLevel                     = "APP_LOGLEVEL"
	envFiberDebugLog                = "APP_FIBER_DEBUG_LOG"
	envCustomMetrics                = "APP_CUSTOM_METRICS"
	envJaegerEnabled                = "APP_JAEGER_ENABLED"
	envJaegerEndpoint               = "APP_JAEGER_COLLECTOR_ENDPOINT"
	envSentryEnabled                = "APP_SENTRY_ENABLED"
	envSentryDSN                    = "APP_SENTRY_DSN"
	envResultLogger                 = "APP_RESULT_LOGGER"
	envGcpProject                   = "APP_GCP_PROJECT"
	envBQDataset                    = "APP_BQ_DATASET"
	envBQTable                      = "APP_BQ_TABLE"
	envBQBatchLoad                  = "APP_BQ_BATCH_LOAD"
	envFluentdHost                  = "APP_FLUENTD_HOST"
	envFluentdPort                  = "APP_FLUENTD_PORT"
	envFluentdTag                   = "APP_FLUENTD_TAG"
	envKafkaBrokers                 = "APP_KAFKA_BROKERS"
	envKafkaTopic                   = "APP_KAFKA_TOPIC"
	envKafkaSerializationFormat     = "APP_KAFKA_SERIALIZATION_FORMAT"
	envKafkaMaxMessageBytes         = "APP_KAFKA_MAX_MESSAGE_BYTES"
	envKafkaCompressionType         = "APP_KAFKA_COMPRESSION_TYPE"
	envRouterConfigFile             = "ROUTER_CONFIG_FILE"
	envRouterProtocol               = "ROUTER_PROTOCOL"
	envGoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	envPluginName                   = "PLUGIN_NAME"
	envPluginsDir                   = "PLUGINS_DIR"
)

// Router service constants
const (
	routerPort                               = 8080
	routerLivenessPath                       = "/v1/internal/live"
	routerReadinessPath                      = "/v1/internal/ready"
	routerConfigFileName                     = "fiber.yml"
	routerConfigMapVolume                    = "config-map-volume"
	routerConfigMapMountPath                 = "/app/config/"
	routerConfigTypeCombiner                 = "COMBINER"
	routerConfigTypeEagerRouter              = "EAGER_ROUTER"
	routerConfigTypeLazyRouter               = "LAZY_ROUTER"
	routerConfigStrategyTypeDefault          = "fiber.DefaultTuringRoutingStrategy"
	routerConfigStrategyTypeFanIn            = "fiber.EnsemblingFanIn"
	routerConfigStrategyTypeTrafficSplitting = "fiber.TrafficSplittingStrategy"

	routerPluginBinaryConfigKey = "plugin_binary"
)

// Router endpoint constants
const (
	defaultIstioGatewayDestination = "istio-ingressgateway.istio-system.svc.cluster.local"
	// Warning given when using FQDN as Gateway
	defaultGateway = "knative-serving/knative-ingress-gateway"
)

// Plugins volume constants
const (
	pluginsMountPath  = "/app/plugins"
	pluginsVolumeName = "plugins-volume"
)

var defaultMatchURIPrefixes = []string{"/v1/predict", "/v1/batch_predict"}

// NewRouterService creates a new cluster Service object with the required config
// for the Turing router to be deployed.
func (sb *clusterSvcBuilder) NewRouterService(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	envType string,
	secretName string,
	experimentConfig json.RawMessage,
	routerDefaults *config.RouterDefaults,
	sentryEnabled bool,
	sentryDSN string,
	knativeQueueProxyResourcePercentage int,
	userContainerLimitRequestFactor float64,
) (*cluster.KnativeService, error) {
	// Create service name
	name := sb.GetRouterServiceName(routerVersion)
	// Namespace is the name of the project
	namespace := GetNamespace(project)

	configMap, err := buildFiberConfigMap(routerVersion, project, experimentConfig)
	if err != nil {
		return nil, err
	}

	volumes, volumeMounts := buildRouterVolumes(routerVersion, configMap.Name, secretName)

	initContainers := buildInitContainers(routerVersion)

	// Build env vars
	envs, err := sb.buildRouterEnvs(namespace, envType, routerDefaults,
		sentryEnabled, sentryDSN, secretName, routerVersion)
	if err != nil {
		return nil, err
	}
	svc := &cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:                 name,
			Namespace:            namespace,
			Image:                routerVersion.Image,
			CPURequests:          routerVersion.ResourceRequest.CPURequest,
			MemoryRequests:       routerVersion.ResourceRequest.MemoryRequest,
			LivenessHTTPGetPath:  routerLivenessPath,
			ReadinessHTTPGetPath: routerReadinessPath,
			Envs:                 envs,
			Labels:               buildLabels(project, routerVersion.Router),
			ConfigMap:            configMap,
			Volumes:              volumes,
			VolumeMounts:         volumeMounts,
			InitContainers:       initContainers,
		},
		IsClusterLocal:                  false,
		ContainerPort:                   routerPort,
		Protocol:                        routerVersion.Protocol,
		MinReplicas:                     routerVersion.ResourceRequest.MinReplica,
		MaxReplicas:                     routerVersion.ResourceRequest.MaxReplica,
		AutoscalingMetric:               string(routerVersion.AutoscalingPolicy.Metric),
		AutoscalingTarget:               routerVersion.AutoscalingPolicy.Target,
		QueueProxyResourcePercentage:    knativeQueueProxyResourcePercentage,
		UserContainerLimitRequestFactor: userContainerLimitRequestFactor,
	}
	return sb.validateKnativeService(svc)
}

func (sb *clusterSvcBuilder) NewRouterEndpoint(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	envType string,
	versionEndpoint string,
) (*cluster.VirtualService, error) {
	labels := buildLabels(project, routerVersion.Router)
	routerName := GetComponentName(routerVersion, ComponentTypes.Router)

	veURL, err := url.Parse(versionEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version endpoint url: %s", err.Error())
	}

	routerEndpointName := fmt.Sprintf("%s-turing-%s", routerVersion.Router.Name, ComponentTypes.Router)
	host := strings.Replace(veURL.Hostname(), routerName, routerEndpointName, 1)

	var matchURIPrefixes []string
	if routerVersion.Protocol == routeConfig.HTTP {
		matchURIPrefixes = defaultMatchURIPrefixes
	}

	return &cluster.VirtualService{
		Name:             routerEndpointName,
		Namespace:        project.Name,
		Labels:           labels,
		Gateway:          defaultGateway,
		Endpoint:         host,
		DestinationHost:  defaultIstioGatewayDestination,
		HostRewrite:      veURL.Hostname(),
		MatchURIPrefixes: matchURIPrefixes,
	}, nil
}

// GetRouterServiceName returns the name of the Router component, used by the Service
func (sb *clusterSvcBuilder) GetRouterServiceName(routerVersion *models.RouterVersion) string {
	return GetComponentName(routerVersion, ComponentTypes.Router)
}

func (sb *clusterSvcBuilder) buildRouterEnvs(
	namespace string,
	environmentType string,
	routerDefaults *config.RouterDefaults,
	sentryEnabled bool,
	sentryDSN string,
	secretName string,
	ver *models.RouterVersion,
) ([]corev1.EnvVar, error) {
	// Add app name, router timeout, jaeger collector
	envs := []corev1.EnvVar{
		{Name: envAppName, Value: fmt.Sprintf("%s-%d.%s", ver.Router.Name, ver.Version, namespace)},
		{Name: envAppEnvironment, Value: environmentType},
		{Name: envRouterTimeout, Value: ver.Timeout},
		{Name: envJaegerEndpoint, Value: routerDefaults.JaegerCollectorEndpoint},
		{Name: envRouterConfigFile, Value: routerConfigMapMountPath + routerConfigFileName},
		{Name: envRouterProtocol, Value: string(ver.Protocol)},
		{Name: envSentryEnabled, Value: strconv.FormatBool(sentryEnabled)},
		{Name: envSentryDSN, Value: sentryDSN},
	}

	// Add enricher / ensembler related env vars, if enabled
	if ver.Enricher != nil {
		endpoint := buildPrePostProcessorEndpoint(ver, namespace,
			ComponentTypes.Enricher, ver.Enricher.Endpoint)
		envs = append(envs, []corev1.EnvVar{
			{Name: envEnricherEndpoint, Value: endpoint},
			{Name: envEnricherTimeout, Value: ver.Enricher.Timeout},
		}...)
	}
	if ver.HasDockerConfig() {
		endpoint := buildPrePostProcessorEndpoint(
			ver,
			namespace,
			ComponentTypes.Ensembler,
			ver.Ensembler.DockerConfig.Endpoint,
		)
		envs = append(envs, []corev1.EnvVar{
			{Name: envEnsemblerEndpoint, Value: endpoint},
			{Name: envEnsemblerTimeout, Value: ver.Ensembler.DockerConfig.Timeout},
		}...)
	}

	// Process Log config
	logConfig := ver.LogConfig
	envs = append(envs, []corev1.EnvVar{
		{Name: envLogLevel, Value: string(logConfig.LogLevel)},
		{Name: envCustomMetrics, Value: strconv.FormatBool(logConfig.CustomMetricsEnabled)},
		{Name: envJaegerEnabled, Value: strconv.FormatBool(logConfig.JaegerEnabled)},
		{Name: envResultLogger, Value: string(logConfig.ResultLoggerType)},
		{Name: envFiberDebugLog, Value: strconv.FormatBool(logConfig.FiberDebugLogEnabled)},
	}...)

	// Add BQ config
	switch logConfig.ResultLoggerType {
	case models.BigQueryLogger:
		if logConfig.BigQueryConfig == nil {
			return envs, errors.New("Missing BigQuery logger config")
		}
		bqFQN := strings.Split(logConfig.BigQueryConfig.Table, ".")
		if len(bqFQN) != 3 {
			return envs, fmt.Errorf("Invalid BigQuery table name %s",
				logConfig.BigQueryConfig.Table)
		}
		envs = append(envs, []corev1.EnvVar{
			{Name: envGcpProject, Value: bqFQN[0]},
			{Name: envBQDataset, Value: bqFQN[1]},
			{Name: envBQTable, Value: bqFQN[2]},
			{Name: envBQBatchLoad, Value: strconv.FormatBool(logConfig.BigQueryConfig.BatchLoad)},
			{Name: envGoogleApplicationCredentials, Value: secretMountPath + secretKeyNameRouter},
		}...)
		if logConfig.BigQueryConfig.BatchLoad {
			envs = append(envs, []corev1.EnvVar{
				{Name: envFluentdHost, Value: buildFluentdHost(ver, namespace)},
				{Name: envFluentdPort, Value: strconv.Itoa(fluentdPort)},
				{Name: envFluentdTag, Value: routerDefaults.FluentdConfig.Tag},
			}...)
		}
	case models.KafkaLogger:
		envs = append(envs, []corev1.EnvVar{
			{Name: envKafkaBrokers, Value: logConfig.KafkaConfig.Brokers},
			{Name: envKafkaTopic, Value: logConfig.KafkaConfig.Topic},
			{Name: envKafkaSerializationFormat, Value: string(logConfig.KafkaConfig.SerializationFormat)},
			{Name: envKafkaMaxMessageBytes, Value: strconv.Itoa(routerDefaults.KafkaConfig.MaxMessageBytes)},
			{Name: envKafkaCompressionType, Value: routerDefaults.KafkaConfig.CompressionType},
		}...)
	}

	return envs, nil
}

func buildRouterVolumes(
	routerVersion *models.RouterVersion,
	configMapName string,
	secretName string,
) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes := make([]corev1.Volume, 0)
	volumeMounts := make([]corev1.VolumeMount, 0)

	// Router config
	volumes = append(volumes, corev1.Volume{
		Name: routerConfigMapVolume,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	})
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      routerConfigMapVolume,
		MountPath: routerConfigMapMountPath,
	})

	// Set up volume and volume mount if experiment engine plugin is set
	if routerVersion.ExperimentEngine.PluginConfig != nil {
		volumes = append(volumes, corev1.Volume{
			Name: pluginsVolumeName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      pluginsVolumeName,
			MountPath: pluginsMountPath,
		})
	}

	// Service account
	if routerVersion.LogConfig.ResultLoggerType == models.BigQueryLogger {
		volumes = append(volumes, corev1.Volume{
			Name: secretVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
					Items: []corev1.KeyToPath{
						{
							Key:  secretKeyNameRouter,
							Path: secretKeyNameRouter,
						},
					},
				},
			},
		})
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      secretVolume,
			MountPath: secretMountPath,
		})
	}
	return volumes, volumeMounts
}

func buildInitContainers(routerVersion *models.RouterVersion) []cluster.Container {
	// Set up initContainer if experiment engine plugin is set
	var initContainers []cluster.Container
	if routerVersion.ExperimentEngine.PluginConfig != nil {
		initContainers = make([]cluster.Container, 0)
		pluginContainer := cluster.Container{
			Name:  fmt.Sprintf("%s-plugin", routerVersion.ExperimentEngine.Type),
			Image: routerVersion.ExperimentEngine.PluginConfig.Image,
			Envs: []cluster.Env{
				{
					Name:  envPluginName,
					Value: routerVersion.ExperimentEngine.Type,
				},
				{
					Name:  envPluginsDir,
					Value: pluginsMountPath,
				},
			},
			VolumeMounts: []cluster.VolumeMount{
				{
					Name:      pluginsVolumeName,
					MountPath: pluginsMountPath,
				},
			},
		}
		initContainers = append(initContainers, pluginContainer)
	}

	return initContainers
}

func buildTrafficSplittingFiberConfig(
	name string,
	routes models.Routes,
	rules models.TrafficRules,
	ensembler *models.Ensembler,
	fiberProperties json.RawMessage,
	protocol fiberProtocol.Protocol,
) (fiberConfig.Config, error) {
	// IDs of routes, that are part of at least one traffic-splitting rule
	conditionalRouteIds := rules.ConditionalRouteIds()

	// routes, that are not part of any traffic-splitting rule
	alwaysActiveRoutes := models.Routes{}
	allRoutesMap := map[string]*models.Route{}
	for _, route := range routes {
		// build hash map to access route by its ID
		allRoutesMap[route.ID] = route

		// build slice of routes, that are not assigned to any
		// traffic-splitting rule and should be active for any request
		if _, exist := conditionalRouteIds[route.ID]; !exist {
			alwaysActiveRoutes = append(alwaysActiveRoutes, route)
		}
	}

	defaultRouteID := "traffic-split-default"

	// build fiber target component, that is active if request
	// doesn't match any of traffic rules
	defaultRouteConfig, err := buildFiberConfig(
		defaultRouteID,
		alwaysActiveRoutes,
		ensembler,
		fiberProperties,
		protocol)

	if err != nil {
		return nil, err
	}

	splitRoutes := []fiberConfig.Config{defaultRouteConfig}
	splitStrategy := fiberapi.TrafficSplittingStrategy{
		DefaultRouteID: defaultRouteID,
		Rules:          nil,
	}

	// iterate over traffic rules and generated nested routed and
	// config for traffic splitting routing strategy
	for idx, rule := range rules {
		routeID := fmt.Sprintf("traffic-split-%d", idx)

		var ruleRoutes models.Routes
		for _, rID := range rule.Routes {
			if route, exist := allRoutesMap[rID]; exist {
				ruleRoutes = append(ruleRoutes, route)
			} else {
				return nil, fmt.Errorf("failed to build fiber config, unknown route id: %s", rID)
			}
		}

		// build nested fiber component with routes, activated by this rule
		routeConfig, err := buildFiberConfig(
			routeID,
			append(alwaysActiveRoutes, ruleRoutes...),
			ensembler,
			fiberProperties,
			protocol)

		if err != nil {
			return nil, err
		}

		// append new route to the top-level traffic splitting router
		splitRoutes = append(splitRoutes, routeConfig)

		// append new rule to the traffic splitting strategy
		splitStrategy.Rules = append(
			splitStrategy.Rules,
			&fiberapi.TrafficSplittingStrategyRule{
				RouteID:    routeID,
				Conditions: rule.Conditions,
			})
	}

	// serialize properties of traffic-splitting strategy
	splitStrategyProps, err := json.Marshal(&splitStrategy)
	if err != nil {
		return nil, err
	}

	routerConfig := &fiberConfig.RouterConfig{
		MultiRouteConfig: fiberConfig.MultiRouteConfig{
			ComponentConfig: fiberConfig.ComponentConfig{
				ID:   name,
				Type: routerConfigTypeLazyRouter,
			},
			Routes: splitRoutes,
		},
		Strategy: fiberConfig.StrategyConfig{
			Type:       routerConfigStrategyTypeTrafficSplitting,
			Properties: splitStrategyProps,
		},
	}

	return routerConfig, nil
}

func buildFiberConfig(
	name string,
	routes models.Routes,
	ensembler *models.Ensembler,
	fiberProperties json.RawMessage,
	protocol fiberProtocol.Protocol,
) (fiberConfig.Config, error) {
	// Create the MultiRouteConfig
	fiberRoutes, err := routes.ToFiberRoutes(protocol)
	if err != nil {
		return nil, err
	}
	multiRouteConfig := fiberConfig.MultiRouteConfig{
		ComponentConfig: fiberConfig.ComponentConfig{
			ID: name,
		},
		Routes: *fiberRoutes,
	}

	// Select router type (eager or combiner) based on the ensembler config.
	// If ensembler uses a DockerConfig to run, use "combiner" router
	// Else, "eager" router is used.
	var routerConfig fiberConfig.Config
	if ensembler != nil && ensembler.DockerConfig != nil {
		multiRouteConfig.Type = routerConfigTypeCombiner
		routerConfig = &fiberConfig.CombinerConfig{
			MultiRouteConfig: multiRouteConfig,
			FanIn: fiberConfig.FanInConfig{
				Type:       routerConfigStrategyTypeFanIn,
				Properties: fiberProperties,
			},
		}
	} else {
		multiRouteConfig.Type = routerConfigTypeEagerRouter
		routerConfig = &fiberConfig.RouterConfig{
			MultiRouteConfig: multiRouteConfig,
			Strategy: fiberConfig.StrategyConfig{
				Type:       routerConfigStrategyTypeDefault,
				Properties: fiberProperties,
			},
		}
	}

	return routerConfig, nil
}

func buildFiberConfigMap(
	ver *models.RouterVersion,
	project *mlp.Project,
	experimentCfg json.RawMessage,
) (*cluster.ConfigMap, error) {
	// Create the properties map for fiber's routing strategy or fanIn
	propsMap := map[string]interface{}{
		"experiment_engine": ver.ExperimentEngine.Type,
	}
	if ver.DefaultRouteID != "" {
		propsMap["default_route_id"] = ver.DefaultRouteID
	}
	if ver.ExperimentEngine.Type != models.ExperimentEngineTypeNop {
		expEngineProps := experimentCfg
		// Tell router, that the experiment runner is implemented as RPC plugin
		if ver.ExperimentEngine.PluginConfig != nil {
			var err error
			pluginBinary := fmt.Sprintf(
				"%s/%s", pluginsMountPath, ver.ExperimentEngine.Type)
			expEngineProps, err = utils.MergeJSON(
				expEngineProps,
				map[string]interface{}{routerPluginBinaryConfigKey: pluginBinary},
			)
			if err != nil {
				return nil, err
			}
		}
		propsMap["experiment_engine_properties"] = expEngineProps
	}

	if ver.Ensembler != nil && ver.Ensembler.Type == models.EnsemblerStandardType {
		if ver.Ensembler.StandardConfig.ExperimentMappings != nil &&
			len(ver.Ensembler.StandardConfig.ExperimentMappings) != 0 {
			propsMap["experiment_mappings"] = ver.Ensembler.StandardConfig.ExperimentMappings
		}
		if ver.Ensembler.StandardConfig.RouteNamePath != "" {
			propsMap["route_name_path"] = ver.Ensembler.StandardConfig.RouteNamePath
		}
	}

	properties, err := json.Marshal(propsMap)
	if err != nil {
		return nil, err
	}

	// default to http
	routeProtocol := fiberProtocol.HTTP
	if ver.Protocol == routeConfig.UPI {
		routeProtocol = fiberProtocol.GRPC
	}

	var routerConfig fiberConfig.Config
	// if the version is configured with traffic splitting rules on it,
	// then define root-level fiber component as a lazy router with
	// a traffic-splitting strategy based on these rules
	if ver.TrafficRules != nil && len(ver.TrafficRules) > 0 {
		// TrafficRule struct used requires the name and conditions field to be specified. But
		// Default Traffic Rule has no name and a hardcoded name can be used instead since
		// the name field is not used for traffic splitting strategy. Likewise, an empty slice
		// of conditions can be used for the same reason.
		rules := append(
			ver.TrafficRules,
			&models.TrafficRule{
				Name:       "default-traffic-rule",
				Conditions: []*router.TrafficRuleCondition{},
				Routes:     ver.DefaultTrafficRule.Routes,
			},
		)
		routerConfig, err = buildTrafficSplittingFiberConfig(
			ver.Router.Name,
			ver.Routes,
			rules,
			ver.Ensembler,
			properties,
			routeProtocol)
	} else {
		routerConfig, err = buildFiberConfig(
			ver.Router.Name,
			ver.Routes,
			ver.Ensembler,
			properties,
			routeProtocol)
	}

	if err != nil {
		return nil, err
	}

	configMapData, err := yaml.Marshal(routerConfig)

	if err != nil {
		return nil, err
	}

	return &cluster.ConfigMap{
		Name:     GetComponentName(ver, ComponentTypes.FiberConfig),
		FileName: routerConfigFileName,
		Data:     string(configMapData),
		Labels:   buildLabels(project, ver.Router),
	}, nil
}

func buildPrePostProcessorEndpoint(
	routerVersion *models.RouterVersion,
	namespace string,
	componentType string,
	relativeEndpoint string,
) string {
	componentName := GetComponentName(routerVersion, componentType)
	// Trim leading slash, if present
	relativeEndpoint = strings.TrimPrefix(relativeEndpoint, "/")
	return fmt.Sprintf("http://%s.%s.svc.cluster.local/%s",
		componentName, namespace, relativeEndpoint)
}
