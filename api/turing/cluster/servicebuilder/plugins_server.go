package servicebuilder

import (
	"fmt"

	mlp "github.com/gojek/mlp/api/client"
	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/models"
	v1 "k8s.io/api/core/v1"
)

const (
	envPluginName = "PLUGIN_NAME"
	envPluginsDir = "PLUGINS_DIR"
)

const (
	nginxImage        = "nginx:1.21.5"
	pluginsVolumeName = "plugins-volume"
	pluginsMountPath  = "/usr/share/nginx/html/plugins"
)

func (sb *clusterSvcBuilder) NewPluginsServerService(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	envType string,
) *cluster.KubernetesService {
	pluginsVolume := v1.Volume{
		Name: pluginsVolumeName,
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}

	return &cluster.KubernetesService{
		BaseService: &cluster.BaseService{
			Name:                  GetComponentName(routerVersion, ComponentTypes.PluginsServer),
			Namespace:             project.Name,
			Image:                 nginxImage,
			Labels:                buildLabels(project, envType, routerVersion.Router),
			ProbePort:             80,
			LivenessHTTPGetPath:   "/",
			ReadinessHTTPGetPath:  "/",
			ProbeInitDelaySeconds: 5,

			VolumeMounts: []v1.VolumeMount{
				{
					Name:      pluginsVolumeName,
					MountPath: pluginsMountPath,
				},
			},
			Volumes: []v1.Volume{
				pluginsVolume,
			},
		},
		Replicas: 1,
		Ports: []cluster.Port{
			{
				Name:     "http",
				Port:     80,
				Protocol: "TCP",
			},
		},
		InitContainers: []cluster.Container{
			{
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
			},
		},
	}
}

func buildPluginsServerHost(
	routerVersion *models.RouterVersion,
	namespace string,
) string {
	componentName := GetComponentName(routerVersion, ComponentTypes.PluginsServer)
	return fmt.Sprintf("%s.%s.svc.cluster.local", componentName, namespace)
}
