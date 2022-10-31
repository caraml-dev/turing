package servicebuilder

// TODO: Delete this file once all existing routers no longer use the old plugins server service to deploy plugins
import (
	"fmt"

	mlp "github.com/gojek/mlp/api/client"
	v1 "k8s.io/api/core/v1"

	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/models"
)

const (
	nginxImage                = "nginx:1.21.5"
	pluginsServerReplicaCount = 1
)

var (
	pluginsVolume = v1.Volume{
		Name: "plugins-volume",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
)

func NewPluginsServerService(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
) *cluster.KubernetesService {
	return &cluster.KubernetesService{
		BaseService: &cluster.BaseService{
			Name:                  GetComponentName(routerVersion, ComponentTypes.PluginsServer),
			Namespace:             project.Name,
			Image:                 nginxImage,
			Labels:                buildLabels(project, routerVersion.Router),
			ProbePort:             80,
			LivenessHTTPGetPath:   "/",
			ReadinessHTTPGetPath:  "/",
			ProbeInitDelaySeconds: 5,
			VolumeMounts: []v1.VolumeMount{
				{
					Name:      pluginsVolume.Name,
					MountPath: "/usr/share/nginx/html/plugins",
				},
			},
			Volumes: []v1.Volume{pluginsVolume},
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
				Name:  fmt.Sprintf("%s-plugin", routerVersion.ExperimentEngine.Type),
				Image: routerVersion.ExperimentEngine.PluginConfig.Image,
				Envs: []cluster.Env{
					{
						Name:  "PLUGIN_NAME",
						Value: routerVersion.ExperimentEngine.Type,
					},
					{
						Name:  "PLUGINS_DIR",
						Value: "/usr/share/nginx/html/plugins",
					},
				},
				VolumeMounts: []cluster.VolumeMount{
					{
						Name:      pluginsVolume.Name,
						MountPath: "/usr/share/nginx/html/plugins",
					},
				},
			},
		},
	}
}
