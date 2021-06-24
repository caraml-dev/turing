package cluster

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Job contains the information to build a Kubernetes Job object
type Job struct {
	Name                    string
	Namespace               string
	Labels                  map[string]string
	Completions             *int32
	BackOffLimit            *int32
	TTLSecondsAfterFinished *int32
	RestartPolicy           corev1.RestartPolicy
	Containers              []Container
	SecretVolumes           []SecretVolume
}

// Build converts the spec into a Kubernetes spec
func (j *Job) Build() *batchv1.Job {
	containers := []corev1.Container{}
	for _, container := range j.Containers {
		containers = append(containers, container.Build())
	}

	volumes := []corev1.Volume{}
	for _, v := range j.SecretVolumes {
		volumes = append(volumes, v.Build())
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      j.Name,
			Namespace: j.Namespace,
			Labels:    j.Labels,
		},
		Spec: batchv1.JobSpec{
			Completions:             j.Completions,
			BackoffLimit:            j.BackOffLimit,
			TTLSecondsAfterFinished: j.TTLSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: j.RestartPolicy,
					Containers:    containers,
					Volumes:       volumes,
				},
			},
		},
	}
}
