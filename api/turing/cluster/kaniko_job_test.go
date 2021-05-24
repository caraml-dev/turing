package cluster

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewKanikoJob(t *testing.T) {
	expectedSpecs := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bob",
			Namespace: "builder",
			Labels: map[string]string{
				"builderName": "bob",
			},
		},
		Spec: batchv1.JobSpec{
			Completions:             &jobCompletions,
			BackoffLimit:            &jobBackOffLimit,
			TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:  imageBuilderContainerName,
							Image: fmt.Sprintf("%s:%s", "gcr.io/kaniko-project/executor", "v1.5.2"),
							Args:  []string{"--dockerfile=Dockerfile"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      kanikoSecretName,
									MountPath: kanikoSecretMountpath,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  googleApplicationEnvVarName,
									Value: fmt.Sprintf("%s/%s", kanikoSecretMountpath, kanikoSecretFileName),
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: kanikoSecretName,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: kanikoSecretName,
								},
							},
						},
					},
				},
			},
		},
	}

	spec := &KanikoJobSpec{
		PodName:   "bob",
		Namespace: "builder",
		Labels: map[string]string{
			"builderName": "bob",
		},
		Args: []string{
			"--dockerfile=Dockerfile",
		},
		Image:         "gcr.io/kaniko-project/executor",
		Version:       "v1.5.2",
		CPURequest:    resource.MustParse("1"),
		MemoryRequest: resource.MustParse("1Gi"),
		CPULimit:      resource.MustParse("1"),
		MemoryLimit:   resource.MustParse("1Gi"),
	}
	kubeSpecs := spec.BuildSpec()
	assert.Equal(t, expectedSpecs, *kubeSpecs)
}
