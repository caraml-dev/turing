package imagebuilder

import (
	"fmt"

	"k8s.io/client-go/kubernetes"
)

// NewEnsemberJobImageBuilder create ImageBuilder for building docker image of prediction job (batch)
func NewEnsemberJobImageBuilder(
	kubeClient kubernetes.Interface,
	imageConfig ImageConfig,
	kanikoConfig KanikoConfig,
) (ImageBuilder, error) {
	return newImageBuilder(
		kubeClient,
		imageConfig,
		kanikoConfig,
		&ensemblerJobNameGenerator{registry: imageConfig.Registry},
	)
}

// ensemblerJobNameGenerator is name generator that will be used for building docker image of prediction job
type ensemblerJobNameGenerator struct {
	registry string
}

// generateBuilderJobName generate pod name that will be used to build docker image of the prediction job
func (n *ensemblerJobNameGenerator) generateBuilderJobName(projectName string, modelName string, versionID int) string {
	return fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, versionID)
}

// generateDockerImageName generate the name of docker image of prediction job that will be created from given model
func (n *ensemblerJobNameGenerator) generateDockerImageName(projectName string, modelName string) string {
	return fmt.Sprintf("%s/%s-%s-job", n.registry, projectName, modelName)
}
