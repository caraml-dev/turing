// +build unit

package servicebuilder

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"

	corev1 "k8s.io/api/core/v1"

	mlp "github.com/gojek/mlp/api/client"
	"github.com/gojek/turing/api/turing/cluster"
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
		"endpoint": "litmus.example.com:8012",
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
	tu.FailOnError(t, err)
	cfgmapEnsembling, err := tu.ReadFile(filepath.Join(testDataBasePath, "router_configmap_ensembling.yml"))
	tu.FailOnError(t, err)
	cfgmapStandardEnsemble, err := tu.ReadFile(filepath.Join(testDataBasePath, "router_configmap_standard_ensembler.yml"))
	tu.FailOnError(t, err)
	cfgmapTrafficSplitting, err := tu.ReadFile(filepath.Join(testDataBasePath, "router_configmap_traffic_splitting.yml"))
	tu.FailOnError(t, err)

	// Define tests
	tests := map[string]testSuiteNewService{
		"success basic": {
			filePath:     filepath.Join(testDataBasePath, "router_version_basic.json"),
			expRawConfig: json.RawMessage(expRunnerConfig),
			expected: cluster.KnativeService{
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
					},
					Envs: []corev1.EnvVar{
						{Name: "APP_NAME", Value: "test-svc-1.test-project"},
						{Name: "APP_ENVIRONMENT", Value: "test-env"},
						{Name: "ROUTER_TIMEOUT", Value: "5s"},
						{Name: "APP_JAEGER_COLLECTOR_ENDPOINT", Value: "jaeger-endpoint"},
						{Name: "ROUTER_CONFIG_FILE", Value: "/app/config/fiber.yml"},
						{Name: "APP_SENTRY_ENABLED", Value: "true"},
						{Name: "APP_SENTRY_DSN", Value: "sentry-dsn"},
						{
							Name: "LITMUS_PASSKEY",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "service-account",
									},
									Key: "experiment_passkey",
								},
							},
						},
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
				ContainerPort:                8080,
				MinReplicas:                  2,
				MaxReplicas:                  4,
				TargetConcurrency:            1,
				QueueProxyResourcePercentage: 20,
			},
			success: true,
		},
		"success all components": {
			filePath:     filepath.Join(testDataBasePath, "router_version_success.json"),
			expRawConfig: json.RawMessage(`{}`),
			expected: cluster.KnativeService{
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
				ContainerPort:                8080,
				MinReplicas:                  2,
				MaxReplicas:                  4,
				TargetConcurrency:            1,
				QueueProxyResourcePercentage: 20,
			},
			success: true,
		},
		"success with standard ensembler": {
			filePath:     filepath.Join(testDataBasePath, "router_version_success_standard_ensembler.json"),
			expRawConfig: json.RawMessage(expRunnerConfig),
			expected: cluster.KnativeService{
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
					},
					Envs: []corev1.EnvVar{
						{Name: "APP_NAME", Value: "test-svc-1.test-project"},
						{Name: "APP_ENVIRONMENT", Value: "test-env"},
						{Name: "ROUTER_TIMEOUT", Value: "5s"},
						{Name: "APP_JAEGER_COLLECTOR_ENDPOINT", Value: "jaeger-endpoint"},
						{Name: "ROUTER_CONFIG_FILE", Value: "/app/config/fiber.yml"},
						{Name: "APP_SENTRY_ENABLED", Value: "true"},
						{Name: "APP_SENTRY_DSN", Value: "sentry-dsn"},
						{
							Name: "LITMUS_PASSKEY",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "service-account",
									},
									Key: "experiment_passkey",
								},
							},
						},
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
				ContainerPort:                8080,
				MinReplicas:                  2,
				MaxReplicas:                  4,
				TargetConcurrency:            1,
				QueueProxyResourcePercentage: 20,
			},
			success: true,
		},
		"success | traffic-splitting": {
			filePath:     filepath.Join(testDataBasePath, "router_version_success_traffic_splitting.json"),
			expRawConfig: json.RawMessage(expRunnerConfig),
			expected: cluster.KnativeService{
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
					},
					Envs: []corev1.EnvVar{
						{Name: "APP_NAME", Value: "test-svc-1.test-project"},
						{Name: "APP_ENVIRONMENT", Value: "test-env"},
						{Name: "ROUTER_TIMEOUT", Value: "5s"},
						{Name: "APP_JAEGER_COLLECTOR_ENDPOINT", Value: "jaeger-endpoint"},
						{Name: "ROUTER_CONFIG_FILE", Value: "/app/config/fiber.yml"},
						{Name: "APP_SENTRY_ENABLED", Value: "true"},
						{Name: "APP_SENTRY_DSN", Value: "sentry-dsn"},
						{
							Name: "LITMUS_PASSKEY",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "service-account",
									},
									Key: "experiment_passkey",
								},
							},
						},
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
				ContainerPort:                8080,
				MinReplicas:                  2,
				MaxReplicas:                  4,
				TargetConcurrency:            1,
				QueueProxyResourcePercentage: 20,
			},
			success: true,
		},
		"failure missing bigquery": {
			filePath: filepath.Join(testDataBasePath, "router_version_missing_bigquery.json"),
			success:  false,
			err:      "Missing BigQuery logger config",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Read router version test data
			fileBytes, err := tu.ReadFile(data.filePath)
			tu.FailOnError(t, err)
			// Convert to RouterVersion type
			var routerVersion models.RouterVersion
			err = json.Unmarshal(fileBytes, &routerVersion)
			tu.FailOnError(t, err)

			// Run test
			project := &mlp.Project{
				Name:   "test-project",
				Stream: "test-stream",
				Team:   "test-team",
			}
			svc, err := sb.NewRouterService(&routerVersion, project, "test-env", "service-account",
				data.expRawConfig, "fluentd-tag", "jaeger-endpoint", true, "sentry-dsn", 1, 20)

			if data.success {
				require.Nil(t, err)
				if !cmp.Equal(*svc, data.expected) {
					t.Log(cmp.Diff(*svc, data.expected))
					t.Errorf("err for input file: %s", data.filePath)
				}
			} else {
				require.NotNil(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}

func TestNewRouterEndpoint(t *testing.T) {
	// Get router version
	sb := NewClusterServiceBuilder(resource.MustParse("2"), resource.MustParse("2Gi"))
	testDataBasePath := filepath.Join("..", "..", "testdata", "cluster", "servicebuilder")
	fileBytes, err := tu.ReadFile(filepath.Join(testDataBasePath, "router_version_success.json"))
	tu.FailOnError(t, err)
	var routerVersion models.RouterVersion
	err = json.Unmarshal(fileBytes, &routerVersion)
	tu.FailOnError(t, err)

	project := &mlp.Project{
		Name:   "test-project",
		Stream: "test-stream",
		Team:   "test-team",
	}

	versionEndpoint := "http://test-svc-turing-router-1.models.example.com"

	expected := cluster.VirtualService{
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
	assert.Equal(t, expected, *got)
}
