// +build unit

package servicebuilder

import (
	"encoding/json"
	"path/filepath"
	"testing"

	mlp "github.com/gojek/mlp/api/client"
	"github.com/gojek/turing/api/turing/cluster"
	tu "github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/gojek/turing/api/turing/models"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type testSuiteNewService struct {
	filePath     string
	expected     cluster.KnativeService
	expRawConfig json.RawMessage
	success      bool
	err          string
}

func TestNewEnricherService(t *testing.T) {
	sb := NewClusterServiceBuilder(resource.MustParse("2"), resource.MustParse("2Gi"))
	testDataBasePath := filepath.Join("..", "..", "testdata", "cluster", "servicebuilder")
	tests := map[string]testSuiteNewService{
		"success": {
			filePath: filepath.Join(testDataBasePath, "router_version_success.json"),
			expected: cluster.KnativeService{
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
				IsClusterLocal: true,
				ContainerPort:  8080,
				MinReplicas:    1,
				MaxReplicas:    2,
			},
			success: true,
		},
		"failure": {
			filePath: filepath.Join(testDataBasePath, "router_version_basic.json"),
			success:  false,
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
			svc, err := sb.NewEnricherService(routerVersion, project, "test-env", "secret")
			if data.success {
				assert.NoError(t, err)
				tu.FailOnError(t, tu.CompareObjects(*svc, data.expected))
			} else {
				assert.Error(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}

func TestNewEnsemblerService(t *testing.T) {
	sb := NewClusterServiceBuilder(resource.MustParse("2"), resource.MustParse("2Gi"))
	testDataBasePath := filepath.Join("..", "..", "testdata", "cluster", "servicebuilder")
	tests := map[string]testSuiteNewService{
		"success": {
			filePath: filepath.Join(testDataBasePath, "router_version_success.json"),
			expected: cluster.KnativeService{
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
				IsClusterLocal: true,
				ContainerPort:  8080,
				MinReplicas:    2,
				MaxReplicas:    3,
			},
			success: true,
		},
		"success with ensembler docker type": {
			filePath: filepath.Join(testDataBasePath, "router_version_success_docker_ensembler.json"),
			expected: cluster.KnativeService{
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
				IsClusterLocal: true,
				ContainerPort:  8080,
				MinReplicas:    2,
				MaxReplicas:    3,
			},
			success: true,
		},
		"failure": {
			filePath: filepath.Join(testDataBasePath, "router_version_basic.json"),
			success:  false,
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
			svc, err := sb.NewEnsemblerService(routerVersion, project, "test-env", "secret")
			if data.success {
				assert.NoError(t, err)
				tu.FailOnError(t, tu.CompareObjects(*svc, data.expected))
			} else {
				assert.Error(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}

func TestNewSecret(t *testing.T) {
	sb := &clusterSvcBuilder{}
	expected := cluster.Secret{
		Name:      "test-router-turing-secret-2",
		Namespace: "test-project",
		Data: map[string]string{
			"router-service-account.json":    "router-key",
			"enricher-service-account.json":  "enricher-key",
			"ensembler-service-account.json": "ensembler-key",
			"experiment_passkey":             "experiment-passkey",
		},
	}

	// Run Tests
	secret := sb.NewSecret(
		&models.RouterVersion{
			Version: 2,
			Router:  &models.Router{Name: "test-router"},
			ExperimentEngine: &models.ExperimentEngine{
				Type: models.ExperimentEngineTypeLitmus,
			},
		},
		&mlp.Project{Name: "test-project"},
		"router-key", "enricher-key", "ensembler-key", "experiment-passkey",
	)
	assert.Equal(t, expected, *secret)
}

func TestValidateKnativeService(t *testing.T) {
	cpuLimit := resource.MustParse("400m")
	memoryLimit := resource.MustParse("512Mi")
	tests := map[string]struct {
		cpu resource.Quantity
		mem resource.Quantity
		err string
	}{
		"success": {
			cpu: resource.MustParse("400m"),
			mem: resource.MustParse("50Mi"),
		},
		"cpu failure": {
			cpu: resource.MustParse("4"),
			mem: resource.MustParse("100Mi"),
			err: "Requested CPU is more than max permissible",
		},
		"mem failure": {
			cpu: resource.MustParse("100m"),
			mem: resource.MustParse("1Gi"),
			err: "Requested Memory is more than max permissible",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			sb := clusterSvcBuilder{
				MaxCPU:    cpuLimit,
				MaxMemory: memoryLimit,
			}
			testSvc := &cluster.KnativeService{
				BaseService: &cluster.BaseService{
					CPURequests:    data.cpu,
					MemoryRequests: data.mem,
				},
			}
			// Run test method and validate
			svc, err := sb.validateKnativeService(testSvc)
			if data.err == "" {
				// Validate success
				assert.Equal(t, testSvc, svc)
				assert.NoError(t, err)
			} else {
				// Validate failure
				assert.Error(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}
