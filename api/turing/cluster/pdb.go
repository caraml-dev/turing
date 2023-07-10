package cluster

import (
	"fmt"

	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	metav1cfg "k8s.io/client-go/applyconfigurations/meta/v1"
	policyv1cfg "k8s.io/client-go/applyconfigurations/policy/v1"
)

type PodDisruptionBudget struct {
	Name                     string                   `json:"name"`
	Namespace                string                   `json:"namespace"`
	Labels                   map[string]string        `json:"labels"`
	MaxUnavailablePercentage *int                     `json:"max_unavailable_percentage"`
	MinAvailablePercentage   *int                     `json:"min_available_percentage"`
	Selector                 *apimetav1.LabelSelector `json:"selector"`
}

func (cfg PodDisruptionBudget) BuildPDBSpec() (*policyv1cfg.PodDisruptionBudgetSpecApplyConfiguration, error) {
	if cfg.MaxUnavailablePercentage == nil && cfg.MinAvailablePercentage == nil {
		return nil, fmt.Errorf("one of maxUnavailable and minAvailable must be specified")
	}

	pdbSpec := &policyv1cfg.PodDisruptionBudgetSpecApplyConfiguration{}

	if cfg.Selector != nil && cfg.Selector.MatchLabels != nil {
		pdbSpec.Selector = &metav1cfg.LabelSelectorApplyConfiguration{
			MatchLabels: cfg.Selector.MatchLabels,
		}
	}

	// Since we can specify only one of maxUnavailable and minAvailable, minAvailable takes precedence
	// https://kubernetes.io/docs/tasks/run-application/configure-pdb/#specifying-a-poddisruptionbudget
	if cfg.MinAvailablePercentage != nil {
		minAvailable := intstr.FromString(fmt.Sprintf("%d%%", *cfg.MinAvailablePercentage))
		pdbSpec.MinAvailable = &minAvailable
	} else if cfg.MaxUnavailablePercentage != nil {
		maxUnavailable := intstr.FromString(fmt.Sprintf("%d%%", *cfg.MaxUnavailablePercentage))
		pdbSpec.MaxUnavailable = &maxUnavailable
	}

	return pdbSpec, nil
}
