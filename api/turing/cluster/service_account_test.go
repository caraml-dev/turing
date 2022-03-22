package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildServiceAccount(t *testing.T) {
	testNamespace := "namespace"
	saCfg := ServiceAccount{
		Name:      "sa-name",
		Namespace: testNamespace,
		Labels:    labels,
	}
	expected := corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sa-name",
			Namespace: testNamespace,
			Labels:    labels,
		},
	}
	got := saCfg.BuildServiceAccount()
	assert.Equal(t, expected, *got)
}
