package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	k8scorev1 "k8s.io/api/core/v1"
)

// Secret represents an MLP secret present in a container that is mounted as an environment variable
type Secret struct {
	// Name of the secret as stored in MLP
	MLPSecretName string `json:"mlp_secret_name"`

	// Name of the environment variable when the secret is mounted
	EnvVarName string `json:"env_var_name"`
}

// Secret is a list of MLP secrets to set in the container.
type Secrets []Secret

func (sec Secrets) Value() (driver.Value, error) {
	return json.Marshal(sec)
}

func (sec *Secrets) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &sec)
}

// ToKubernetesEnvVars returns the representation of Kubernetes'
// v1.EnvVars.
func (sec Secrets) ToKubernetesEnvVars(secretKeyRefName string) []k8scorev1.EnvVar {
	kubeEnvVars := make([]k8scorev1.EnvVar, len(sec))

	for k, secret := range sec {
		kubeEnvVars[k] = k8scorev1.EnvVar{
			Name: secret.EnvVarName,
			ValueFrom: &k8scorev1.EnvVarSource{
				SecretKeyRef: &k8scorev1.SecretKeySelector{
					LocalObjectReference: k8scorev1.LocalObjectReference{
						Name: secretKeyRefName,
					},
					Key: secret.MLPSecretName,
				},
			},
		}
	}

	return kubeEnvVars
}
