package servicebuilder

import (
	"encoding/json"
	"path/filepath"
	"testing"

	corev1 "k8s.io/api/core/v1"

	mlp "github.com/gojek/mlp/api/client"
	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/config"
	tu "github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/gojek/turing/api/turing/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNewRouterService(t *testing.T) {
	sb := NewClusterServiceBuilder(resource.MustParse("2"), resource.MustParse("2Gi"))
	testDataBasePath := filepath.Join("..", "..", "testdata", "cluster", "servicebuilder")
	enrEndpoint := "http://test-svc-turing-enricher-1.test-project.svc.cluster.local/echo?delay=10ms"
	ensEndpoint := "http://test-svc-turing-ensembler-1.test-project.svc.cluster.local/echo?delay=20ms"
	expRunnerConfig := `{
		"client_id": "client_id",
		"endpoint": "exp-engine:8080",
		"experiments": [
			{
				"experiment_name": "exp_exp_test_experiment_1",
				"segmentation_field": "customer_id",
				"segmentation_field_source": "payload",
				"segmentation_unit": "customer"
			}
		],
		"timeout": "500ms",
		"user_data": {
			"app_version": {
				"field": "appVer",
				"field_source": "header"
			}
		}
	}`
	// Read configmap test data
	cfgmapDefault, err := tu.ReadFile(filepath.Join(testDataBasePath, "router_configmap_default.yml"))
	require.NoError(t, err)
	cfgmapEnsembling, err := tu.ReadFile(filepath.Join(testDataBasePath, "router_configmap_ensembling.yml"))
	require.NoError(t, err)
	cfgmapStandardEnsemble, err := tu.ReadFile(filepath.Join(testDataBasePath, "router_configmap_standard_ensembler.yml"))
	require.NoError(t, err)
	cfgmapTrafficSplitting, err := tu.ReadFile(filepath.Join(testDataBasePath, "router_configmap_traffic_splitting.yml"))
	require.NoError(t, err)
	cfgmapExpEngine, err := tu.ReadFile(filepath.Join(testDataBasePath, "router_configmap_exp_engine.yml"))
	require.NoError(t, err)

	// Define tests
	tests := map[string]testSuiteNewService{
		"success | basic": {
			filePath:     filepath.Join(testDataBasePath, "router_version_basic.json"),
			expRawConfig: json.RawMessage(expRunnerConfig),
			expected: &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					Name:                 "test-svc-turing-router-1",
					Namespace:            "test-project",
					Image:                "asia.gcr.io/gcp-project-id/turing-router:latest",
					CPURequests:          resource.MustParse("400m"),
					MemoryRequests:       resource.MustParse("512Mi"),
					LivenessHTTPGetPath:  "/v1/internal/live",
					ReadinessHTTPGetPath: "/v1/internal/ready",
					ConfigMap: &cluster.ConfigMap{
						Name:     "test-svc-turing-fiber-config-1",
						FileName: "fiber.yml",
						Data:     string(cfgmapDefault),
						Labels: map[string]string{
							"app":          "test-svc",
							"environment":  "",
							"orchestrator": "turing",
							"stream":       "test-stream",
							"team":         "test-team",
						},
					},
					Envs: []corev1.EnvVar{
						{Name: "APP_NAME", Value: "test-svc-1.test-project"},
						{Name: "APP_ENVIRONMENT", Value: "test-env"},
						{Name: "ROUTER_TIMEOUT", Value: "5s"},
						{Name: "APP_JAEGER_COLLECTOR_ENDPOINT", Value: "jaeger-endpoint"},
						{Name: "ROUTER_CONFIG_FILE", Value: "/app/config/fiber.yml"},
						{Name: "APP_SENTRY_ENABLED", Value: "true"},
						{Name: "APP_SENTRY_DSN", Value: "sentry-dsn"},
						{Name: "APP_LOGLEVEL", Value: "INFO"},
						{Name: "APP_CUSTOM_METRICS", Value: "false"},
						{Name: "APP_JAEGER_ENABLED", Value: "false"},
						{Name: "APP_RESULT_LOGGER", Value: "bigquery"},
						{Name: "APP_FIBER_DEBUG_LOG", Value: "false"},
						{Name: "APP_GCP_PROJECT", Value: "gcp-project-id"},
						{Name: "APP_BQ_DATASET", Value: "dataset_id"},
						{Name: "APP_BQ_TABLE", Value: "turing_log_test"},
						{Name: "APP_BQ_BATCH_LOAD", Value: "false"},
						{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "/var/secret/router-service-account.json"},
					},
					Labels: map[string]string{
						"app":          "test-svc",
						"environment":  "",
						"orchestrator": "turing",
						"stream":       "test-stream",
						"team":         "test-team",
					},
					Volumes: []corev1.Volume{
						{
							Name: routerConfigMapVolume,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "test-svc-turing-fiber-config-1",
									},
								},
							},
						},
						{
							Name: secretVolume,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "service-account",
									Items: []corev1.KeyToPath{
										{
											Key:  secretKeyNameRouter,
											Path: secretKeyNameRouter,
										},
									},
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      routerConfigMapVolume,
							MountPath: routerConfigMapMountPath,
						},
						{
							Name:      secretVolume,
							MountPath: secretMountPath,
						},
					},
				},
				ContainerPort:                   8080,
				MinReplicas:                     2,
				MaxReplicas:                     4,
				TargetConcurrency:               1,
				QueueProxyResourcePercentage:    20,
				UserContainerLimitRequestFactor: 1.5,
			},
		},
		"success | all components": {
			filePath:     filepath.Join(testDataBasePath, "router_version_success.json"),
			expRawConfig: json.RawMessage(`{}`),
			expected: &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					Name:                 "test-svc-turing-router-1",
					Namespace:            "test-project",
					Image:                "asia.gcr.io/gcp-project-id/turing-router:latest",
					CPURequests:          resource.MustParse("400m"),
					MemoryRequests:       resource.MustParse("512Mi"),
					LivenessHTTPGetPath:  "/v1/internal/live",
					ReadinessHTTPGetPath: "/v1/internal/ready",
					ConfigMap: &cluster.ConfigMap{
						Name:     "test-svc-turing-fiber-config-1",
						FileName: "fiber.yml",
						Data:     string(cfgmapEnsembling),
						Labels: map[string]string{
							"app":          "test-svc",
							"environment":  "",
							"orchestrator": "turing",
							"stream":       "test-stream",
							"team":         "test-team",
						},
					},
					Envs: []corev1.EnvVar{
						{Name: "APP_NAME", Value: "test-svc-1.test-project"},
						{Name: "APP_ENVIRONMENT", Value: "test-env"},
						{Name: "ROUTER_TIMEOUT", Value: "5s"},
						{Name: "APP_JAEGER_COLLECTOR_ENDPOINT", Value: "jaeger-endpoint"},
						{Name: "ROUTER_CONFIG_FILE", Value: "/app/config/fiber.yml"},
						{Name: "APP_SENTRY_ENABLED", Value: "true"},
						{Name: "APP_SENTRY_DSN", Value: "sentry-dsn"},
						{Name: "ENRICHER_ENDPOINT", Value: enrEndpoint},
						{Name: "ENRICHER_TIMEOUT", Value: "2s"},
						{Name: "ENSEMBLER_ENDPOINT", Value: ensEndpoint},
						{Name: "ENSEMBLER_TIMEOUT", Value: "3s"},
						{Name: "APP_LOGLEVEL", Value: "INFO"},
						{Name: "APP_CUSTOM_METRICS", Value: "false"},
						{Name: "APP_JAEGER_ENABLED", Value: "false"},
						{Name: "APP_RESULT_LOGGER", Value: "bigquery"},
						{Name: "APP_FIBER_DEBUG_LOG", Value: "false"},
						{Name: "APP_GCP_PROJECT", Value: "gcp-project-id"},
						{Name: "APP_BQ_DATASET", Value: "dataset_id"},
						{Name: "APP_BQ_TABLE", Value: "turing_log_test"},
						{Name: "APP_BQ_BATCH_LOAD", Value: "true"},
						{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "/var/secret/router-service-account.json"},
						{Name: "APP_FLUENTD_HOST",
							Value: "test-svc-turing-fluentd-logger-1.test-project.svc.cluster.local"},
						{Name: "APP_FLUENTD_PORT", Value: "24224"},
						{Name: "APP_FLUENTD_TAG", Value: "fluentd-tag"},
					},
					Labels: map[string]string{
						"app":          "test-svc",
						"environment":  "",
						"orchestrator": "turing",
						"stream":       "test-stream",
						"team":         "test-team",
					},
					Volumes: []corev1.Volume{
						{
							Name: routerConfigMapVolume,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "test-svc-turing-fiber-config-1",
									},
								},
							},
						},
						{
							Name: secretVolume,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "service-account",
									Items: []corev1.KeyToPath{
										{
											Key:  secretKeyNameRouter,
											Path: secretKeyNameRouter,
										},
									},
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      routerConfigMapVolume,
							MountPath: routerConfigMapMountPath,
						},
						{
							Name:      secretVolume,
							MountPath: secretMountPath,
						},
					},
				},
				ContainerPort:                   8080,
				MinReplicas:                     2,
				MaxReplicas:                     4,
				TargetConcurrency:               1,
				QueueProxyResourcePercentage:    20,
				UserContainerLimitRequestFactor: 1.5,
			},
		},
		"success | standard ensembler": {
			filePath:     filepath.Join(testDataBasePath, "router_version_success_standard_ensembler.json"),
			expRawConfig: json.RawMessage(expRunnerConfig),
			expected: &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					Name:                 "test-svc-turing-router-1",
					Namespace:            "test-project",
					Image:                "asia.gcr.io/gcp-project-id/turing-router:latest",
					CPURequests:          resource.MustParse("400m"),
					MemoryRequests:       resource.MustParse("512Mi"),
					LivenessHTTPGetPath:  "/v1/internal/live",
					ReadinessHTTPGetPath: "/v1/internal/ready",
					ConfigMap: &cluster.ConfigMap{
						Name:     "test-svc-turing-fiber-config-1",
						FileName: "fiber.yml",
						Data:     string(cfgmapStandardEnsemble),
						Labels: map[string]string{
							"app":          "test-svc",
							"environment":  "",
							"orchestrator": "turing",
							"stream":       "test-stream",
							"team":         "test-team",
						},
					},
					Envs: []corev1.EnvVar{
						{Name: "APP_NAME", Value: "test-svc-1.test-project"},
						{Name: "APP_ENVIRONMENT", Value: "test-env"},
						{Name: "ROUTER_TIMEOUT", Value: "5s"},
						{Name: "APP_JAEGER_COLLECTOR_ENDPOINT", Value: "jaeger-endpoint"},
						{Name: "ROUTER_CONFIG_FILE", Value: "/app/config/fiber.yml"},
						{Name: "APP_SENTRY_ENABLED", Value: "true"},
						{Name: "APP_SENTRY_DSN", Value: "sentry-dsn"},
						{Name: "APP_LOGLEVEL", Value: "INFO"},
						{Name: "APP_CUSTOM_METRICS", Value: "false"},
						{Name: "APP_JAEGER_ENABLED", Value: "false"},
						{Name: "APP_RESULT_LOGGER", Value: "bigquery"},
						{Name: "APP_FIBER_DEBUG_LOG", Value: "false"},
						{Name: "APP_GCP_PROJECT", Value: "gcp-project-id"},
						{Name: "APP_BQ_DATASET", Value: "dataset_id"},
						{Name: "APP_BQ_TABLE", Value: "turing_log_test"},
						{Name: "APP_BQ_BATCH_LOAD", Value: "false"},
						{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "/var/secret/router-service-account.json"},
					},
					Labels: map[string]string{
						"app":          "test-svc",
						"environment":  "",
						"orchestrator": "turing",
						"stream":       "test-stream",
						"team":         "test-team",
					},
					Volumes: []corev1.Volume{
						{
							Name: routerConfigMapVolume,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "test-svc-turing-fiber-config-1",
									},
								},
							},
						},
						{
							Name: secretVolume,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "service-account",
									Items: []corev1.KeyToPath{
										{
											Key:  secretKeyNameRouter,
											Path: secretKeyNameRouter,
										},
									},
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      routerConfigMapVolume,
							MountPath: routerConfigMapMountPath,
						},
						{
							Name:      secretVolume,
							MountPath: secretMountPath,
						},
					},
				},
				ContainerPort:                   8080,
				MinReplicas:                     2,
				MaxReplicas:                     4,
				TargetConcurrency:               1,
				QueueProxyResourcePercentage:    20,
				UserContainerLimitRequestFactor: 1.5,
			},
		},
		"success | traffic-splitting": {
			filePath:     filepath.Join(testDataBasePath, "router_version_success_traffic_splitting.json"),
			expRawConfig: json.RawMessage(expRunnerConfig),
			expected: &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					Name:                 "test-svc-turing-router-1",
					Namespace:            "test-project",
					Image:                "asia.gcr.io/gcp-project-id/turing-router:latest",
					CPURequests:          resource.MustParse("400m"),
					MemoryRequests:       resource.MustParse("512Mi"),
					LivenessHTTPGetPath:  "/v1/internal/live",
					ReadinessHTTPGetPath: "/v1/internal/ready",
					ConfigMap: &cluster.ConfigMap{
						Name:     "test-svc-turing-fiber-config-1",
						FileName: "fiber.yml",
						Data:     string(cfgmapTrafficSplitting),
						Labels: map[string]string{
							"app":          "test-svc",
							"environment":  "",
							"orchestrator": "turing",
							"stream":       "test-stream",
							"team":         "test-team",
						},
					},
					Envs: []corev1.EnvVar{
						{Name: "APP_NAME", Value: "test-svc-1.test-project"},
						{Name: "APP_ENVIRONMENT", Value: "test-env"},
						{Name: "ROUTER_TIMEOUT", Value: "5s"},
						{Name: "APP_JAEGER_COLLECTOR_ENDPOINT", Value: "jaeger-endpoint"},
						{Name: "ROUTER_CONFIG_FILE", Value: "/app/config/fiber.yml"},
						{Name: "APP_SENTRY_ENABLED", Value: "true"},
						{Name: "APP_SENTRY_DSN", Value: "sentry-dsn"},
						{Name: "APP_LOGLEVEL", Value: "INFO"},
						{Name: "APP_CUSTOM_METRICS", Value: "false"},
						{Name: "APP_JAEGER_ENABLED", Value: "false"},
						{Name: "APP_RESULT_LOGGER", Value: "bigquery"},
						{Name: "APP_FIBER_DEBUG_LOG", Value: "false"},
						{Name: "APP_GCP_PROJECT", Value: "gcp-project-id"},
						{Name: "APP_BQ_DATASET", Value: "dataset_id"},
						{Name: "APP_BQ_TABLE", Value: "turing_log_test"},
						{Name: "APP_BQ_BATCH_LOAD", Value: "false"},
						{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "/var/secret/router-service-account.json"},
					},
					Labels: map[string]string{
						"app":          "test-svc",
						"environment":  "",
						"orchestrator": "turing",
						"stream":       "test-stream",
						"team":         "test-team",
					},
					Volumes: []corev1.Volume{
						{
							Name: routerConfigMapVolume,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "test-svc-turing-fiber-config-1",
									},
								},
							},
						},
						{
							Name: secretVolume,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "service-account",
									Items: []corev1.KeyToPath{
										{
											Key:  secretKeyNameRouter,
											Path: secretKeyNameRouter,
										},
									},
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      routerConfigMapVolume,
							MountPath: routerConfigMapMountPath,
						},
						{
							Name:      secretVolume,
							MountPath: secretMountPath,
						},
					},
				},
				ContainerPort:                   8080,
				MinReplicas:                     2,
				MaxReplicas:                     4,
				TargetConcurrency:               1,
				QueueProxyResourcePercentage:    20,
				UserContainerLimitRequestFactor: 1.5,
			},
		},
		"success | experiment engine": {
			filePath:     filepath.Join(testDataBasePath, "router_version_success_experiment_engine.json"),
			expRawConfig: json.RawMessage(`{"key-1": "value-1"}`),
			expected: &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					Name:                 "router-with-exp-engine-turing-router-1",
					Namespace:            "test-project",
					Image:                "ghcr.io/gojek/turing/turing-router:latest",
					CPURequests:          resource.MustParse("400m"),
					MemoryRequests:       resource.MustParse("512Mi"),
					LivenessHTTPGetPath:  "/v1/internal/live",
					ReadinessHTTPGetPath: "/v1/internal/ready",
					ConfigMap: &cluster.ConfigMap{
						Name:     "router-with-exp-engine-turing-fiber-config-1",
						FileName: "fiber.yml",
						Data:     string(cfgmapExpEngine),
						Labels: map[string]string{
							"app":          "router-with-exp-engine",
							"environment":  "",
							"orchestrator": "turing",
							"stream":       "test-stream",
							"team":         "test-team",
						},
					},
					Envs: []corev1.EnvVar{
						{Name: "APP_NAME", Value: "router-with-exp-engine-1.test-project"},
						{Name: "APP_ENVIRONMENT", Value: "test-env"},
						{Name: "ROUTER_TIMEOUT", Value: "5s"},
						{Name: "APP_JAEGER_COLLECTOR_ENDPOINT", Value: "jaeger-endpoint"},
						{Name: "ROUTER_CONFIG_FILE", Value: "/app/config/fiber.yml"},
						{Name: "APP_SENTRY_ENABLED", Value: "true"},
						{Name: "APP_SENTRY_DSN", Value: "sentry-dsn"},
						{Name: "APP_LOGLEVEL", Value: "INFO"},
						{Name: "APP_CUSTOM_METRICS", Value: "false"},
						{Name: "APP_JAEGER_ENABLED", Value: "false"},
						{Name: "APP_RESULT_LOGGER", Value: "nop"},
						{Name: "APP_FIBER_DEBUG_LOG", Value: "false"},
					},
					Labels: map[string]string{
						"app":          "router-with-exp-engine",
						"environment":  "",
						"orchestrator": "turing",
						"stream":       "test-stream",
						"team":         "test-team",
					},
					Volumes: []corev1.Volume{
						{
							Name: routerConfigMapVolume,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "router-with-exp-engine-turing-fiber-config-1",
									},
								},
							},
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      routerConfigMapVolume,
							MountPath: routerConfigMapMountPath,
						},
					},
				},
				ContainerPort:                   8080,
				MinReplicas:                     2,
				MaxReplicas:                     4,
				TargetConcurrency:               1,
				QueueProxyResourcePercentage:    20,
				UserContainerLimitRequestFactor: 1.5,
			},
		},
		"failure missing bigquery": {
			filePath: filepath.Join(testDataBasePath, "router_version_missing_bigquery.json"),
			err:      "Missing BigQuery logger config",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Read router version test data
			fileBytes, err := tu.ReadFile(data.filePath)
			require.NoError(t, err)
			// Convert to RouterVersion type
			var routerVersion models.RouterVersion
			err = json.Unmarshal(fileBytes, &routerVersion)
			require.NoError(t, err)

			// Run test
			project := &mlp.Project{
				Name:   "test-project",
				Stream: "test-stream",
				Team:   "test-team",
			}
			svc, err := sb.NewRouterService(
				&routerVersion,
				project,
				"test-env",
				"service-account",
				data.expRawConfig,
				&config.RouterDefaults{
					JaegerCollectorEndpoint: "jaeger-endpoint",
					FluentdConfig:           &config.FluentdConfig{Tag: "fluentd-tag"},
				},
				true,
				"sentry-dsn",
				1,
				20,
				1.5)

			if data.err == "" {
				require.NoError(t, err)
				assert.Equal(t, data.expected, svc)
			} else {
				assert.EqualError(t, err, data.err)
			}
		})
	}
}

func TestNewRouterEndpoint(t *testing.T) {
	// Get router version
	sb := NewClusterServiceBuilder(resource.MustParse("2"), resource.MustParse("2Gi"))
	testDataBasePath := filepath.Join("..", "..", "testdata", "cluster", "servicebuilder")
	fileBytes, err := tu.ReadFile(filepath.Join(testDataBasePath, "router_version_success.json"))
	require.NoError(t, err)
	var routerVersion models.RouterVersion
	err = json.Unmarshal(fileBytes, &routerVersion)
	require.NoError(t, err)

	project := &mlp.Project{
		Name:   "test-project",
		Stream: "test-stream",
		Team:   "test-team",
	}

	versionEndpoint := "http://test-svc-turing-router-1.models.example.com"

	expected := &cluster.VirtualService{
		Name:      "test-svc-turing-router",
		Namespace: project.Name,
		Labels: map[string]string{
			"app":          "test-svc",
			"environment":  "",
			"orchestrator": "turing",
			"stream":       "test-stream",
			"team":         "test-team",
		},
		Endpoint:         "test-svc-turing-router.models.example.com",
		HostRewrite:      "test-svc-turing-router-1.models.example.com",
		Gateway:          defaultGateway,
		DestinationHost:  defaultIstioGateway,
		MatchURIPrefixes: defaultMatchURIPrefixes,
	}

	got, err := sb.NewRouterEndpoint(&routerVersion, project, "test-env", versionEndpoint)
	assert.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestBuildRouterEnvsResultLogger(t *testing.T) {
	type args struct {
		namespace       string
		environmentType string
		routerDefaults  *config.RouterDefaults
		sentryEnabled   bool
		sentryDSN       string
		secretName      string
		ver             *models.RouterVersion
	}
	namespace := "testnamespace"
	tests := []struct {
		name string
		args args
		want []corev1.EnvVar
	}{
		{
			name: "KafkaLogger",
			args: args{
				namespace:       "testnamespace",
				environmentType: "dev",
				routerDefaults: &config.RouterDefaults{
					JaegerCollectorEndpoint: "",
					FluentdConfig:           &config.FluentdConfig{Tag: ""},
					KafkaConfig: &config.KafkaConfig{
						MaxMessageBytes: 123,
						CompressionType: "gzip",
					},
				},
				sentryEnabled: false,
				sentryDSN:     "",
				secretName:    "",
				ver: &models.RouterVersion{
					Router:  &models.Router{Name: "test1"},
					Version: 1,
					Timeout: "10s",
					LogConfig: &models.LogConfig{
						LogLevel:             "DEBUG",
						CustomMetricsEnabled: false,
						FiberDebugLogEnabled: false,
						JaegerEnabled:        false,
						ResultLoggerType:     "kafka",
						KafkaConfig: &models.KafkaConfig{
							Brokers:             "1.1.1.1:1111",
							Topic:               "kafkatopic",
							SerializationFormat: "protobuf",
						},
					},
				},
			},
			want: []corev1.EnvVar{
				{Name: "APP_NAME", Value: "test1-1.testnamespace"},
				{Name: "APP_ENVIRONMENT", Value: "dev"},
				{Name: "ROUTER_TIMEOUT", Value: "10s"},
				{Name: "APP_JAEGER_COLLECTOR_ENDPOINT", Value: ""},
				{Name: "ROUTER_CONFIG_FILE", Value: "/app/config/fiber.yml"},
				{Name: "APP_SENTRY_ENABLED", Value: "false"},
				{Name: "APP_SENTRY_DSN", Value: ""},
				{Name: "APP_LOGLEVEL", Value: "DEBUG"},
				{Name: "APP_CUSTOM_METRICS", Value: "false"},
				{Name: "APP_JAEGER_ENABLED", Value: "false"},
				{Name: "APP_RESULT_LOGGER", Value: "kafka"},
				{Name: "APP_FIBER_DEBUG_LOG", Value: "false"},
				{Name: "APP_KAFKA_BROKERS", Value: "1.1.1.1:1111"},
				{Name: "APP_KAFKA_TOPIC", Value: "kafkatopic"},
				{Name: "APP_KAFKA_SERIALIZATION_FORMAT", Value: "protobuf"},
				{Name: "APP_KAFKA_MAX_MESSAGE_BYTES", Value: "123"},
				{Name: "APP_KAFKA_COMPRESSION_TYPE", Value: "gzip"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := &clusterSvcBuilder{
				MaxCPU:    resource.MustParse("2"),
				MaxMemory: resource.MustParse("2Gi"),
			}
			got, _ := sb.buildRouterEnvs(
				namespace,
				tt.args.environmentType,
				tt.args.routerDefaults,
				tt.args.sentryEnabled,
				tt.args.sentryDSN,
				tt.args.secretName,
				tt.args.ver)
			assert.Equal(t, tt.want, got)
		})
	}
}
