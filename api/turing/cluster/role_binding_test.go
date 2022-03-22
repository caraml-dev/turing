package cluster

import (
	"testing"

	"github.com/stretchr/testify/assert"
	apirbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestBuildRoleBinding(t *testing.T) {
	testNamespace := "namespace"
	serviceAccountName := "service-account-name"
	roleName := "role-name"
	roleBindingName := "role-binding-name"
	roleBindingCfg := RoleBinding{
		Name:               roleBindingName,
		Namespace:          testNamespace,
		Labels:             labels,
		RoleName:           roleName,
		ServiceAccountName: serviceAccountName,
	}
	expected := apirbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      roleBindingName,
			Namespace: testNamespace,
			Labels:    labels,
		},
		Subjects: []apirbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Namespace: testNamespace,
				Name:      serviceAccountName,
			},
		},
		RoleRef: apirbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     roleName,
		},
	}
	got := roleBindingCfg.BuildRoleBinding()
	assert.Equal(t, expected, *got)
}
