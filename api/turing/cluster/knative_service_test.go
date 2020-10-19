// +build unit

package cluster

import (
	"strconv"
	"testing"

	"k8s.io/apimachinery/pkg/util/intstr"

	tu "github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
	knservingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

// Define a valid Service configuration for tests
var testValidKnSvc = KnativeService{
	BaseService: &BaseService{
		Name:                 "test-svc",
		Namespace:            "test-namespace",
		Image:                "asia.gcr.io/gcp-project-id/turing-router:latest",
		CPURequests:          resource.MustParse("400m"),
		MemoryRequests:       resource.MustParse("512Mi"),
		ProbePort:            8080,
		LivenessHTTPGetPath:  "/v1/internal/live",
		ReadinessHTTPGetPath: "/v1/internal/ready",
		ConfigMap: &ConfigMap{
			Name:     "test-service-fiber-config-default",
			FileName: "fiber.yml",
			Data:     "data",
		},
		Envs: []corev1.EnvVar{
			{Name: "APP_NAME", Value: "test-svc.test-namespace"},
			{Name: "ROUTER_TIMEOUT", Value: "5s"},
			{Name: "APP_JAEGER_COLLECTOR_ENDPOINT", Value: "jaeger-endpoint"},
			{Name: "APP_LOGLEVEL", Value: "INFO"},
			{Name: "APP_CUSTOM_METRICS", Value: "false"},
			{Name: "APP_JAEGER_ENABLED", Value: "false"},
			{Name: "APP_RESULT_LOGGER", Value: "bigquery"},
			{Name: "APP_FIBER_DEBUG_LOG", Value: "false"},
			{Name: "APP_GCP_PROJECT", Value: "gcp-project-id"},
			{Name: "APP_BQ_DATASET", Value: "dataset_id"},
			{Name: "APP_BQ_TABLE", Value: "turing_log_test"},
			{Name: "APP_BQ_BATCH_LOAD", Value: "false"},
		},
		Labels: map[string]string{
			"labelKey": "labelVal",
		},
		Volumes: []corev1.Volume{
			{
				Name: "config-map-volume",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "test-svc-turing-fiber-config",
						},
					},
				},
			},
			{
				Name: "service-account-volume",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: "service-account",
						Items: []corev1.KeyToPath{
							{
								Key:  "service-account.json",
								Path: "service-account.json",
							},
						},
					},
				},
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config-map-volume",
				MountPath: "/app/config/",
			},
			{
				Name:      "service-account-volume",
				MountPath: "/var/secret",
			},
		},
	},
	ContainerPort:  8080,
	MinReplicas:    1,
	MaxReplicas:    2,
	IsClusterLocal: true,
}

func TestBuildKnativeServiceConfig(t *testing.T) {
	// Build expected configuration
	var timeout int64 = 30
	annotations := map[string]string{
		"autoscaling.knative.dev/minScale": "1",
		"autoscaling.knative.dev/maxScale": "2",
		"autoscaling.knative.dev/target":   strconv.Itoa(knativeSvcDefaults.TargetConcurrency),
		"autoscaling.knative.dev/class":    "kpa.autoscaling.knative.dev",
	}
	resources := corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse("800m"),
			corev1.ResourceMemory: resource.MustParse("1024Mi"),
		},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse("400m"),
			corev1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}
	envs := []corev1.EnvVar{
		{Name: "APP_NAME", Value: "test-svc.test-namespace"},
		{Name: "ROUTER_TIMEOUT", Value: "5s"},
		{Name: "APP_JAEGER_COLLECTOR_ENDPOINT", Value: "jaeger-endpoint"},
		{Name: "APP_LOGLEVEL", Value: "INFO"},
		{Name: "APP_CUSTOM_METRICS", Value: "false"},
		{Name: "APP_JAEGER_ENABLED", Value: "false"},
		{Name: "APP_RESULT_LOGGER", Value: "bigquery"},
		{Name: "APP_FIBER_DEBUG_LOG", Value: "false"},
		{Name: "APP_GCP_PROJECT", Value: "gcp-project-id"},
		{Name: "APP_BQ_DATASET", Value: "dataset_id"},
		{Name: "APP_BQ_TABLE", Value: "turing_log_test"},
		{Name: "APP_BQ_BATCH_LOAD", Value: "false"},
	}
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Image: "asia.gcr.io/gcp-project-id/turing-router:latest",
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 8080,
					},
				},
				Resources: resources,
				LivenessProbe: &corev1.Probe{
					Handler: corev1.Handler{
						HTTPGet: &corev1.HTTPGetAction{
							Port: intstr.FromInt(8080),
							Path: "/v1/internal/live",
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
							Path: "/v1/internal/ready",
						},
					},
					InitialDelaySeconds: 20,
					PeriodSeconds:       10,
					TimeoutSeconds:      5,
					FailureThreshold:    5,
				},
				VolumeMounts: testValidKnSvc.VolumeMounts,
				Env:          envs,
			},
		},
		Volumes: testValidKnSvc.Volumes,
	}
	expected := &knservingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
			Namespace: "test-namespace",
			Labels: map[string]string{
				"labelKey":                       "labelVal",
				"serving.knative.dev/visibility": "cluster-local",
			},
		},
		Spec: knservingv1alpha1.ServiceSpec{
			ConfigurationSpec: knservingv1alpha1.ConfigurationSpec{
				Template: &knservingv1alpha1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Name: "test-svc-0",
						Labels: map[string]string{
							"labelKey": "labelVal",
						},
						Annotations: annotations,
					},
					Spec: knservingv1alpha1.RevisionSpec{
						RevisionSpec: knservingv1.RevisionSpec{
							TimeoutSeconds: &timeout,
							PodSpec:        podSpec,
						},
					},
				},
			},
		},
	}

	// Run test and validate
	svc := testValidKnSvc.BuildKnativeServiceConfig()
	err := tu.CompareObjects(svc, expected)
	assert.NoError(t, err)
}
