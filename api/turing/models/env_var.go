package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	k8scorev1 "k8s.io/api/core/v1"
)

// EnvVar represents an environment variable present in a container.
type EnvVar struct {
	// Name of the environment variable.
	Name string `json:"name"`

	// Value of the environment variable.
	// Defaults to "".
	Value string `json:"value"`
}

// EnvVars is a list of environment variables to set in the container.
type EnvVars []*EnvVar

func (evs EnvVars) Value() (driver.Value, error) {
	return json.Marshal(evs)
}

func (evs *EnvVars) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &evs)
}

// ToKubernetesEnvVars returns the representation of Kubernetes'
// v1.EnvVars.
func (evs EnvVars) ToKubernetesEnvVars() []k8scorev1.EnvVar {
	kubeEnvVars := make([]k8scorev1.EnvVar, len(evs))

	for k, ev := range evs {
		kubeEnvVars[k] = k8scorev1.EnvVar{Name: ev.Name, Value: ev.Value}
	}

	return kubeEnvVars
}
