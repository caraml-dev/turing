package cluster

import (
	"testing"

	tu "github.com/gojek/turing/api/turing/internal/testutils"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildSecret(t *testing.T) {
	secret := Secret{
		Name:      "svc-account",
		Namespace: "test-project",
		Data: map[string]string{
			"key.json": "asdf",
		},
		Labels: map[string]string{
			"key": "val",
		},
	}
	expected := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "svc-account",
			Namespace: "test-project",
			Labels: map[string]string{
				"key": "val",
			},
		},
		Data: map[string][]byte{
			"key.json": []byte("asdf"),
		},
		Type: corev1.SecretTypeOpaque,
	}
	got := secret.BuildSecret()
	err := tu.CompareObjects(*got, expected)
	assert.NilError(t, err)
}
