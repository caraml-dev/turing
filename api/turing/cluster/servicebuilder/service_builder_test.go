package servicebuilder

import (
	"encoding/json"
	"path/filepath"
	"testing"

	mlp "github.com/caraml-dev/mlp/api/client"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caraml-dev/turing/api/turing/cluster"
	tu "github.com/caraml-dev/turing/api/turing/internal/testutils"
	"github.com/caraml-dev/turing/api/turing/models"
)

var testTopologySpreadConstraints = []corev1.TopologySpreadConstraint{
	{
		MaxSkew:           2,
		TopologyKey:       "kubernetes.io/hostname",
		WhenUnsatisfiable: corev1.DoNotSchedule,
		LabelSelector: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "app-expression",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"1"},
				},
			},
		},
	},
}

type testSuiteNewService struct {
	filePath     string
	initialScale *int
	expected     *cluster.KnativeService
	expRawConfig json.RawMessage
	err          string
}

func TestNewEnricherService(t *testing.T) {
	userContainerCPULimitRequestFactor := 2.0
	userContainerMemoryLimitRequestFactor := 1.5

	sb := NewClusterServiceBuilder(
		resource.MustParse("2"),
		resource.MustParse("2Gi"),
		30,
		testTopologySpreadConstraints,
		&config.KnativeServiceDefaults{
			QueueProxyResourcePercentage:          10,
			UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
			UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
		},
	)

	cpuRequest := resource.MustParse("400m")
	cpuLimit := cluster.ComputeResource(cpuRequest, userContainerCPULimitRequestFactor)

	memoryRequest := resource.MustParse("256Mi")
	memoryLimit := cluster.ComputeResource(memoryRequest, userContainerMemoryLimitRequestFactor)

	testDataBasePath := filepath.Join("..", "..", "testdata", "cluster", "servicebuilder")
	testInitialScale := 5

	tests := map[string]testSuiteNewService{
		"success": {
			filePath:     filepath.Join(testDataBasePath, "router_version_success.json"),
			initialScale: &testInitialScale,
			expected: &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					Name:           "test-svc-turing-enricher-1",
					Namespace:      "test-project",
					Image:          "asia.gcr.io/gcp-project-id/echo:1.0.2",
					CPURequests:    cpuRequest,
					CPULimit:       &cpuLimit,
					MemoryRequests: memoryRequest,
					MemoryLimit:    &memoryLimit,
					Envs: []corev1.EnvVar{
						{Name: "TEST_ENV", Value: "enricher"},
						{Name: "GOOGLE_APPLICATION_CREDENTIALS", Value: "/var/secret/enricher-service-account.json"},
					},
					Labels: map[string]string{
						"app":              "test-svc",
						"environment":      "",
						"orchestrator":     "turing",
						"stream":           "test-stream",
						"team":             "test-team",
						"custom-label-key": "value-1",
					},
					Volumes: []corev1.Volume{
						{
							Name: secretVolume,
							VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
								SecretName: "secret",
								Items:      []corev1.KeyToPath{{Key: SecretKeyNameEnricher, Path: SecretKeyNameEnricher}},
							}},
						},
					},
					VolumeMounts: []corev1.VolumeMount{{Name: secretVolume, MountPath: secretMountPath}},
				},
				IsClusterLocal:               true,
				ContainerPort:                8080,
				MinReplicas:                  1,
				MaxReplicas:                  2,
				InitialScale:                 &testInitialScale,
				AutoscalingMetric:            "concurrency",
				AutoscalingTarget:            "1",
				TopologySpreadConstraints:    testTopologySpreadConstraints,
				QueueProxyResourcePercentage: 10,
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
				Labels: []mlp.Label{{Key: "custom-label-key", Value: "value-1"}},
			}
			svc, err := sb.NewEnricherService(routerVersion, project, "secret", data.initialScale)
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
	userContainerCPULimitRequestFactor := 2.0
	userContainerMemoryLimitRequestFactor := 1.5

	sb := NewClusterServiceBuilder(
		resource.MustParse("2"),
		resource.MustParse("2Gi"),
		30,
		testTopologySpreadConstraints,
		&config.KnativeServiceDefaults{
			QueueProxyResourcePercentage:          20,
			UserContainerCPULimitRequestFactor:    userContainerCPULimitRequestFactor,
			UserContainerMemoryLimitRequestFactor: userContainerMemoryLimitRequestFactor,
		},
	)

	cpuRequest := resource.MustParse("200m")
	cpuLimit := cluster.ComputeResource(cpuRequest, userContainerCPULimitRequestFactor)

	memoryRequest := resource.MustParse("1024Mi")
	memoryLimit := cluster.ComputeResource(memoryRequest, userContainerMemoryLimitRequestFactor)

	testDataBasePath := filepath.Join("..", "..", "testdata", "cluster", "servicebuilder")
	testInitialScale := 5

	tests := map[string]testSuiteNewService{
		"success": {
			filePath:     filepath.Join(testDataBasePath, "router_version_success.json"),
			initialScale: &testInitialScale,
			expected: &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					Name:           "test-svc-turing-ensembler-1",
					Namespace:      "test-project",
					Image:          "asia.gcr.io/gcp-project-id/echo:1.0.2",
					CPURequests:    cpuRequest,
					CPULimit:       &cpuLimit,
					MemoryRequests: memoryRequest,
					MemoryLimit:    &memoryLimit,
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
								Items:      []corev1.KeyToPath{{Key: SecretKeyNameEnsembler, Path: SecretKeyNameEnsembler}},
							}},
						},
					},
					VolumeMounts: []corev1.VolumeMount{{Name: secretVolume, MountPath: secretMountPath}},
				},
				IsClusterLocal:               true,
				ContainerPort:                8080,
				MinReplicas:                  2,
				MaxReplicas:                  3,
				AutoscalingMetric:            "concurrency",
				AutoscalingTarget:            "1",
				InitialScale:                 &testInitialScale,
				TopologySpreadConstraints:    testTopologySpreadConstraints,
				QueueProxyResourcePercentage: 20,
			},
		},
		"success with ensembler docker type": {
			filePath: filepath.Join(testDataBasePath, "router_version_success_docker_ensembler.json"),
			expected: &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					Name:           "test-svc-turing-ensembler-1",
					Namespace:      "test-project",
					Image:          "asia.gcr.io/gcp-project-id/echo:1.0.2",
					CPURequests:    cpuRequest,
					CPULimit:       &cpuLimit,
					MemoryRequests: memoryRequest,
					MemoryLimit:    &memoryLimit,
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
								Items:      []corev1.KeyToPath{{Key: SecretKeyNameEnsembler, Path: SecretKeyNameEnsembler}},
							}},
						},
					},
					VolumeMounts: []corev1.VolumeMount{{Name: secretVolume, MountPath: secretMountPath}},
				},
				IsClusterLocal:               true,
				ContainerPort:                8080,
				MinReplicas:                  2,
				MaxReplicas:                  3,
				AutoscalingMetric:            "cpu",
				AutoscalingTarget:            "90",
				TopologySpreadConstraints:    testTopologySpreadConstraints,
				QueueProxyResourcePercentage: 20,
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
			svc, err := sb.NewEnsemblerService(routerVersion, project, "secret", data.initialScale)
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
	secretMap := map[string]string{
		SecretKeyNameRouter:    "router-key",
		SecretKeyNameEnricher:  "enricher-key",
		SecretKeyNameEnsembler: "ensembler-key",
		SecretKeyNameExpEngine: "exp-engine-key",
	}

	tests := map[string]struct {
		version   *models.RouterVersion
		project   *mlp.Project
		envType   string
		secretMap map[string]string
		expected  *cluster.Secret
	}{
		"success": {
			version: &models.RouterVersion{
				Version: 2,
				Router:  &models.Router{Name: "test-router"},
				ExperimentEngine: &models.ExperimentEngine{
					Type: "exp-engine",
				},
			},
			project:   &mlp.Project{Name: "test-project"},
			envType:   "test",
			secretMap: secretMap,
			expected: &cluster.Secret{
				Name:      "test-router-turing-secret-2",
				Namespace: "test-project",
				Data:      secretMap,
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
			secret := sb.NewSecret(tt.version, tt.project, tt.secretMap)
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
