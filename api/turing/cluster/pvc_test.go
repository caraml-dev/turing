// +build unit

package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildPVC(t *testing.T) {
	qty, err := resource.ParseQuantity("2Gi")
	assert.NoError(t, err)
	testNamespace := "namespace"
	pvcCfg := PersistentVolumeClaim{
		Name:        "cache-volume",
		Namespace:   testNamespace,
		AccessModes: []string{"ReadWriteOnce"},
		Size:        qty,
	}
	volumeMode := corev1.PersistentVolumeFilesystem
	expected := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cache-volume",
			Namespace: testNamespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": qty,
				},
			},
			VolumeMode: &volumeMode,
		},
	}
	got := pvcCfg.BuildPersistentVolumeClaim()
	assert.Equal(t, expected, *got)
}
