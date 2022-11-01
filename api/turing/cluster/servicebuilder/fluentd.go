package servicebuilder

import (
	"fmt"
	"strconv"
	"strings"

	mlp "github.com/gojek/mlp/api/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/models"
)

const (
	fluentdReplicaCount  = 1
	fluentdCPURequest    = "1"
	fluentdMemRequest    = "512Mi"
	fluentdPort          = 24224
	cacheVolumeMountPath = "/cache/"
	cacheVolumeSize      = "2Gi"
)

// NewFluentdService builds a fluentd KubernetesService configuration
func (sb *clusterSvcBuilder) NewFluentdService(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	serviceAccountSecretName string,
	fluentdConfig *config.FluentdConfig,
) *cluster.KubernetesService {
	name := GetComponentName(routerVersion, ComponentTypes.FluentdLogger)
	fluentdHealthCheckPath := "/fluentd.pod.healthcheck?json=%7B%22log%22%3A+%22health+check%22%7D"

	tableSplit := strings.Split(routerVersion.LogConfig.BigQueryConfig.Table, ".")
	envs := []corev1.EnvVar{
		{Name: "FLUENTD_LOG_LEVEL", Value: "info"},
		{Name: "FLUENTD_LOG_PATH", Value: "/cache/log/bq_load_logs.*.buffer"},
		{Name: "FLUENTD_GCP_JSON_KEY_PATH", Value: secretMountPath + secretKeyNameRouter},
		{Name: "FLUENTD_BUFFER_LIMIT", Value: "10g"},
		{Name: "FLUENTD_FLUSH_INTERVAL_SECONDS", Value: strconv.Itoa(fluentdConfig.FlushIntervalSeconds)},
		{Name: "FLUENTD_TAG", Value: fluentdConfig.Tag},
		{Name: "FLUENTD_GCP_PROJECT", Value: tableSplit[0]},
		{Name: "FLUENTD_BQ_DATASET", Value: tableSplit[1]},
		{Name: "FLUENTD_BQ_TABLE", Value: tableSplit[2]},
	}

	pvcName := GetComponentName(routerVersion, ComponentTypes.CacheVolume)
	persistentVolumeClaim := &cluster.PersistentVolumeClaim{
		Name:        pvcName,
		Namespace:   project.Name,
		AccessModes: []string{"ReadWriteOnce"},
		Size:        resource.MustParse(cacheVolumeSize),
		Labels:      buildLabels(project, routerVersion.Router),
	}
	volumes, volumeMounts := buildFluentdVolumes(serviceAccountSecretName, persistentVolumeClaim.Name)

	// Overriding the security context so that fluentd is able to write logs
	// to the persistent volume.
	securityContextID := int64(999)
	return &cluster.KubernetesService{
		BaseService: &cluster.BaseService{
			Name:                  name,
			Namespace:             project.Name,
			Image:                 fluentdConfig.Image,
			CPURequests:           resource.MustParse(fluentdCPURequest),
			MemoryRequests:        resource.MustParse(fluentdMemRequest),
			ProbePort:             9880,
			LivenessHTTPGetPath:   fluentdHealthCheckPath,
			ReadinessHTTPGetPath:  fluentdHealthCheckPath,
			ProbeInitDelaySeconds: 10,
			Labels:                buildLabels(project, routerVersion.Router),
			Envs:                  envs,
			PersistentVolumeClaim: persistentVolumeClaim,
			Volumes:               volumes,
			VolumeMounts:          volumeMounts,
		},
		Replicas: fluentdReplicaCount,
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
			RunAsUser:  &securityContextID,
			RunAsGroup: &securityContextID,
			FSGroup:    &securityContextID,
		},
	}
}

func buildFluentdVolumes(
	svcAccountSecretName string,
	cacheVolumePVCName string,
) ([]corev1.Volume, []corev1.VolumeMount) {
	volumes := make([]corev1.Volume, 0)
	volumeMounts := make([]corev1.VolumeMount, 0)

	// Service account
	volumes = append(volumes, corev1.Volume{
		Name: secretVolume,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: svcAccountSecretName,
				Items: []corev1.KeyToPath{
					{
						Key:  secretKeyNameRouter,
						Path: secretKeyNameRouter,
					},
				},
			},
		},
	})
	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      secretVolume,
		MountPath: secretMountPath,
	})

	volumes = append(volumes, corev1.Volume{
		Name: ComponentTypes.CacheVolume,
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: cacheVolumePVCName,
			},
		},
	})

	volumeMounts = append(volumeMounts, corev1.VolumeMount{
		Name:      ComponentTypes.CacheVolume,
		MountPath: cacheVolumeMountPath,
	})

	return volumes, volumeMounts
}

func buildFluentdHost(
	routerVersion *models.RouterVersion,
	namespace string,
) string {
	componentName := GetComponentName(routerVersion, ComponentTypes.FluentdLogger)
	return fmt.Sprintf("%s.%s.svc.cluster.local", componentName, namespace)
}
