package cluster

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"

	tu "github.com/caraml-dev/turing/api/turing/internal/testutils"
	routerConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"
)

func TestBuildKnativeServiceConfig(t *testing.T) {
	// Test configuration
	baseSvc := &BaseService{
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
	}

	// Expected specs
	var defaultConcurrency, defaultTrafficPercent int64 = 0, 100
	var defaultLatestRevision = true
	var timeout int64 = 30
	defaultRouteSpec := knservingv1.RouteSpec{
		Traffic: []knservingv1.TrafficTarget{
			{
				LatestRevision: &defaultLatestRevision,
				Percent:        &defaultTrafficPercent,
			},
		},
	}
	resources := corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:    resource.MustParse("600m"),
			corev1.ResourceMemory: resource.MustParse("768Mi"),
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
				Name:  "user-container",
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
					SuccessThreshold:    1,
					TimeoutSeconds:      5,
					FailureThreshold:    5,
				},
				VolumeMounts: baseSvc.VolumeMounts,
				Env:          envs,
			},
		},
		Volumes: baseSvc.Volumes,
	}

	tests := map[string]struct {
		serviceCfg   KnativeService
		expectedSpec knservingv1.Service
	}{
		"basic": {
			serviceCfg: KnativeService{
				BaseService:                     baseSvc,
				ContainerPort:                   8080,
				MinReplicas:                     1,
				MaxReplicas:                     2,
				AutoscalingMetric:               "concurrency",
				AutoscalingTarget:               "1",
				IsClusterLocal:                  true,
				QueueProxyResourcePercentage:    30,
				UserContainerLimitRequestFactor: 1.5,
			},
			expectedSpec: knservingv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: "test-namespace",
					Labels: map[string]string{
						"labelKey":                          "labelVal",
						"networking.knative.dev/visibility": "cluster-local",
					},
				},
				Spec: knservingv1.ServiceSpec{
					ConfigurationSpec: knservingv1.ConfigurationSpec{
						Template: knservingv1.RevisionTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-svc-0",
								Labels: map[string]string{
									"labelKey": "labelVal",
								},
								Annotations: map[string]string{
									"autoscaling.knative.dev/minScale":                     "1",
									"autoscaling.knative.dev/maxScale":                     "2",
									"autoscaling.knative.dev/metric":                       "concurrency",
									"autoscaling.knative.dev/target":                       "1",
									"autoscaling.knative.dev/class":                        "kpa.autoscaling.knative.dev",
									"queue.sidecar.serving.knative.dev/resourcePercentage": "30",
								},
							},
							Spec: knservingv1.RevisionSpec{
								PodSpec:              podSpec,
								TimeoutSeconds:       &timeout,
								ContainerConcurrency: &defaultConcurrency,
							},
						},
					},
					RouteSpec: defaultRouteSpec,
				},
			},
		},
		// upi has no liveness probe in pod spec and user-container is using h2c
		"upi": {
			serviceCfg: KnativeService{
				BaseService:                     baseSvc,
				ContainerPort:                   8080,
				MinReplicas:                     1,
				MaxReplicas:                     2,
				AutoscalingMetric:               "concurrency",
				AutoscalingTarget:               "1",
				IsClusterLocal:                  true,
				QueueProxyResourcePercentage:    30,
				UserContainerLimitRequestFactor: 1.5,
				Protocol:                        routerConfig.UPI,
			},
			expectedSpec: knservingv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: "test-namespace",
					Labels: map[string]string{
						"labelKey":                          "labelVal",
						"networking.knative.dev/visibility": "cluster-local",
					},
				},
				Spec: knservingv1.ServiceSpec{
					ConfigurationSpec: knservingv1.ConfigurationSpec{
						Template: knservingv1.RevisionTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-svc-0",
								Labels: map[string]string{
									"labelKey": "labelVal",
								},
								Annotations: map[string]string{
									"autoscaling.knative.dev/minScale":                     "1",
									"autoscaling.knative.dev/maxScale":                     "2",
									"autoscaling.knative.dev/metric":                       "concurrency",
									"autoscaling.knative.dev/target":                       "1",
									"autoscaling.knative.dev/class":                        "kpa.autoscaling.knative.dev",
									"queue.sidecar.serving.knative.dev/resourcePercentage": "30",
								},
							},
							Spec: knservingv1.RevisionSpec{
								PodSpec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "user-container",
											Image: "asia.gcr.io/gcp-project-id/turing-router:latest",
											Ports: []corev1.ContainerPort{
												{
													Name:          "h2c",
													ContainerPort: 8080,
												},
											},
											Resources: resources,
											ReadinessProbe: &corev1.Probe{
												Handler: corev1.Handler{
													HTTPGet: &corev1.HTTPGetAction{
														Port: intstr.FromInt(8080),
														Path: "/v1/internal/ready",
													},
												},
												InitialDelaySeconds: 20,
												PeriodSeconds:       10,
												SuccessThreshold:    1,
												TimeoutSeconds:      5,
												FailureThreshold:    5,
											},
											VolumeMounts: baseSvc.VolumeMounts,
											Env:          envs,
										},
									},
									Volumes: baseSvc.Volumes,
								},
								TimeoutSeconds:       &timeout,
								ContainerConcurrency: &defaultConcurrency,
							},
						},
					},
					RouteSpec: defaultRouteSpec,
				},
			},
		},
		"annotations": {
			serviceCfg: KnativeService{
				BaseService:                     baseSvc,
				ContainerPort:                   8080,
				MinReplicas:                     5,
				MaxReplicas:                     6,
				AutoscalingMetric:               "memory",
				AutoscalingTarget:               "70",
				IsClusterLocal:                  false,
				UserContainerLimitRequestFactor: 1.5,
			},
			expectedSpec: knservingv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-svc",
					Namespace: "test-namespace",
					Labels: map[string]string{
						"labelKey": "labelVal",
					},
				},
				Spec: knservingv1.ServiceSpec{
					ConfigurationSpec: knservingv1.ConfigurationSpec{
						Template: knservingv1.RevisionTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Name: "test-svc-0",
								Labels: map[string]string{
									"labelKey": "labelVal",
								},
								Annotations: map[string]string{
									"autoscaling.knative.dev/minScale": "5",
									"autoscaling.knative.dev/maxScale": "6",
									"autoscaling.knative.dev/metric":   "memory",
									"autoscaling.knative.dev/target":   "358",
									"autoscaling.knative.dev/class":    "hpa.autoscaling.knative.dev",
								},
							},
							Spec: knservingv1.RevisionSpec{
								PodSpec:              podSpec,
								TimeoutSeconds:       &timeout,
								ContainerConcurrency: &defaultConcurrency,
							},
						},
					},
					RouteSpec: defaultRouteSpec,
				},
			},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			// Run test and validate
			svc, err := data.serviceCfg.BuildKnativeServiceConfig()
			require.NoError(t, err)
			err = tu.CompareObjects(*svc, data.expectedSpec)
			assert.NoError(t, err)
		})
	}
}

func TestGetAutoscalingTarget(t *testing.T) {
	tests := map[string]struct {
		cfg            *KnativeService
		expectedTarget string
		expectedErr    string
	}{
		"concurrency": {
			cfg: &KnativeService{
				AutoscalingMetric: "concurrency",
				AutoscalingTarget: "10",
			},
			expectedTarget: "10",
		},
		"rps": {
			cfg: &KnativeService{
				AutoscalingMetric: "rps",
				AutoscalingTarget: "100",
			},
			expectedTarget: "100",
		},
		"cpu": {
			cfg: &KnativeService{
				AutoscalingMetric: "cpu",
				AutoscalingTarget: "80",
			},
			expectedTarget: "80",
		},
		"memory | Mi": {
			cfg: &KnativeService{
				BaseService: &BaseService{
					MemoryRequests: resource.MustParse("512Mi"),
				},
				AutoscalingMetric: "memory",
				AutoscalingTarget: "70",
			},
			expectedTarget: "358", // (70/100) * (512 Mi)
		},
		"memory | Gi": {
			cfg: &KnativeService{
				BaseService: &BaseService{
					MemoryRequests: resource.MustParse("2Gi"),
				},
				AutoscalingMetric: "memory",
				AutoscalingTarget: "70",
			},
			expectedTarget: "1434", // (70/100) * (2Gi * 1024)
		},
		"memory | G": {
			cfg: &KnativeService{
				BaseService: &BaseService{
					MemoryRequests: resource.MustParse("2G"),
				},
				AutoscalingMetric: "memory",
				AutoscalingTarget: "70",
			},
			expectedTarget: "1335", // (70/100) * ((2G * 10^9) bytes / 1024^2)Mi
		},
		"memory | failure": {
			cfg: &KnativeService{
				BaseService: &BaseService{
					MemoryRequests: resource.MustParse("1Gi"),
				},
				AutoscalingMetric: "memory",
				AutoscalingTarget: "$^",
			},
			expectedErr: "strconv.ParseFloat: parsing \"$^\": invalid syntax",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := data.cfg.getAutoscalingTarget()
			if data.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, data.expectedErr)
			}
			assert.Equal(t, data.expectedTarget, result)
		})
	}
}
