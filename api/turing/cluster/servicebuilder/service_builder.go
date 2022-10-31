package servicebuilder

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"

	mlp "github.com/gojek/mlp/api/client"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/cluster/labeller"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/models"
)

const (
	secretVolume    = "svc-acct-secret-volume"
	secretMountPath = "/var/secret/"
	// Kubernetes secret key name for usage in: router, ensembler, enricher.
	// They will share the same Kubernetes secret for every RouterVersion deployment.
	// Hence, the key name should be used to retrieve different credentials.
	secretKeyNameRouter    = "router-service-account.json"
	secretKeyNameEnsembler = "ensembler-service-account.json"
	secretKeyNameEnricher  = "enricher-service-account.json"
)

var ComponentTypes = struct {
	Enricher             string
	Ensembler            string
	Router               string
	FluentdLogger        string
	Secret               string
	ServiceAccountSecret string
	CacheVolume          string
	FiberConfig          string
	PluginsServer        string
}{
	Enricher:      "enricher",
	Ensembler:     "ensembler",
	Router:        "router",
	FluentdLogger: "fluentd-logger",
	Secret:        "secret",
	CacheVolume:   "cache-volume",
	FiberConfig:   "fiber-config",
	PluginsServer: "plugins-server",
}

// ClusterServiceBuilder parses the Router Config to build a service definition
// compatible with the cluster package
type ClusterServiceBuilder interface {
	NewEnricherService(
		ver *models.RouterVersion,
		project *mlp.Project,
		envType string,
		secretName string,
		knativeQueueProxyResourcePercentage int,
		userContainerLimitRequestFactor float64,
	) (*cluster.KnativeService, error)
	NewEnsemblerService(
		ver *models.RouterVersion,
		project *mlp.Project,
		envType string,
		secretName string,
		knativeQueueProxyResourcePercentage int,
		userContainerLimitRequestFactor float64,
	) (*cluster.KnativeService, error)
	NewRouterService(
		ver *models.RouterVersion,
		project *mlp.Project,
		envType string,
		secretName string,
		experimentConfig json.RawMessage,
		routerDefaults *config.RouterDefaults,
		sentryEnabled bool,
		sentryDSN string,
		knativeQueueProxyResourcePercentage int,
		userContainerLimitRequestFactor float64,
	) (*cluster.KnativeService, error)
	NewFluentdService(
		routerVersion *models.RouterVersion,
		project *mlp.Project,
		secretName string,
		fluentdConfig *config.FluentdConfig,
	) *cluster.KubernetesService
	NewRouterEndpoint(
		routerVersion *models.RouterVersion,
		project *mlp.Project,
		envType string,
		versionEndpoint string,
	) (*cluster.VirtualService, error)
	NewSecret(
		routerVersion *models.RouterVersion,
		project *mlp.Project,
		routerServiceAccountKey string,
		enricherServiceAccountKey string,
		ensemblerServiceAccountKey string,
	) *cluster.Secret
	GetRouterServiceName(ver *models.RouterVersion) string
}

// clusterSvcBuilder implements ClusterServiceBuilder
type clusterSvcBuilder struct {
	MaxCPU    resource.Quantity
	MaxMemory resource.Quantity
}

// NewClusterServiceBuilder creates a new service builder with the supplied configs for defaults
func NewClusterServiceBuilder(
	cpuLimit resource.Quantity,
	memoryLimit resource.Quantity,
) ClusterServiceBuilder {
	return &clusterSvcBuilder{
		MaxCPU:    cpuLimit,
		MaxMemory: memoryLimit,
	}
}

// NewEnricherService creates a new cluster Service object with the required config
// for the enricher component to be deployed.
func (sb *clusterSvcBuilder) NewEnricherService(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	envType string,
	secretName string,
	knativeQueueProxyResourcePercentage int,
	userContainerLimitRequestFactor float64,
) (*cluster.KnativeService, error) {
	// Get the enricher reference
	enricher := routerVersion.Enricher
	if enricher == nil {
		return nil, errors.New("Enricher reference is empty")
	}

	// Create service name
	name := GetComponentName(routerVersion, ComponentTypes.Enricher)
	// Namespace is the name of the project
	namespace := GetNamespace(project)

	volumes := make([]corev1.Volume, 0)
	volumeMounts := make([]corev1.VolumeMount, 0)

	// If service account for enricher is not empty, we need to mount the service account key and set the
	// the environment variable GOOGLE_APPLICATION_CREDENTIALS
	if enricher.ServiceAccount != "" {
		v := corev1.Volume{
			Name: secretVolume,
			VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
				Items:      []corev1.KeyToPath{{Key: secretKeyNameEnricher, Path: secretKeyNameEnricher}},
			}},
		}
		volumes = append(volumes, v)

		vm := corev1.VolumeMount{Name: secretVolume, MountPath: secretMountPath}
		volumeMounts = append(volumeMounts, vm)

		// If there is existing environment envGoogleApplicationCredentials, replace the value, else add a new one
		existingReplaced := false
		for _, env := range enricher.Env {
			if env.Name == envGoogleApplicationCredentials {
				env.Value = filepath.Join(secretMountPath, secretKeyNameEnricher)
				existingReplaced = true
			}
		}
		if !existingReplaced {
			env := &models.EnvVar{
				Name:  envGoogleApplicationCredentials,
				Value: filepath.Join(secretMountPath, secretKeyNameEnricher),
			}
			enricher.Env = append(enricher.Env, env)
		}
	}

	return sb.validateKnativeService(&cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:           name,
			Namespace:      namespace,
			Image:          enricher.Image,
			CPURequests:    enricher.ResourceRequest.CPURequest,
			MemoryRequests: enricher.ResourceRequest.MemoryRequest,
			Envs:           enricher.Env.ToKubernetesEnvVars(),
			Labels:         buildLabels(project, routerVersion.Router),
			Volumes:        volumes,
			VolumeMounts:   volumeMounts,
		},
		IsClusterLocal:                  true,
		ContainerPort:                   int32(enricher.Port),
		MinReplicas:                     enricher.ResourceRequest.MinReplica,
		MaxReplicas:                     enricher.ResourceRequest.MaxReplica,
		AutoscalingMetric:               string(enricher.AutoscalingPolicy.Metric),
		AutoscalingTarget:               enricher.AutoscalingPolicy.Target,
		QueueProxyResourcePercentage:    knativeQueueProxyResourcePercentage,
		UserContainerLimitRequestFactor: userContainerLimitRequestFactor,
	})
}

// NewEnsemblerService creates a new cluster Service object with the required config
// for the ensembler component to be deployed.
func (sb *clusterSvcBuilder) NewEnsemblerService(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	envType string,
	secretName string,
	knativeQueueProxyResourcePercentage int,
	userContainerLimitRequestFactor float64,
) (*cluster.KnativeService, error) {
	// Get the ensembler reference
	ensembler := routerVersion.Ensembler
	if ensembler == nil {
		return nil, errors.New("Ensembler reference is empty")
	}
	docker := ensembler.DockerConfig

	// Create service name
	name := GetComponentName(routerVersion, ComponentTypes.Ensembler)
	// Namespace is the name of the project
	namespace := GetNamespace(project)

	volumes := make([]corev1.Volume, 0)
	volumeMounts := make([]corev1.VolumeMount, 0)

	// If service account for ensembler is not empty, we need to mount the service account key and set the
	// the environment variable GOOGLE_APPLICATION_CREDENTIALS
	if docker.ServiceAccount != "" {
		v := corev1.Volume{
			Name: secretVolume,
			VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
				Items:      []corev1.KeyToPath{{Key: secretKeyNameEnsembler, Path: secretKeyNameEnsembler}},
			}},
		}
		volumes = append(volumes, v)

		vm := corev1.VolumeMount{Name: secretVolume, MountPath: secretMountPath}
		volumeMounts = append(volumeMounts, vm)

		// If there is existing environment envGoogleApplicationCredentials, replace the value, else add a new one
		existingReplaced := false
		for _, env := range docker.Env {
			if env.Name == envGoogleApplicationCredentials {
				env.Value = filepath.Join(secretMountPath, secretKeyNameEnsembler)
				existingReplaced = true
			}
		}
		if !existingReplaced {
			env := &models.EnvVar{
				Name:  envGoogleApplicationCredentials,
				Value: filepath.Join(secretMountPath, secretKeyNameEnsembler),
			}
			docker.Env = append(docker.Env, env)
		}
	}

	return sb.validateKnativeService(&cluster.KnativeService{
		BaseService: &cluster.BaseService{
			Name:           name,
			Namespace:      namespace,
			Image:          docker.Image,
			CPURequests:    docker.ResourceRequest.CPURequest,
			MemoryRequests: docker.ResourceRequest.MemoryRequest,
			Envs:           docker.Env.ToKubernetesEnvVars(),
			Labels:         buildLabels(project, routerVersion.Router),
			Volumes:        volumes,
			VolumeMounts:   volumeMounts,
		},
		IsClusterLocal:                  true,
		ContainerPort:                   int32(docker.Port),
		MinReplicas:                     docker.ResourceRequest.MinReplica,
		MaxReplicas:                     docker.ResourceRequest.MaxReplica,
		AutoscalingMetric:               string(docker.AutoscalingPolicy.Metric),
		AutoscalingTarget:               docker.AutoscalingPolicy.Target,
		QueueProxyResourcePercentage:    knativeQueueProxyResourcePercentage,
		UserContainerLimitRequestFactor: userContainerLimitRequestFactor,
	})
}

// NewSecret creates a new `cluster.Secret` secret from the given service accounts.
// If [router/enricher/ensembler]ServiceAccountKey is empty no secret key will be created
// for that component. This happens when user does not specify service accounts.
func (sb *clusterSvcBuilder) NewSecret(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	routerServiceAccountKey string,
	enricherServiceAccountKey string,
	ensemblerServiceAccountKey string,
) *cluster.Secret {
	data := map[string]string{
		secretKeyNameRouter:    routerServiceAccountKey,
		secretKeyNameEnricher:  enricherServiceAccountKey,
		secretKeyNameEnsembler: ensemblerServiceAccountKey,
	}
	return &cluster.Secret{
		Name: fmt.Sprintf(
			"%s-turing-%s-%d",
			routerVersion.Router.Name,
			ComponentTypes.Secret,
			routerVersion.Version,
		),
		Namespace: project.Name,
		Data:      data,
		Labels:    buildLabels(project, routerVersion.Router),
	}
}

func (sb *clusterSvcBuilder) validateKnativeService(
	svc *cluster.KnativeService,
) (*cluster.KnativeService, error) {
	if svc.CPURequests.Cmp(sb.MaxCPU) > 0 {
		return nil, errors.New("Requested CPU is more than max permissible")
	}
	if svc.MemoryRequests.Cmp(sb.MaxMemory) > 0 {
		return nil, errors.New("Requested Memory is more than max permissible")
	}
	return svc, nil
}

func GetComponentName(routerVersion *models.RouterVersion, componentType string) string {
	return fmt.Sprintf("%s-turing-%s-%d", routerVersion.Router.Name, componentType, routerVersion.Version)
}

func GetNamespace(project *mlp.Project) string {
	return project.Name
}

func buildLabels(
	project *mlp.Project,
	router *models.Router,
) map[string]string {
	r := labeller.KubernetesLabelsRequest{
		Stream: project.Stream,
		Team:   project.Team,
		App:    router.Name,
	}
	return labeller.BuildLabels(r)
}
