package cluster

import (
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

// NameValuePair captures a pair of name and value
// TODO: DATA-2094: This may no longer be necessary
type NameValuePair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// BaseService defines the common properties of services that can be speficied
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
}

func (cfg *BaseService) buildResourceReqs() corev1.ResourceRequirements {
	reqs := map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    cfg.CPURequests,
		corev1.ResourceMemory: cfg.MemoryRequests,
	}

	// Set resource limits to twice the request
	cpuLimit := cfg.CPURequests.DeepCopy()
	cpuLimit.Add(cpuLimit)
	memoryLimit := cfg.MemoryRequests.DeepCopy()
	memoryLimit.Add(memoryLimit)
	limits := map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    cpuLimit,
		corev1.ResourceMemory: memoryLimit,
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

type Port struct {
	Name     string `json:"name"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol"`
}

// ConfigMap contains information to create a config map
type ConfigMap struct {
	Name     string `json:"name"`
	FileName string `json:"file_name"`
	Data     string `json:"data"`
}
