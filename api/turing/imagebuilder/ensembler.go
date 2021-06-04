package imagebuilder

import (
	"fmt"

	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/models"
)

// NewEnsemberJobImageBuilder create ImageBuilder for building docker image of prediction job (batch)
func NewEnsemberJobImageBuilder(
	clusterController cluster.Controller,
	imageConfig config.ImageConfig,
	kanikoConfig config.KanikoConfig,
) (ImageBuilder, error) {
	return newImageBuilder(
		clusterController,
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
func (n *ensemblerJobNameGenerator) generateBuilderJobName(
	projectName string,
	modelName string,
	versionID models.ID,
) string {
	return fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, versionID)
}

// generateDockerImageName generate the name of docker image of prediction job that will be created from given model
func (n *ensemblerJobNameGenerator) generateDockerImageName(projectName string, modelName string) string {
	return fmt.Sprintf("%s/%s-%s-job", n.registry, projectName, modelName)
}
