package cluster

import (
	"fmt"

	apipolicyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type PodDisruptionBudget struct {
	Name           string                `json:"name"`
	Namespace      string                `json:"namespace"`
	Labels         map[string]string     `json:"labels"`
	MaxUnavailable string                `json:"max_unavailable"`
	MinAvailable   string                `json:"min_available"`
	Selector       *metav1.LabelSelector `json:"selector"`
}

func (cfg PodDisruptionBudget) BuildPodDisruptionBudget() (*apipolicyv1.PodDisruptionBudget, error) {
	if cfg.MaxUnavailable == "" && cfg.MinAvailable == "" {
		return nil, fmt.Errorf("one of maxUnavailable and minAvailable must be specified")
	}

	spec := apipolicyv1.PodDisruptionBudgetSpec{
		Selector: cfg.Selector,
	}

	if cfg.MaxUnavailable != "" {
		maxUnavailable := intstr.FromString(cfg.MaxUnavailable)
		spec.MaxUnavailable = &maxUnavailable
	}

	// Since we can specify only one of maxUnavailable and minAvailable, minAvailable takes precedence
	// https://kubernetes.io/docs/tasks/run-application/configure-pdb/#specifying-a-poddisruptionbudget
	if cfg.MinAvailable != "" {
		minAvailable := intstr.FromString(cfg.MinAvailable)
		spec.MinAvailable = &minAvailable
		spec.MaxUnavailable = nil
	}

	pdb := &apipolicyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
			Labels:    cfg.Labels,
		},
		Spec: spec,
	}
	return pdb, nil
}
