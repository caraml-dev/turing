package servicebuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gojek/turing/engines/router/missionctl/fiberapi"

	"net/url"

	"github.com/ghodss/yaml"
	fiberconfig "github.com/gojek/fiber/config"
	mlp "github.com/gojek/mlp/api/client"
	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/models"
	corev1 "k8s.io/api/core/v1"
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
	envLitmusPasskey                = "LITMUS_PASSKEY"
	envXpPasskey                    = "XP_PASSKEY"
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
	envRouterConfigFile             = "ROUTER_CONFIG_FILE"
	envGoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
)

// router service constants
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
)

// Router endpoint constants
const (
	defaultIstioGateway = "istio-ingressgateway.istio-system.svc.cluster.local"
	defaultGateway      = "knative-ingress-gateway.knative-serving"
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
	fluentdTag string,
	jaegerCollectorEndpoint string,
	sentryEnabled bool,
	sentryDSN string,
	knativeTargetConcurrency int,
	knativeQueueProxyResourcePercentage int,
) (*cluster.KnativeService, error) {
	// Create service name
	name := sb.GetRouterServiceName(routerVersion)
	// Namespace is the name of the project
	namespace := GetNamespace(project)

	configMap, err := buildFiberConfigMap(routerVersion, experimentConfig)
	if err != nil {
		return nil, err
	}

	volumes, volumeMounts := buildRouterVolumes(routerVersion, configMap.Name, secretName)

	// Build env vars
	envs, err := sb.buildRouterEnvs(namespace, envType, fluentdTag, jaegerCollectorEndpoint,
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
			Labels:               buildLabels(project, envType, routerVersion.Router),
			ConfigMap:            configMap,
			Volumes:              volumes,
			VolumeMounts:         volumeMounts,
		},
		IsClusterLocal:               false,
		ContainerPort:                routerPort,
		MinReplicas:                  routerVersion.ResourceRequest.MinReplica,
		MaxReplicas:                  routerVersion.ResourceRequest.MaxReplica,
		TargetConcurrency:            knativeTargetConcurrency,
		QueueProxyResourcePercentage: knativeQueueProxyResourcePercentage,
	}
	return sb.validateKnativeService(svc)
}

func (sb *clusterSvcBuilder) NewRouterEndpoint(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	envType string,
	versionEndpoint string,
) (*cluster.VirtualService, error) {
	labels := buildLabels(project, envType, routerVersion.Router)
	routerName := GetComponentName(routerVersion, ComponentTypes.Router)

	veURL, err := url.Parse(versionEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse version endpoint url: %s", err.Error())
	}

	routerEndpointName := fmt.Sprintf("%s-turing-%s", routerVersion.Router.Name, ComponentTypes.Router)
	host := strings.Replace(veURL.Hostname(), routerName, routerEndpointName, 1)

	return &cluster.VirtualService{
		Name:             routerEndpointName,
		Namespace:        project.Name,
		Labels:           labels,
		Gateway:          defaultGateway,
		Endpoint:         host,
		DestinationHost:  defaultIstioGateway,
		HostRewrite:      veURL.Hostname(),
		MatchURIPrefixes: defaultMatchURIPrefixes,
	}, nil
}

// GetRouterServiceName returns the name of the Router component, used by the Service
func (sb *clusterSvcBuilder) GetRouterServiceName(routerVersion *models.RouterVersion) string {
	return GetComponentName(routerVersion, ComponentTypes.Router)
}

func (sb *clusterSvcBuilder) buildRouterEnvs(
	namespace string,
	environmentType string,
	fluentdTag string,
	jaegerCollectorEndpoint string,
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
		{Name: envJaegerEndpoint, Value: jaegerCollectorEndpoint},
		{Name: envRouterConfigFile, Value: routerConfigMapMountPath + routerConfigFileName},
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
	if ver.Ensembler != nil && ver.Ensembler.Type == models.EnsemblerDockerType {
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

	// Add Experiment Engine config
	if ver.ExperimentEngine.Type != models.ExperimentEngineTypeNop {
		var envVarName string
		switch ver.ExperimentEngine.Type {
		case models.ExperimentEngineTypeLitmus:
			envVarName = envLitmusPasskey
		case models.ExperimentEngineTypeXp:
			envVarName = envXpPasskey
		}
		// Add env var
		envs = append(envs, corev1.EnvVar{
			Name: envVarName,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretName,
					},
					Key: secretKeyNameExperiment,
				},
			},
		})
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
				{Name: envFluentdTag, Value: fluentdTag},
			}...)
		}
	case models.KafkaLogger:
		envs = append(envs, []corev1.EnvVar{
			{Name: envKafkaBrokers, Value: logConfig.KafkaConfig.Brokers},
			{Name: envKafkaTopic, Value: logConfig.KafkaConfig.Topic},
			{Name: envKafkaSerializationFormat, Value: string(logConfig.KafkaConfig.SerializationFormat)},
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

func buildTrafficSplittingFiberConfig(
	name string,
	routes models.Routes,
	rules models.TrafficRules,
	ensembler *models.Ensembler,
	fiberProperties json.RawMessage,
) (fiberconfig.Config, error) {
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
		fiberProperties)

	if err != nil {
		return nil, err
	}

	splitRoutes := []fiberconfig.Config{defaultRouteConfig}
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
			fiberProperties)

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

	routerConfig := &fiberconfig.RouterConfig{
		MultiRouteConfig: fiberconfig.MultiRouteConfig{
			ComponentConfig: fiberconfig.ComponentConfig{
				ID:   name,
				Type: routerConfigTypeLazyRouter,
			},
			Routes: splitRoutes,
		},
		Strategy: fiberconfig.StrategyConfig{
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
) (fiberconfig.Config, error) {
	// Create the MultiRouteConfig
	fiberRoutes, err := routes.ToFiberRoutes()
	if err != nil {
		return nil, err
	}
	multiRouteConfig := fiberconfig.MultiRouteConfig{
		ComponentConfig: fiberconfig.ComponentConfig{
			ID: name,
		},
		Routes: *fiberRoutes,
	}

	// Select router type (eager or combiner) based on the ensembler config.
	// If ensembler is set and is of Docker type, use "combiner" router
	// Else, "eager" router is used.
	var routerConfig fiberconfig.Config
	if ensembler != nil && ensembler.Type == models.EnsemblerDockerType {
		multiRouteConfig.Type = routerConfigTypeCombiner
		routerConfig = &fiberconfig.CombinerConfig{
			MultiRouteConfig: multiRouteConfig,
			FanIn: fiberconfig.FanInConfig{
				Type:       routerConfigStrategyTypeFanIn,
				Properties: fiberProperties,
			},
		}
	} else {
		multiRouteConfig.Type = routerConfigTypeEagerRouter
		routerConfig = &fiberconfig.RouterConfig{
			MultiRouteConfig: multiRouteConfig,
			Strategy: fiberconfig.StrategyConfig{
				Type:       routerConfigStrategyTypeDefault,
				Properties: fiberProperties,
			},
		}
	}

	return routerConfig, nil
}

func buildFiberConfigMap(ver *models.RouterVersion, experimentCfg json.RawMessage) (*cluster.ConfigMap, error) {
	// Create the properties map for fiber's routing strategy or fanIn
	propsMap := map[string]interface{}{
		"default_route_id":  ver.DefaultRouteID,
		"experiment_engine": string(ver.ExperimentEngine.Type),
	}
	if ver.ExperimentEngine.Type != models.ExperimentEngineTypeNop {
		propsMap["experiment_engine_properties"] = experimentCfg
	}

	if ver.Ensembler != nil && ver.Ensembler.Type == models.EnsemblerStandardType {
		propsMap["experiment_mappings"] = ver.Ensembler.StandardConfig.ExperimentMappings
	}

	properties, err := json.Marshal(propsMap)
	if err != nil {
		return nil, err
	}

	var routerConfig fiberconfig.Config
	// if the version is configured with traffic splitting rules on it,
	// then define root-level fiber component as a lazy router with
	// a traffic-splitting strategy based on these rules
	if ver.TrafficRules != nil && len(ver.TrafficRules) > 0 {
		routerConfig, err = buildTrafficSplittingFiberConfig(
			ver.Router.Name,
			ver.Routes,
			ver.TrafficRules,
			ver.Ensembler,
			properties)
	} else {
		routerConfig, err = buildFiberConfig(
			ver.Router.Name,
			ver.Routes,
			ver.Ensembler,
			properties)
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
	}, nil
}

func buildFluentdHost(
	routerVersion *models.RouterVersion,
	namespace string,
) string {
	componentName := GetComponentName(routerVersion, ComponentTypes.FluentdLogger)
	return fmt.Sprintf("%s.%s.svc.cluster.local", componentName, namespace)
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
