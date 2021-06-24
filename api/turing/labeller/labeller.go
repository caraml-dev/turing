package labeller

import "fmt"

const (
	// orchestratorValue is the value of the orchestrator (which is Turing)
	orchestratorValue = "turing"
	// environmentLabel refers to the environment the Kubernetes object is in
	environmentLabel = "environment"
	// streamLabel refers to the stream the Kubernetes object it belongs to
	streamLabel = "stream"
	// teamLabel refers to the team the Kubernetes object it belongs to
	teamLabel = "team"
	// orchestratorLabel refers to the orchestrator that deployed the Kubernetes object
	orchestratorLabel = "orchestrator"
	// appLabel refers to application of the Kubernetes Object
	appLabel = "app"
)

var prefix string

// InitKubernetesLabeller builds a new KubernetesLabeller Singleton
func InitKubernetesLabeller(p string) {
	prefix = p
}

// KubernetesLabelsRequest helps to build the Kubernetes labels needed for Kubernetes objects.
type KubernetesLabelsRequest struct {
	Stream      string
	Team        string
	App         string
	Environment string
}

// BuildLabels builds the labels for the Kubernetes object
func BuildLabels(r KubernetesLabelsRequest) map[string]string {
	return map[string]string{
		fmt.Sprintf("%s%s", prefix, orchestratorLabel): orchestratorValue,
		fmt.Sprintf("%s%s", prefix, streamLabel):       r.Stream,
		fmt.Sprintf("%s%s", prefix, teamLabel):         r.Team,
		fmt.Sprintf("%s%s", prefix, appLabel):          r.App,
		fmt.Sprintf("%s%s", prefix, environmentLabel):  r.Environment,
	}
}
