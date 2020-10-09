// +build unit

package models

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestEnvVarsValue(t *testing.T) {
	vars := EnvVars{
		{
			Name:  "n1",
			Value: "v1",
		},
		{
			Name:  "n2",
			Value: "v2",
		},
	}

	value, err := vars.Value()
	// Convert to string for comparison
	byteValue, ok := value.([]byte)
	assert.True(t, ok)
	// Validate
	assert.NoError(t, err)
	assert.JSONEq(t, `
		[{
			"name": "n1",
			"value": "v1"
		},
		{
			"name": "n2",
			"value": "v2"
		}]
	`, string(byteValue))
}

func TestEnvVarsScan(t *testing.T) {
	tests := map[string]struct {
		value    interface{}
		success  bool
		expected EnvVars
		err      string
	}{
		"success": {
			value:   []byte(`[{"name": "n1","value": "v1"}]`),
			success: true,
			expected: EnvVars{
				{
					Name:  "n1",
					Value: "v1",
				},
			},
		},
		"failure | invalid value": {
			value:   100,
			success: false,
			err:     "type assertion to []byte failed",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			var envs EnvVars
			err := envs.Scan(data.value)
			if data.success {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, envs)
			} else {
				assert.Error(t, err)
				assert.Equal(t, data.err, err.Error())
			}
		})
	}
}

func TestEnvVarsToKubernetesEnvVars(t *testing.T) {
	tests := []struct {
		name string
		e    EnvVars
		want []corev1.EnvVar
	}{
		{
			"empty",
			EnvVars{},
			[]corev1.EnvVar{},
		},
		{
			"1",
			EnvVars{
				&EnvVar{Name: "foo", Value: "bar"},
			},
			[]corev1.EnvVar{
				{Name: "foo", Value: "bar"},
			},
		},
		{
			"2",
			EnvVars{
				&EnvVar{Name: "1", Value: "1"},
				&EnvVar{Name: "2", Value: "2"},
			},
			[]corev1.EnvVar{
				{Name: "1", Value: "1"},
				{Name: "2", Value: "2"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.ToKubernetesEnvVars(); !cmp.Equal(got, tt.want) {
				t.Errorf("EnvVars.ToKubernetesEnvVars() = %v, want %v", got, tt.want)
			}
		})
	}
}
