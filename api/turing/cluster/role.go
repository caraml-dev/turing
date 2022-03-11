package cluster

import (
	apirbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Role contains the information to build a role
type Role struct {
	Name        string                 `json:"name"`
	Namespace   string                 `json:"namespace"`
	Labels      map[string]string      `json:"labels"`
	PolicyRules []apirbacv1.PolicyRule `json:"rules"`
}

func (r *Role) BuildRole() *apirbacv1.Role {
	return &apirbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		Rules: r.PolicyRules,
	}
}
