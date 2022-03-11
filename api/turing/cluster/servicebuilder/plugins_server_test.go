package servicebuilder

import (
	"testing"

	mlp "github.com/gojek/mlp/api/client"
	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/models"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestClusterSvcBuilder_NewPluginsServerService(t *testing.T) {
	tests := map[string]struct {
		version  *models.RouterVersion
		project  *mlp.Project
		envType  string
		expected *cluster.KubernetesService
		panics   bool
	}{
		"success": {
			version: &models.RouterVersion{
				Version: 42,
				Router:  &models.Router{Name: "router-1"},
				ExperimentEngine: &models.ExperimentEngine{
					Type: "exp-engine-1",
					PluginConfig: &config.ExperimentEnginePluginConfig{
						Image: "ghcr.io/myproject/exp-engine-1-plugin:latest",
					},
				},
			},
			project: &mlp.Project{Name: "integration-test"},
			envType: "test",
			expected: &cluster.KubernetesService{
				BaseService: &cluster.BaseService{
					Name:                  "router-1-turing-plugins-server-42",
					Namespace:             "integration-test",
					Image:                 nginxImage,
					ProbePort:             80,
					LivenessHTTPGetPath:   "/",
					ReadinessHTTPGetPath:  "/",
					ProbeInitDelaySeconds: 5,
					Labels: buildLabels(
						&mlp.Project{Name: "integration-test"},
						&models.Router{Name: "router-1"}),
					Volumes: []v1.Volume{
						pluginsVolume,
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      pluginsVolume.Name,
							MountPath: pluginsMountPath,
						},
					},
				},
				Replicas: pluginsServerReplicaCount,
				Ports: []cluster.Port{
					{
						Name:     "http",
						Port:     80,
						Protocol: "TCP",
					},
				},
				InitContainers: []cluster.Container{
					{
						Name:  "exp-engine-1-plugin",
						Image: "ghcr.io/myproject/exp-engine-1-plugin:latest",
						Envs: []cluster.Env{
							{
								Name:  envPluginName,
								Value: "exp-engine-1",
							},
							{
								Name:  envPluginsDir,
								Value: pluginsMountPath,
							},
						},
						VolumeMounts: []cluster.VolumeMount{
							{
								Name:      pluginsVolume.Name,
								MountPath: pluginsMountPath,
							},
						},
					},
				},
			},
		},
		"failure | no exp engine": {
			version: &models.RouterVersion{
				Version: 42,
				Router:  &models.Router{Name: "router-1"},
			},
			project: &mlp.Project{Name: "integration-test"},
			panics:  true,
		},
	}

	sb := NewClusterServiceBuilder(resource.MustParse("100m"), resource.MustParse("500Mi"))

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if tt.panics {
				assert.Panics(t, func() { sb.NewPluginsServerService(tt.version, tt.project, tt.envType) })
			} else {
				actual := sb.NewPluginsServerService(tt.version, tt.project, tt.envType)
				assert.Equal(t, tt.expected, actual)
			}

		})
	}
}
