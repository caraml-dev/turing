package cluster

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// defaultCPULimitRequestFactor is the default multiplication factor applied to the CPU request,
	// to be set as the limit
	defaultCPULimitRequestFactor = 1.0
	// defaultMemoryLimitRequestFactor is the default multiplication factor applied to the memory request,
	// to be set as the limit
	defaultMemoryLimitRequestFactor = 1.0
)

// KubernetesService defines the properties for Kubernetes services
type KubernetesService struct {
	*BaseService

	InitContainers  []Container                `json:"init_containers"`
	Command         []string                   `json:"command"`
	Args            []string                   `json:"args"`
	Replicas        int                        `json:"replicas"`
	Ports           []Port                     `json:"ports"`
	SecurityContext *corev1.PodSecurityContext `json:"security_context"`
}

func (cfg *KubernetesService) BuildKubernetesServiceConfig() (*appsv1.Deployment, *corev1.Service) {
	deployment := cfg.buildDeployment(cfg.Labels)
	service := cfg.buildService(cfg.Labels)
	return deployment, service
}

func (cfg *KubernetesService) buildDeployment(labels map[string]string) *appsv1.Deployment {
	replicas := int32(cfg.Replicas)

	labels["app"] = cfg.Name

	initContainers := make([]corev1.Container, len(cfg.InitContainers))
	for idx, containerCfg := range cfg.InitContainers {
		initContainers[idx] = containerCfg.Build()
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": cfg.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cfg.Name,
					Namespace: cfg.Namespace,
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: initContainers,
					Containers: []corev1.Container{
						{
							Name:            cfg.Name,
							Image:           cfg.Image,
							Command:         cfg.Command,
							Args:            cfg.Command,
							Ports:           cfg.buildContainerPorts(),
							Env:             cfg.Envs,
							Resources:       cfg.buildResourceReqs(defaultCPULimitRequestFactor, defaultMemoryLimitRequestFactor),
							VolumeMounts:    cfg.VolumeMounts,
							LivenessProbe:   cfg.buildContainerProbe(livenessProbeType, int(cfg.ProbePort)),
							ReadinessProbe:  cfg.buildContainerProbe(readinessProbeType, int(cfg.ProbePort)),
							ImagePullPolicy: "IfNotPresent",
						},
					},
					Volumes:         cfg.Volumes,
					SecurityContext: cfg.SecurityContext,
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
		},
	}

}

func (cfg *KubernetesService) buildService(labels map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: cfg.buildServicePorts(),
			Selector: map[string]string{
				"app": cfg.Name,
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

func (cfg *KubernetesService) buildContainerPorts() []corev1.ContainerPort {
	containerPorts := make([]corev1.ContainerPort, len(cfg.Ports))
	for i, p := range cfg.Ports {
		containerPorts[i] = corev1.ContainerPort{
			Name:          p.Name,
			ContainerPort: int32(p.Port),
			Protocol:      corev1.Protocol(p.Protocol),
		}
	}
	return containerPorts
}

func (cfg *KubernetesService) buildServicePorts() []corev1.ServicePort {
	svcPorts := make([]corev1.ServicePort, len(cfg.Ports))
	for i, p := range cfg.Ports {
		svcPorts[i] = corev1.ServicePort{
			Name:       p.Name,
			Port:       int32(p.Port),
			TargetPort: intstr.FromInt(p.Port),
			Protocol:   corev1.Protocol(p.Protocol),
		}
	}
	return svcPorts
}
