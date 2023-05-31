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
	// componentLabel refers to the component that deployed the Kubernetes object
	componentLabel = "component"
	// AppLabel refers to application of the Kubernetes Object
	AppLabel = "app"
)

var (
	prefix      string
	environment string
)

// InitKubernetesLabeller builds a new KubernetesLabeller Singleton
func InitKubernetesLabeller(p, e string) {
	prefix = p
	environment = e
}

// KubernetesLabelsRequest helps to build the Kubernetes labels needed for Kubernetes objects.
type KubernetesLabelsRequest struct {
	Stream    string
	Team      string
	App       string
	Component string
}

// GetLabelName prefixes the label with the config specified label and returns the formatted label name
func GetLabelName(name string) string {
	return fmt.Sprintf("%s%s", prefix, name)
}

// BuildLabels builds the labels for the Kubernetes object
func BuildLabels(r KubernetesLabelsRequest) map[string]string {
	return map[string]string{
		GetLabelName(orchestratorLabel): orchestratorValue,
		GetLabelName(streamLabel):       r.Stream,
		GetLabelName(teamLabel):         r.Team,
		GetLabelName(AppLabel):          r.App,
		GetLabelName(componentLabel):    r.Component,
		GetLabelName(environmentLabel):  environment,
	}
}
