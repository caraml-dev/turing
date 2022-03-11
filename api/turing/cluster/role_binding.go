package cluster

import (
	apirbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RoleBinding contains the information to build a role binding
type RoleBinding struct {
	Name               string            `json:"name"`
	Namespace          string            `json:"namespace"`
	Labels             map[string]string `json:"labels"`
	RoleName           string            `json:"role_name"`
	ServiceAccountName string            `json:"service_account_name"`
}

func (r *RoleBinding) BuildRoleBinding() *apirbacv1.RoleBinding {
	return &apirbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		Subjects: []apirbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Namespace: r.Namespace,
				Name:      r.ServiceAccountName,
			},
		},
		RoleRef: apirbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     r.RoleName,
		},
	}
}
