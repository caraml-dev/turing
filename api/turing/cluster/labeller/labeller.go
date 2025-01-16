package labeller

import (
	"fmt"
	"regexp"
	"strconv"

	mlp "github.com/caraml-dev/mlp/api/client"
)

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
	// AppLabel refers to application of the Kubernetes Object
	AppLabel = "app"
	// managedLabel mark the kubernetes object as being managed by specific dev
	managedLabel = "caraml.dev/managed"
)

var (
	prefix      string
	nsPrefix    string
	environment string
)

var validLabelRegex = regexp.MustCompile("^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$")

var reservedKeys = map[string]bool{
	orchestratorLabel: true,
	environmentLabel:  true,
	streamLabel:       true,
	teamLabel:         true,
	AppLabel:          true,
}

// ValidLabel logic reused from
// https://github.com/caraml-dev/merlin/blob/06f121c6da05c5b5f1a28389e1078aaafed67541/api/models/metadata.go#L57
func IsValidLabel(name string) error {
	lengthOfName := len(name) < 64
	if !(lengthOfName) {
		return fmt.Errorf("length of name is greater than 63 characters")
	}

	if isValidName := validLabelRegex.MatchString(name); !isValidName {
		return fmt.Errorf("name violates kubernetes label constraint")
	}

	return nil
}

// InitKubernetesLabeller builds a new KubernetesLabeller Singleton
func InitKubernetesLabeller(p, ns, e string) {
	prefix = p
	nsPrefix = ns
	environment = e
}

// KubernetesLabelsRequest helps to build the Kubernetes labels needed for Kubernetes objects.
type KubernetesLabelsRequest struct {
	Stream string
	Team   string
	App    string
	Labels []mlp.Label
}

// GetLabelName prefixes the label with the config specified label and returns the formatted label name
func GetLabelName(name string) string {
	return fmt.Sprintf("%s%s", prefix, name)
}

// GetNamespaceLabelName prefixes the label with the specified label and returns the formatted label name
func GetNamespaceLabelName(name string) string {
	return fmt.Sprintf("%s%s", nsPrefix, name)
}

// BuildLabels builds the labels for the Kubernetes object
// Combines resource labels with project labels
func BuildLabels(r KubernetesLabelsRequest) map[string]string {
	labels := map[string]string{
		GetLabelName(orchestratorLabel): orchestratorValue,
		GetLabelName(streamLabel):       r.Stream,
		GetLabelName(teamLabel):         r.Team,
		GetLabelName(AppLabel):          r.App,
		GetLabelName(environmentLabel):  environment,
	}
	appendFromLabelsRequest(labels, r)
	return labels
}

// BuildNamespaceLabels builds the labels for a Kubernetes namespace.
// Combines resource labels with project labels
func BuildNamespaceLabels(r KubernetesLabelsRequest) map[string]string {
	labels := map[string]string{
		GetNamespaceLabelName(managedLabel): strconv.FormatBool(true),
	}
	appendFromLabelsRequest(labels, r)
	return labels
}

func appendFromLabelsRequest(labels map[string]string, r KubernetesLabelsRequest) {
	for _, label := range r.Labels {
		// skip label that is trying to override reserved key
		if _, usingReservedKeys := reservedKeys[prefix+label.Key]; usingReservedKeys {
			continue
		}

		// skip label that has invalid key name
		if err := IsValidLabel(label.Key); err != nil {
			continue
		}

		// skip label that has invalid value name
		if err := IsValidLabel(label.Value); err != nil {
			continue
		}

		labels[label.Key] = label.Value
	}
}
