package cluster

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolumeClaim contains the information to build a persistent volume claim
type PersistentVolumeClaim struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	AccessModes []string          `json:"access_modes"`
	Size        resource.Quantity `json:"size"`
	Labels      map[string]string `json:"labels"`
}

func (pvc *PersistentVolumeClaim) BuildPersistentVolumeClaim() *corev1.PersistentVolumeClaim {
	accessModes := make([]corev1.PersistentVolumeAccessMode, len(pvc.AccessModes))
	for i := range pvc.AccessModes {
		accessModes[i] = corev1.PersistentVolumeAccessMode(pvc.AccessModes[i])
	}
	volumeMode := corev1.PersistentVolumeFilesystem
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvc.Name,
			Namespace: pvc.Namespace,
			Labels:    pvc.Labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: accessModes,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": pvc.Size,
				},
			},
			VolumeMode: &volumeMode,
		},
	}
}
