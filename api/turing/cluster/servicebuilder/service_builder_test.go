package servicebuilder

import (
	"encoding/json"
	"path/filepath"
	"testing"

	mlp "github.com/gojek/mlp/api/client"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/caraml-dev/turing/api/turing/cluster"
	tu "github.com/caraml-dev/turing/api/turing/internal/testutils"
	"github.com/caraml-dev/turing/api/turing/models"
)

type testSuiteNewService struct {
	filePath     string
	expected     *cluster.KnativeService
	expRawConfig json.RawMessage
	err          string
}

func TestNewEnricherService(t *testing.T) {
	sb := NewClusterServiceBuilder(resource.MustParse("2"), resource.MustParse("2Gi"), 30, 0, 10)
	testDataBasePath := filepath.Join("..", "..", "testdata", "cluster", "servicebuilder")

	tests := map[string]testSuiteNewService{
		"success": {
			filePath: filepath.Join(testDataBasePath, "router_version_success.json"),
			expected: &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					Name:           "test-svc-turing-enricher-1",
					Namespace:      "test-project",
					Image:          "asia.gcr.io/gcp-project-id/echo:1.0.2",
					CPURequests:    resource.MustParse("400m"),
					MemoryRequests: resource.MustParse("256Mi"),
					Envs: []corev1.EnvVar{
						{Name: "TEST_ENV", Value: "enricher"},
						{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "/var/secret/enricher-service-account.json"},
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
							Name: secretVolume,
							VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
								SecretName: "secret",
								Items:      []corev1.KeyToPath{{Key: secretKeyNameEnricher, Path: secretKeyNameEnricher}},
							}},
						},
					},
					VolumeMounts: []corev1.VolumeMount{{Name: secretVolume, MountPath: secretMountPath}},
				},
				IsClusterLocal:                  true,
				ContainerPort:                   8080,
				MinReplicas:                     1,
				MaxReplicas:                     2,
				AutoscalingMetric:               "concurrency",
				AutoscalingTarget:               "1",
				QueueProxyResourcePercentage:    10,
				UserContainerLimitRequestFactor: 1.5,
			},
		},
		"failure": {
			filePath: filepath.Join(testDataBasePath, "router_version_basic_upi.json"),
			err:      "Enricher reference is empty",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Read router version test data
			routerVersion := tu.GetRouterVersion(t, data.filePath)

			// Run test and validate
			project := &mlp.Project{
				Name:   "test-project",
				Stream: "test-stream",
				Team:   "test-team",
			}
			svc, err := sb.NewEnricherService(routerVersion, project, "test-env", "secret", 10, 1.5)
			if data.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, svc)
			} else {
				assert.EqualError(t, err, data.err)
			}
		})
	}
}

func TestNewEnsemblerService(t *testing.T) {
	sb := NewClusterServiceBuilder(resource.MustParse("2"), resource.MustParse("2Gi"), 30, 0, 10)
	testDataBasePath := filepath.Join("..", "..", "testdata", "cluster", "servicebuilder")
	tests := map[string]testSuiteNewService{
		"success": {
			filePath: filepath.Join(testDataBasePath, "router_version_success.json"),
			expected: &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					Name:           "test-svc-turing-ensembler-1",
					Namespace:      "test-project",
					Image:          "asia.gcr.io/gcp-project-id/echo:1.0.2",
					CPURequests:    resource.MustParse("200m"),
					MemoryRequests: resource.MustParse("1024Mi"),
					Envs: []corev1.EnvVar{
						{Name: "TEST_ENV", Value: "ensembler"},
						{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "/var/secret/ensembler-service-account.json"},
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
							Name: secretVolume,
							VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
								SecretName: "secret",
								Items:      []corev1.KeyToPath{{Key: secretKeyNameEnsembler, Path: secretKeyNameEnsembler}},
							}},
						},
					},
					VolumeMounts: []corev1.VolumeMount{{Name: secretVolume, MountPath: secretMountPath}},
				},
				IsClusterLocal:                  true,
				ContainerPort:                   8080,
				MinReplicas:                     2,
				MaxReplicas:                     3,
				AutoscalingMetric:               "concurrency",
				AutoscalingTarget:               "1",
				QueueProxyResourcePercentage:    20,
				UserContainerLimitRequestFactor: 1.5,
			},
		},
		"success with ensembler docker type": {
			filePath: filepath.Join(testDataBasePath, "router_version_success_docker_ensembler.json"),
			expected: &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					Name:           "test-svc-turing-ensembler-1",
					Namespace:      "test-project",
					Image:          "asia.gcr.io/gcp-project-id/echo:1.0.2",
					CPURequests:    resource.MustParse("200m"),
					MemoryRequests: resource.MustParse("1024Mi"),
					Envs: []corev1.EnvVar{
						{Name: "TEST_ENV", Value: "ensembler"},
						{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "/var/secret/ensembler-service-account.json"},
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
							Name: secretVolume,
							VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
								SecretName: "secret",
								Items:      []corev1.KeyToPath{{Key: secretKeyNameEnsembler, Path: secretKeyNameEnsembler}},
							}},
						},
					},
					VolumeMounts: []corev1.VolumeMount{{Name: secretVolume, MountPath: secretMountPath}},
				},
				IsClusterLocal:                  true,
				ContainerPort:                   8080,
				MinReplicas:                     2,
				MaxReplicas:                     3,
				AutoscalingMetric:               "cpu",
				AutoscalingTarget:               "90",
				QueueProxyResourcePercentage:    20,
				UserContainerLimitRequestFactor: 1.5,
			},
		},
		"failure": {
			filePath: filepath.Join(testDataBasePath, "router_version_basic_upi.json"),
			err:      "Ensembler reference is empty",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Read router version test data
			routerVersion := tu.GetRouterVersion(t, data.filePath)

			// Run test and validate
			project := &mlp.Project{
				Name:   "test-project",
				Stream: "test-stream",
				Team:   "test-team",
			}
			svc, err := sb.NewEnsemblerService(routerVersion, project, "test-env", "secret", 20, 1.5)
			if data.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, svc)
			} else {
				assert.EqualError(t, err, data.err)
			}
		})
	}
}

func TestNewSecret(t *testing.T) {
	tests := map[string]struct {
		version         *models.RouterVersion
		project         *mlp.Project
		envType         string
		routerSvcKey    string
		enricherSvcKey  string
		ensemblerSvcKey string
		expEngineSvcKey string
		expected        *cluster.Secret
	}{
		"success": {
			version: &models.RouterVersion{
				Version: 2,
				Router:  &models.Router{Name: "test-router"},
				ExperimentEngine: &models.ExperimentEngine{
					Type: "exp-engine",
				},
			},
			project:         &mlp.Project{Name: "test-project"},
			envType:         "test",
			routerSvcKey:    "router-key",
			enricherSvcKey:  "enricher-key",
			ensemblerSvcKey: "ensembler-key",
			expEngineSvcKey: "exp-engine-key",
			expected: &cluster.Secret{
				Name:      "test-router-turing-secret-2",
				Namespace: "test-project",
				Data: map[string]string{
					"router-service-account.json":     "router-key",
					"enricher-service-account.json":   "enricher-key",
					"ensembler-service-account.json":  "ensembler-key",
					"exp-engine-service-account.json": "exp-engine-key",
				},
				Labels: map[string]string{
					"app":          "test-router",
					"environment":  "",
					"orchestrator": "turing",
					"stream":       "",
					"team":         "",
				},
			},
		},
	}
	sb := &clusterSvcBuilder{}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			secret := sb.NewSecret(tt.version, tt.project, tt.routerSvcKey, tt.enricherSvcKey, tt.ensemblerSvcKey,
				tt.expEngineSvcKey)
			assert.Equal(t, tt.expected, secret)
		})
	}
}

func TestValidateKnativeService(t *testing.T) {
	cpuLimit := resource.MustParse("400m")
	memoryLimit := resource.MustParse("512Mi")
	maxAllowedReplica := 30
	tests := map[string]struct {
		cpu         resource.Quantity
		mem         resource.Quantity
		maxReplicas int
		err         string
	}{
		"success": {
			cpu:         resource.MustParse("400m"),
			mem:         resource.MustParse("50Mi"),
			maxReplicas: 10,
		},
		"cpu failure": {
			cpu:         resource.MustParse("4"),
			mem:         resource.MustParse("100Mi"),
			maxReplicas: 10,
			err:         "Requested CPU is more than max permissible",
		},
		"mem failure": {
			cpu:         resource.MustParse("100m"),
			mem:         resource.MustParse("1Gi"),
			maxReplicas: 10,
			err:         "Requested Memory is more than max permissible",
		},
		"max replica failure": {
			cpu:         resource.MustParse("100m"),
			mem:         resource.MustParse("100Mi"),
			maxReplicas: 50,
			err:         "Requested Max Replica (50) is more than max permissible (30)",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			sb := clusterSvcBuilder{
				MaxCPU:            cpuLimit,
				MaxMemory:         memoryLimit,
				MaxAllowedReplica: maxAllowedReplica,
			}
			testSvc := &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					CPURequests:    data.cpu,
					MemoryRequests: data.mem,
				},
				MaxReplicas: data.maxReplicas,
			}
			// Run test method and validate
			svc, err := sb.validateKnativeService(testSvc)
			if data.err == "" {
				assert.Equal(t, testSvc, svc)
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, data.err)
			}
		})
	}
}
