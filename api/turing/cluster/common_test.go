package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	containerName   = "kaniko-builder"
	imageName       = "imagename"
	argsLine        = "argsline"
	secretName      = "kaniko-secret"
	secretMountPath = "/secret"
	envVarName      = "envvarname"
	envVarValue     = "valuemeal"
)

var (
	cpu    = resource.MustParse("500m")
	memory = resource.MustParse("500Mi")
)

func CreateContainer() Container {
	return Container{

		Name:  containerName,
		Image: imageName,
		Args:  []string{argsLine},
		VolumeMounts: []VolumeMount{
			{
				Name:      secretName,
				MountPath: secretMountPath,
			},
		},
		Envs: []Env{
			{
				Name:  envVarName,
				Value: envVarValue,
			},
		},
		Resources: RequestLimitResources{
			Request: Resource{
				CPU:    cpu,
				Memory: memory,
			},
			Limit: Resource{
				CPU:    cpu,
				Memory: memory,
			},
		},
	}
}

func CreateKubernetesContainer() corev1.Container {
	return corev1.Container{
		Name:  containerName,
		Image: imageName,
		Args:  []string{argsLine},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      secretName,
				MountPath: secretMountPath,
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  envVarName,
				Value: envVarValue,
			},
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    cpu,
				corev1.ResourceMemory: memory,
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    cpu,
				corev1.ResourceMemory: memory,
			},
		},
	}
}

func TestContainer(t *testing.T) {
	expected := CreateKubernetesContainer()
	c := CreateContainer()

	assert.Equal(t, expected, c.Build())
}

func CreateSecretVolume() SecretVolume {
	return SecretVolume{
		Name:       secretName,
		SecretName: secretName,
	}
}

func CreateKubernetesSecretVolume() corev1.Volume {
	return corev1.Volume{
		Name: secretName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretName,
			},
		},
	}
}

func TestSecretVolume(t *testing.T) {
	expected := CreateKubernetesSecretVolume()
	v := CreateSecretVolume()

	assert.Equal(t, expected, v.Build())
}
