package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	tu "github.com/caraml-dev/turing/api/turing/internal/testutils"
)

func TestBuildKubernetesServiceConfig(t *testing.T) {
	id := int64(999)
	svcConf := KubernetesService{
		BaseService: &BaseService{
			Name:                 "test-svc-fluentd-logger",
			Namespace:            "namespace",
			Image:                "fluentdimage:1.0.0",
			CPURequests:          resource.MustParse("1"),
			MemoryRequests:       resource.MustParse("1"),
			ProbePort:            8080,
			LivenessHTTPGetPath:  "/fluentd.pod.healthcheck?json=%7B%22log%22%3A+%22health+check%22%7D",
			ReadinessHTTPGetPath: "/fluentd.pod.healthcheck?json=%7B%22log%22%3A+%22health+check%22%7D",
			Labels:               map[string]string{},
			Envs: []corev1.EnvVar{
				{Name: "FLUENTD_LOG_LEVEL", Value: "info"},
			},
			PersistentVolumeClaim: &PersistentVolumeClaim{
				Name: "pvc",
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "cache-volume",
					MountPath: "/tmp/cache/",
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "cache-volume",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "pvc",
						},
					},
				},
			},
		},
		Replicas: 1,
		Ports: []Port{
			{
				Name:     "tcp-input",
				Port:     24224,
				Protocol: "TCP",
			},
		},
		SecurityContext: &corev1.PodSecurityContext{
			RunAsUser:  &id,
			RunAsGroup: &id,
			FSGroup:    &id,
		},
	}

	labels := map[string]string{}

	replicas := int32(1)
	labels["app"] = "test-svc-fluentd-logger"
	expectedDeployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc-fluentd-logger",
			Namespace: "namespace",
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "test-svc-fluentd-logger",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc-fluentd-logger",
					Namespace: "namespace",
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{},
					Containers: []corev1.Container{
						{
							Name:  "test-svc-fluentd-logger",
							Image: "fluentdimage:1.0.0",
							Ports: []corev1.ContainerPort{
								{
									Name:          "tcp-input",
									Protocol:      "TCP",
									ContainerPort: 24224,
								},
							},
							Env: []corev1.EnvVar{
								{Name: "FLUENTD_LOG_LEVEL", Value: "info"},
							},
							Resources: corev1.ResourceRequirements{
								Requests: map[corev1.ResourceName]resource.Quantity{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1"),
								},
								Limits: map[corev1.ResourceName]resource.Quantity{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1"),
								},
							},
							VolumeMounts: svcConf.VolumeMounts,
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Port: intstr.FromInt(8080),
										Path: "/fluentd.pod.healthcheck?json=%7B%22log%22%3A+%22health+check%22%7D",
									},
								},
								InitialDelaySeconds: 20,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								FailureThreshold:    5,
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Port: intstr.FromInt(8080),
										Path: "/fluentd.pod.healthcheck?json=%7B%22log%22%3A+%22health+check%22%7D",
									},
								},
								InitialDelaySeconds: 20,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								FailureThreshold:    5,
							},
							ImagePullPolicy: "IfNotPresent",
						},
					},
					Volumes: svcConf.Volumes,
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:  &id,
						RunAsGroup: &id,
						FSGroup:    &id,
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
		},
	}

	expectedService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc-fluentd-logger",
			Namespace: "namespace",
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "tcp-input",
					Protocol:   "TCP",
					Port:       24224,
					TargetPort: intstr.FromInt(24224),
				},
			},
			Selector: map[string]string{
				"app": "test-svc-fluentd-logger",
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
	gotDeployment, gotService := svcConf.BuildKubernetesServiceConfig()
	err := tu.CompareObjects(*gotDeployment, expectedDeployment)
	assert.NoError(t, err)
	err = tu.CompareObjects(*gotService, expectedService)
	assert.NoError(t, err)
}
