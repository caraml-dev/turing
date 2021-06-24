package cluster

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// Container contains the Kubernetes Container to be deployed
type Container struct {
	Name         string
	Image        string
	Args         []string
	VolumeMounts []VolumeMount
	Envs         []Env
	Resources    RequestLimitResources
}

// Build converts the spec into a Kubernetes spec
func (c *Container) Build() corev1.Container {
	volumeMounts := []corev1.VolumeMount{}
	for _, vm := range c.VolumeMounts {
		volumeMounts = append(volumeMounts, vm.Build())
	}

	envs := []corev1.EnvVar{}
	for _, e := range c.Envs {
		envs = append(envs, e.Build())
	}

	return corev1.Container{
		Name:         c.Name,
		Image:        c.Image,
		Args:         c.Args,
		VolumeMounts: volumeMounts,
		Env:          envs,
		Resources:    c.Resources.Build(),
	}
}

// VolumeMount is a Kubernetes VolumeMount
type VolumeMount struct {
	Name      string
	MountPath string
}

// Build converts the spec into a Kubernetes spec
func (vm *VolumeMount) Build() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      vm.Name,
		MountPath: vm.MountPath,
	}
}

// SecretVolume is a Kubernetes volume that mounted by a secret
type SecretVolume struct {
	Name       string
	SecretName string
}

// Build converts the spec into a Kubernetes spec
func (v *SecretVolume) Build() corev1.Volume {
	return corev1.Volume{
		Name: v.Name,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: v.SecretName,
			},
		},
	}
}

// Env is a Kubernetes environment variable
type Env struct {
	Name  string
	Value string
}

// Build converts the spec into a Kubernetes spec
func (e *Env) Build() corev1.EnvVar {
	return corev1.EnvVar{
		Name:  e.Name,
		Value: e.Value,
	}
}

// RequestLimitResources is a Kubernetes resource request and limits
type RequestLimitResources struct {
	Request Resource
	Limit   Resource
}

// Build converts the spec into a Kubernetes spec
func (r *RequestLimitResources) Build() corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: r.Request.Build(),
		Limits:   r.Limit.Build(),
	}
}

// Resource is a Kubernetes resource
type Resource struct {
	CPU    resource.Quantity
	Memory resource.Quantity
}

// Build converts the spec into a Kubernetes spec
func (r *Resource) Build() corev1.ResourceList {
	return corev1.ResourceList{
		corev1.ResourceCPU:    r.CPU,
		corev1.ResourceMemory: r.Memory,
	}
}
