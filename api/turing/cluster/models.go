package cluster

import (
	"math"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var healthCheckDefaults = struct {
	InitDelaySeconds int32
	PeriodSeconds    int32
	FailureThreshold int32
	TimeoutSeconds   int32
}{
	InitDelaySeconds: 20,
	PeriodSeconds:    10,
	FailureThreshold: 5,
	TimeoutSeconds:   5,
}

type probeType string

const (
	livenessProbeType  probeType = "LivenessProbe"
	readinessProbeType probeType = "ReadinessProbe"
)

// BaseService defines the common properties of services that can be specified
// for its deployment by the cluster controller
type BaseService struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Image     string `json:"image"`

	// Resources
	CPURequests    resource.Quantity `json:"cpu_requests"`
	MemoryRequests resource.Quantity `json:"memory_requests"`

	// Health Checks
	ProbePort             int32  `json:"probe_port"`
	LivenessHTTPGetPath   string `json:"liveness_path"`
	ReadinessHTTPGetPath  string `json:"readiness_path"`
	ProbeInitDelaySeconds int32  `json:"probe_delay_seconds"`

	// Env vars
	Envs []corev1.EnvVar `json:"envs"`

	// Labels
	Labels map[string]string `json:"labels"`

	ConfigMap *ConfigMap

	PersistentVolumeClaim *PersistentVolumeClaim `json:"persistent_volume_claim"`

	// Volumes
	Volumes      []corev1.Volume      `json:"volumes"`
	VolumeMounts []corev1.VolumeMount `json:"volume_mounts"`

	// InitContainers
	InitContainers []Container `json:"init_containers"`
}

func (cfg *BaseService) buildResourceReqs(
	UserContainerCPULimitRequestFactor float64,
	UserContainerMemoryLimitRequestFactor float64,
) corev1.ResourceRequirements {
	reqs := map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    cfg.CPURequests,
		corev1.ResourceMemory: cfg.MemoryRequests,
	}

	// Set resource limits to request * userContainerCPULimitRequestFactor or UserContainerMemoryLimitRequestFactor
	limits := map[corev1.ResourceName]resource.Quantity{}
	if UserContainerCPULimitRequestFactor != 0 {
		limits[corev1.ResourceCPU] = ComputeResource(cfg.CPURequests, UserContainerCPULimitRequestFactor)
	}
	if UserContainerMemoryLimitRequestFactor != 0 {
		limits[corev1.ResourceMemory] = ComputeResource(cfg.MemoryRequests, UserContainerMemoryLimitRequestFactor)
	}

	return corev1.ResourceRequirements{
		Limits:   limits,
		Requests: reqs,
	}
}

func (cfg *BaseService) buildContainerProbe(ptype probeType, port int) *corev1.Probe {
	// Apply default init delay if unset
	if cfg.ProbeInitDelaySeconds == 0 {
		cfg.ProbeInitDelaySeconds = healthCheckDefaults.InitDelaySeconds
	}
	buildProbe := func(httpPath string, port int) *corev1.Probe {
		probe := &corev1.Probe{
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Path: httpPath,
				},
			},
			InitialDelaySeconds: cfg.ProbeInitDelaySeconds,
			TimeoutSeconds:      healthCheckDefaults.TimeoutSeconds,
			PeriodSeconds:       healthCheckDefaults.PeriodSeconds,
			FailureThreshold:    healthCheckDefaults.FailureThreshold,
		}
		if port != 0 {
			probe.Handler.HTTPGet.Port = intstr.FromInt(port)
		}
		return probe
	}

	switch ptype {
	case livenessProbeType:
		return buildProbe(cfg.LivenessHTTPGetPath, port)
	case readinessProbeType:
		return buildProbe(cfg.ReadinessHTTPGetPath, port)
	}
	return nil
}

func (cfg *BaseService) buildInitContainer(initContainerSpecs []Container) []corev1.Container {
	initContainers := make([]corev1.Container, 0)
	for _, initContainerSpec := range initContainerSpecs {
		initContainers = append(initContainers, initContainerSpec.Build())
	}
	return initContainers
}

type Port struct {
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// ConfigMap contains information to create a config map
type ConfigMap struct {
	Name     string            `json:"name"`
	FileName string            `json:"file_name"`
	Data     string            `json:"data"`
	Labels   map[string]string `json:"labels"`
}

// Ref:
// https://github.com/knative/serving/blob/release-0.14/pkg/reconciler/revision/resources/queue.go#L115
func ComputeResource(resourceQuantity resource.Quantity, fraction float64) resource.Quantity {
	scaledValue := resourceQuantity.Value()
	scaledMilliValue := int64(math.MaxInt64 - 1)
	if scaledValue < (math.MaxInt64 / 1000) {
		scaledMilliValue = resourceQuantity.MilliValue()
	}

	percentageValue := float64(scaledMilliValue) * fraction
	newValue := int64(math.MaxInt64)
	if percentageValue < math.MaxInt64 {
		newValue = int64(percentageValue)
	}

	newquantity := resource.NewMilliQuantity(newValue, resource.BinarySI)
	return *newquantity
}
