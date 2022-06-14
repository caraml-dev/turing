package cluster

import (
	"context"
	"fmt"
	"strconv"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

// Default values for Knative related resources
const KnativeServiceLabelKey = "serving.knative.dev/service"
const KnativeUserContainerName = "user-container"

// Define default values used in the creation of the knative service
var knativeSvcDefaults = struct {
	// AutoscalingClass holds the name of the default knative autoscaling class (Knative Pod Autoscaler)
	AutoscalingClass string
	// RequestTimeoutSeconds is the the max duration the instance is allowed for responding
	// to requests
	RequestTimeoutSeconds int
}{
	AutoscalingClass:      "kpa.autoscaling.knative.dev",
	RequestTimeoutSeconds: 30,
}

// KnativeService defines the properties for Knative services
type KnativeService struct {
	*BaseService

	IsClusterLocal bool  `json:"is_cluster_local"`
	ContainerPort  int32 `json:"containerPort"`

	// Autoscaling properties
	MinReplicas       int `json:"minReplicas"`
	MaxReplicas       int `json:"maxReplicas"`
	TargetConcurrency int `json:"targetConcurrency"`

	// Resource properties
	QueueProxyResourcePercentage    int     `json:"queueProxyResourcePercentage"`
	UserContainerLimitRequestFactor float64 `json:"userContainerLimitRequestFactor"`
}

// Creates a new config object compatible with the knative serving API, from
// the given config
func (cfg *KnativeService) BuildKnativeServiceConfig() *knservingv1.Service {
	// clone creates a copy of a map object
	clone := func(l map[string]string) map[string]string {
		ll := map[string]string{}
		for k, v := range l {
			ll[k] = v
		}
		return ll
	}

	kserviceLabels := clone(cfg.Labels)
	if cfg.IsClusterLocal {
		// Kservice should only be accessible from within the cluster
		// https://knative.dev/v1.2-docs/serving/services/private-services/
		kserviceLabels["networking.knative.dev/visibility"] = "cluster-local"
	}
	kserviceObjectMeta := cfg.buildSvcObjectMeta(kserviceLabels)

	revisionLabels := clone(cfg.Labels)
	kserviceSpec := cfg.buildSvcSpec(revisionLabels)

	svc := &knservingv1.Service{
		ObjectMeta: *kserviceObjectMeta,
		Spec:       *kserviceSpec,
	}
	// Call setDefaults on desired knative service here to avoid diffs generated because
	// knative defaulter webhook is called when creating or updating the knative service.
	// Ref: https://github.com/kserve/kserve/blob/v0.8.0/pkg/controller/v1beta1
	// /inferenceservice/reconcilers/knative/ksvc_reconciler.go#L159
	svc.SetDefaults(context.TODO())
	return svc
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
) *knservingv1.ServiceSpec {
	// Set max timeout for responding to requests
	timeout := int64(knativeSvcDefaults.RequestTimeoutSeconds)

	// Build annotations, set target concurrency of 1
	annotations := map[string]string{
		"autoscaling.knative.dev/minScale": strconv.Itoa(cfg.MinReplicas),
		"autoscaling.knative.dev/maxScale": strconv.Itoa(cfg.MaxReplicas),
		"autoscaling.knative.dev/target":   strconv.Itoa(cfg.TargetConcurrency),
		"autoscaling.knative.dev/class":    knativeSvcDefaults.AutoscalingClass,
	}

	if cfg.QueueProxyResourcePercentage > 0 {
		annotations["queue.sidecar.serving.knative.dev/resourcePercentage"] = strconv.Itoa(cfg.QueueProxyResourcePercentage)
	}

	// Revision name
	revisionName := fmt.Sprintf("%s-0", cfg.Name)

	// Build resource requirements for the user container
	resourceReqs := cfg.buildResourceReqs(cfg.UserContainerLimitRequestFactor)

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

	return &knservingv1.ServiceSpec{
		ConfigurationSpec: knservingv1.ConfigurationSpec{
			Template: knservingv1.RevisionTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:        revisionName,
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: knservingv1.RevisionSpec{
					PodSpec: corev1.PodSpec{
						Containers: []corev1.Container{container},
						Volumes:    cfg.Volumes,
					},
					TimeoutSeconds: &timeout,
				},
			},
		},
	}
}
