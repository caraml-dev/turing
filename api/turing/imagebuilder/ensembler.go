package imagebuilder

import (
	"fmt"

	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/models"
)

// NewEnsemblerJobImageBuilder create ImageBuilder for building docker image of ensembling job (batch)
func NewEnsemblerJobImageBuilder(
	clusterController cluster.Controller,
	imageBuildingConfig config.ImageBuildingConfig,
) (ImageBuilder, error) {
	return newImageBuilder(
		clusterController,
		imageBuildingConfig,
		&ensemblerJobNameGenerator{registry: imageBuildingConfig.DestinationRegistry},
	)
}

// ensemblerJobNameGenerator is name generator that will be used for building docker image of ensembling job
type ensemblerJobNameGenerator struct {
	registry string
}

// generateBuilderJobName generate pod name that will be used to build docker image of the ensembling job
func (n *ensemblerJobNameGenerator) generateBuilderName(
	projectName string,
	modelName string,
	versionID models.ID,
) string {
	return fmt.Sprintf("batch-builder-%s-%s-%d", projectName, modelName, versionID)
}

// generateDockerImageName generate the name of docker image of prediction job that will be created from given model
func (n *ensemblerJobNameGenerator) generateDockerImageName(projectName string, modelName string) string {
	return fmt.Sprintf("%s/%s-%s-job", n.registry, projectName, modelName)
}

// NewEnsemblerServiceImageBuilder create ImageBuilder for building docker image of the ensembling service (real-time)
func NewEnsemblerServiceImageBuilder(
	clusterController cluster.Controller,
	imageBuildingConfig config.ImageBuildingConfig,
) (ImageBuilder, error) {
	return newImageBuilder(
		clusterController,
		imageBuildingConfig,
		&ensemblerServiceNameGenerator{registry: imageBuildingConfig.DestinationRegistry},
	)
}

// ensemblerServiceNameGenerator is name generator that will be used for building docker image of the ensembling service
type ensemblerServiceNameGenerator struct {
	registry string
}

// generateBuilderServiceName generate pod name that will be used to build docker image of the ensembling service
func (n *ensemblerServiceNameGenerator) generateBuilderName(
	projectName string,
	modelName string,
	versionID models.ID,
) string {
	return fmt.Sprintf("service-builder-%s-%s-%d", projectName, modelName, versionID)
}

// generateServiceImageName generate the name of docker image of the ensembling service that will be created from given
// model
func (n *ensemblerServiceNameGenerator) generateDockerImageName(projectName string, modelName string) string {
	return fmt.Sprintf("%s/%s-%s-service", n.registry, projectName, modelName)
}
