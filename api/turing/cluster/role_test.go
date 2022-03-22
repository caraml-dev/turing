package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apirbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildRole(t *testing.T) {
	testNamespace := "namespace"
	roleCfg := Role{
		Name:      "role-name",
		Namespace: testNamespace,
		Labels:    labels,
	}
	expected := apirbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "role-name",
			Namespace: testNamespace,
			Labels:    labels,
		},
	}
	got := roleCfg.BuildRole()
	assert.Equal(t, expected, *got)
}
