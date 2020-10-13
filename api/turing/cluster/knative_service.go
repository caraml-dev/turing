package cluster

import (
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
	knservingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
)

// Default values for Knative related resources
const KnativeServiceLabelKey = "serving.knative.dev/service"
const KnativeUserContainerName = "user-container"

// Define default values used in the creation of the knative service
var knativeSvcDefaults = struct {
	// AutoscalingClass holds the name of the default knative autoscaling class (Knative Pod Autoscaler)
	AutoscalingClass string
	// TargetConcurrency holds the target knative observed concurrecy value for autoscaling
	TargetConcurrency int
	// RequestTimeoutSeconds is the the max duration the instance is allowed for responding
	// to requests
	RequestTimeoutSeconds int
}{
	AutoscalingClass:      "kpa.autoscaling.knative.dev",
	TargetConcurrency:     1,
	RequestTimeoutSeconds: 30,
}

// KnativeService defines the properties for Knative services
type KnativeService struct {
	*BaseService

	IsClusterLocal bool  `json:"is_cluster_local"`
	ContainerPort  int32 `json:"containerPort"`

	// Autoscaling properties
	MinReplicas int `json:"minReplicas"`
	MaxReplicas int `json:"maxReplicas"`
}

// Creates a new config object compatible with the knative serving API, from
// the given config
func (cfg *KnativeService) BuildKnativeServiceConfig() *knservingv1alpha1.Service {
	// Build labels
	labels := cfg.Labels
	if cfg.IsClusterLocal {
		labels["serving.knative.dev/visibility"] = "cluster-local"
	}

	// Build object meta data
	objMeta := cfg.buildSvcObjectMeta(labels)

	// Build service spec
	svcSpec := cfg.buildSvcSpec(labels)

	// Return the knative service object
	return &knservingv1alpha1.Service{
		ObjectMeta: *objMeta,
		Spec:       *svcSpec,
	}
}

func (cfg *KnativeService) buildSvcObjectMeta(labels map[string]string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      cfg.Name,
		Namespace: cfg.Namespace,
		Labels:    labels,
	}
}

func (cfg *KnativeService) buildSvcSpec(
	labels map[string]string,
) *knservingv1alpha1.ServiceSpec {
	// Set max timeout for responding to requests
	timeout := int64(knativeSvcDefaults.RequestTimeoutSeconds)

	// Build annotations, set target concurrency of 1
	annotations := map[string]string{
		"autoscaling.knative.dev/minScale": strconv.Itoa(cfg.MinReplicas),
		"autoscaling.knative.dev/maxScale": strconv.Itoa(cfg.MaxReplicas),
		"autoscaling.knative.dev/target":   strconv.Itoa(knativeSvcDefaults.TargetConcurrency),
		"autoscaling.knative.dev/class":    knativeSvcDefaults.AutoscalingClass,
	}

	// Revision name
	revisionName := fmt.Sprintf("%s-0", cfg.Name)

	// Build resource requirements for the user container
	resourceReqs := cfg.buildResourceReqs()

	// Build container spec
	container := corev1.Container{
		Image: cfg.Image,
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: cfg.ContainerPort,
			},
		},
		Resources:    resourceReqs,
		VolumeMounts: cfg.VolumeMounts,
		Env:          cfg.Envs,
	}
	if cfg.LivenessHTTPGetPath != "" {
		container.LivenessProbe = cfg.buildContainerProbe(livenessProbeType, int(cfg.ProbePort))
	}
	if cfg.ReadinessHTTPGetPath != "" {
		container.ReadinessProbe = cfg.buildContainerProbe(readinessProbeType, int(cfg.ProbePort))
	}

	// Create the Knative service spec
	revSpec := knservingv1.RevisionSpec{
		TimeoutSeconds: &timeout,
		PodSpec: corev1.PodSpec{
			Containers: []corev1.Container{container},
			Volumes:    cfg.Volumes,
		},
	}

	return &knservingv1alpha1.ServiceSpec{
		ConfigurationSpec: knservingv1alpha1.ConfigurationSpec{
			Template: &knservingv1alpha1.RevisionTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:        revisionName,
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: knservingv1alpha1.RevisionSpec{
					RevisionSpec: revSpec,
				},
			},
		},
	}
}
