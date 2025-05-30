package servicebuilder

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	mlp "github.com/caraml-dev/mlp/api/client"

	corev1 "k8s.io/api/core/v1"

	"github.com/caraml-dev/turing/api/turing/config"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/caraml-dev/turing/api/turing/cluster"
	tu "github.com/caraml-dev/turing/api/turing/internal/testutils"
)

func TestNewFluentdService(t *testing.T) {
	sb := clusterSvcBuilder{
		MaxCPU:    resource.MustParse("400m"),
		MaxMemory: resource.MustParse("512Mi"),
	}

	testDataBasePath := filepath.Join("..", "..", "testdata", "cluster", "servicebuilder")
	routerVersion := tu.GetRouterVersion(t,
		filepath.Join(testDataBasePath, "router_version_success.json"))
	fluentdConfig := config.FluentdConfig{
		Image:                "fluentdimage:1.0.0",
		Tag:                  "fluentd-tag",
		FlushIntervalSeconds: 30,
		WorkerCount:          1,
	}

	project := &mlp.Project{
		Name:   "test-project",
		Stream: "test-stream",
		Team:   "test-team",
	}

	cpuRequest := resource.MustParse(fluentdCPURequest)
	cpuLimits := cluster.ComputeResource(cpuRequest, defaultCPULimitRequestFactor)
	memoryRequest := resource.MustParse(fluentdMemRequest)
	memoryLimits := cluster.ComputeResource(memoryRequest, defaultMemoryLimitRequestFactor)

	id := int64(999)
	volSize, _ := resource.ParseQuantity(cacheVolumeSize)
	expected := &cluster.KubernetesService{
		BaseService: &cluster.BaseService{
			Name:                  "test-svc-turing-fluentd-logger-1",
			Namespace:             project.Name,
			Image:                 "fluentdimage:1.0.0",
			CPURequests:           cpuRequest,
			CPULimit:              &cpuLimits,
			MemoryRequests:        memoryRequest,
			MemoryLimit:           &memoryLimits,
			LivenessHTTPGetPath:   "/fluentd.pod.healthcheck?json=%7B%22log%22%3A+%22health+check%22%7D",
			ReadinessHTTPGetPath:  "/fluentd.pod.healthcheck?json=%7B%22log%22%3A+%22health+check%22%7D",
			ProbeInitDelaySeconds: 10,
			ProbePort:             9880,
			Labels: map[string]string{
				"app": "test-svc",
				// environment is empty string because its value will only be injected
				// when Singleton is initialized from turing/api/turing/cluster/labeller/labeller.go
				"environment":  "",
				"orchestrator": "turing",
				"stream":       "test-stream",
				"team":         "test-team",
			},
			Envs: []corev1.EnvVar{
				{Name: "FLUENTD_WORKER_COUNT", Value: "1"},
				{Name: "FLUENTD_LOG_LEVEL", Value: "info"},
				{Name: "FLUENTD_LOG_PATH", Value: "/cache/log/bq_load_logs.*.buffer"},
				{Name: "FLUENTD_GCP_JSON_KEY_PATH", Value: "/var/secret/router-service-account.json"},
				{Name: "FLUENTD_BUFFER_LIMIT", Value: "10g"},
				{Name: "FLUENTD_FLUSH_INTERVAL_SECONDS", Value: "60"},
				{Name: "FLUENTD_TAG", Value: "fluentd-tag"},
				{Name: "FLUENTD_GCP_PROJECT", Value: "gcp-project-id"},
				{Name: "FLUENTD_BQ_DATASET", Value: "dataset_id"},
				{Name: "FLUENTD_BQ_TABLE", Value: "turing_log_test"},
			},
			PersistentVolumeClaim: &cluster.PersistentVolumeClaim{
				Name:        GetComponentName(routerVersion, ComponentTypes.CacheVolume),
				Namespace:   project.Name,
				AccessModes: []string{"ReadWriteOnce"},
				Size:        volSize,
				Labels: map[string]string{
					"app":          "test-svc",
					"environment":  "",
					"orchestrator": "turing",
					"stream":       "test-stream",
					"team":         "test-team",
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: secretVolume,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName: "service-account",
							Items: []corev1.KeyToPath{
								{
									Key:  SecretKeyNameRouter,
									Path: SecretKeyNameRouter,
								},
							},
						},
					},
				},
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      secretVolume,
					MountPath: secretMountPath,
				},
				{
					Name:      "test-svc-turing-cache-volume-1",
					MountPath: cacheVolumeMountPath,
				},
			},
		},
		Replicas: FluentdReplicaCount,
		Ports: []cluster.Port{
			{
				Name:     "tcp-input",
				Port:     24224,
				Protocol: "TCP",
			},
			{
				Name:     "http-input",
				Port:     9880,
				Protocol: "TCP",
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser:  &id,
			RunAsGroup: &id,
			FSGroup:    &id,
		},
	}

	actual := sb.NewFluentdService(routerVersion, project, "service-account", &fluentdConfig)
	assert.Equal(t, expected, actual)
}
