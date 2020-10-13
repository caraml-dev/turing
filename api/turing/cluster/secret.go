package cluster

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Secret defines a kubernetes secret.
type Secret struct {
	Name      string
	Namespace string
	Data      map[string]string
}

// BuildSecret builds a kubernetes secret from the given config.
func (cfg *Secret) BuildSecret() *corev1.Secret {
	data := make(map[string][]byte, len(cfg.Data))
	for k, v := range cfg.Data {
		data[k] = []byte(v)
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
		},
		Data: data,
		Type: corev1.SecretTypeOpaque,
	}
}
