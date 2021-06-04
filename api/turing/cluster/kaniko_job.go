package cluster

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	googleApplicationEnvVarName = "GOOGLE_APPLICATION_CREDENTIALS"
	kanikoSecretName            = "kaniko-secret"
	kanikoSecretMountpath       = "/secret"
	kanikoSecretFileName        = "kaniko-secret.json"
	imageBuilderContainerName   = "kaniko-builder"
)

var (
	jobCompletions            int32 = 1
	jobBackOffLimit           int32 = 3
	jobTTLSecondAfterComplete int32 = 3600 * 24
)

// KanikoJobSpec contains the information required to build a OCI image.
type KanikoJobSpec struct {
	JobName       string
	Namespace     string
	Labels        map[string]string
	Args          []string
	Image         string
	Version       string
	CPURequest    resource.Quantity
	MemoryRequest resource.Quantity
	CPULimit      resource.Quantity
	MemoryLimit   resource.Quantity
}

// BuildSpec builds the struct into a Kubernetes spec.
func (s *KanikoJobSpec) BuildSpec() *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.JobName,
			Namespace: s.Namespace,
			Labels:    s.Labels,
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
							Image: fmt.Sprintf("%s:%s", s.Image, s.Version),
							Args:  s.Args,
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
									corev1.ResourceCPU:    s.CPURequest,
									corev1.ResourceMemory: s.MemoryRequest,
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    s.CPULimit,
									corev1.ResourceMemory: s.MemoryLimit,
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
}
