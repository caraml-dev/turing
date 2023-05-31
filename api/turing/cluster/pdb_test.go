package cluster

import (
	"reflect"
	"testing"

	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestPodDisruptionBudget_BuildPodDisruptionBudget(t *testing.T) {
	defaultLabels := map[string]string{
		"key": "value",
	}

	type fields struct {
		Name           string
		Namespace      string
		Labels         map[string]string
		MaxUnavailable string
		MinAvailable   string
		Selector       *metav1.LabelSelector
	}
	tests := []struct {
		name    string
		fields  fields
		want    *policyv1.PodDisruptionBudget
		wantErr bool
	}{
		{
			"invalid: empty maxUnavailable and minAvailable",
			fields{
				Name:      "test-pdb-turing-router",
				Namespace: "test-namespace",
				Labels:    defaultLabels,
			},
			nil,
			true,
		},
		{
			"only maxUnavailable",
			fields{
				Name:           "test-pdb-turing-router",
				Namespace:      "test-namespace",
				Labels:         defaultLabels,
				MaxUnavailable: "20%",
				Selector: &metav1.LabelSelector{
					MatchLabels: defaultLabels,
				},
			},
			&policyv1.PodDisruptionBudget{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pdb-turing-router",
					Namespace: "test-namespace",
					Labels:    defaultLabels,
				},
				Spec: policyv1.PodDisruptionBudgetSpec{
					MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "20%"},
					Selector: &metav1.LabelSelector{
						MatchLabels: defaultLabels,
					},
				},
			},
			false,
		},
		{
			"only minAvailable",
			fields{
				Name:         "test-pdb-turing-router",
				Namespace:    "test-namespace",
				Labels:       defaultLabels,
				MinAvailable: "20%",
				Selector: &metav1.LabelSelector{
					MatchLabels: defaultLabels,
				},
			},
			&policyv1.PodDisruptionBudget{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pdb-turing-router",
					Namespace: "test-namespace",
					Labels:    defaultLabels,
				},
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{Type: intstr.String, StrVal: "20%"},
					Selector: &metav1.LabelSelector{
						MatchLabels: defaultLabels,
					},
				},
			},
			false,
		},
		{
			"both maxUnavailable and minAvailable exist, choose minAvailable",
			fields{
				Name:           "test-pdb-turing-router",
				Namespace:      "test-namespace",
				Labels:         defaultLabels,
				MaxUnavailable: "20%",
				MinAvailable:   "20%",
				Selector: &metav1.LabelSelector{
					MatchLabels: defaultLabels,
				},
			},
			&policyv1.PodDisruptionBudget{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pdb-turing-router",
					Namespace: "test-namespace",
					Labels:    defaultLabels,
				},
				Spec: policyv1.PodDisruptionBudgetSpec{
					MinAvailable: &intstr.IntOrString{Type: intstr.String, StrVal: "20%"},
					Selector: &metav1.LabelSelector{
						MatchLabels: defaultLabels,
					},
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := PodDisruptionBudget{
				Name:           tt.fields.Name,
				Namespace:      tt.fields.Namespace,
				Labels:         tt.fields.Labels,
				MaxUnavailable: tt.fields.MaxUnavailable,
				MinAvailable:   tt.fields.MinAvailable,
				Selector:       tt.fields.Selector,
			}
			got, err := cfg.BuildPodDisruptionBudget()
			if (err != nil) != tt.wantErr {
				t.Errorf("PodDisruptionBudget.BuildPodDisruptionBudget() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PodDisruptionBudget.BuildPodDisruptionBudget() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
