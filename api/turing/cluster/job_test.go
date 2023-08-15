package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace = "test-ns"
)

var (
	labels = map[string]string{
		"foo": "bar",
	}
	annotations = map[string]string{
		"key": "value",
	}
	jobCompletions            int32 = 1
	jobBackOffLimit           int32 = 3
	jobTTLSecondAfterComplete int32 = 3600 * 24
)

func TestJob(t *testing.T) {
	expected := batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        jobName,
			Namespace:   namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
			Completions:             &jobCompletions,
			BackoffLimit:            &jobBackOffLimit,
			TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						CreateKubernetesContainer(),
					},
					Volumes: []corev1.Volume{
						CreateKubernetesSecretVolume(),
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "tolerate-this",
							Operator: corev1.TolerationOpEqual,
							Value:    "true",
							Effect:   corev1.TaintEffectNoSchedule,
						},
					},
					NodeSelector: map[string]string{
						"node-workload-type": "image",
					},
					ServiceAccountName: serviceAccountName,
				},
			},
		},
	}

	tolerationName := "tolerate-this"

	j := Job{
		Name:                    jobName,
		Namespace:               namespace,
		Labels:                  labels,
		Annotations:             annotations,
		Completions:             &jobCompletions,
		BackOffLimit:            &jobBackOffLimit,
		TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
		RestartPolicy:           corev1.RestartPolicyNever,
		Containers: []Container{
			CreateContainer(),
		},
		SecretVolumes: []SecretVolume{
			CreateSecretVolume(),
		},
		TolerationName: &tolerationName,
		NodeSelector: map[string]string{
			"node-workload-type": "image",
		},
		ServiceAccount: serviceAccountName,
	}

	assert.Equal(t, expected, *j.Build())
}
