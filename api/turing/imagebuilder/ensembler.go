package imagebuilder

import (
	"fmt"

	"github.com/caraml-dev/mlp/api/pkg/artifact"
	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/models"
)

// NewEnsemblerJobImageBuilder create ImageBuilder for building docker image of ensembling job (batch)
func NewEnsemblerJobImageBuilder(
	clusterController cluster.Controller,
	imageBuildingConfig config.ImageBuildingConfig,
	artifactService artifact.Service,
) (ImageBuilder, error) {
	return newImageBuilder(
		clusterController,
		imageBuildingConfig,
		&ensemblerJobNameGenerator{registry: imageBuildingConfig.DestinationRegistry},
		models.EnsemblerRunnerTypeJob,
		artifactService,
	)
}

// ensemblerJobNameGenerator is name generator that will be used for building docker image of ensembling job
type ensemblerJobNameGenerator struct {
	registry string
}

// generateBuilderJobName generate pod name that will be used to build docker image of the ensembling job
func (n *ensemblerJobNameGenerator) generateBuilderName(
	projectName string,
	ensemblerName string,
	ensemblerID models.ID,
	versionID string,
) string {
	// Creates a unique resource name with partial versioning (part of the versionID hash) as max char count is limited
	// by k8s job name length (63)
	partialVersionID := getPartialVersionID(versionID, 5)
	return fmt.Sprintf("batch-%s-%s-%d-%s", projectName, ensemblerName, ensemblerID, partialVersionID)
}

// generateDockerImageName generate the name of docker image of prediction job that will be created from given model
func (n *ensemblerJobNameGenerator) generateDockerImageName(projectName string, ensemblerName string) string {
	return fmt.Sprintf("%s/%s/ensembler-jobs/%s", n.registry, projectName, ensemblerName)
}

// NewEnsemblerServiceImageBuilder create ImageBuilder for building docker image of the ensembling service (real-time)
func NewEnsemblerServiceImageBuilder(
	clusterController cluster.Controller,
	imageBuildingConfig config.ImageBuildingConfig,
	artifactService artifact.Service,
) (ImageBuilder, error) {
	return newImageBuilder(
		clusterController,
		imageBuildingConfig,
		&ensemblerServiceNameGenerator{registry: imageBuildingConfig.DestinationRegistry},
		models.EnsemblerRunnerTypeService,
		artifactService,
	)
}

// ensemblerServiceNameGenerator is name generator that will be used for building docker image of the ensembling service
type ensemblerServiceNameGenerator struct {
	registry string
}

// generateBuilderServiceName generate pod name that will be used to build docker image of the ensembling service
func (n *ensemblerServiceNameGenerator) generateBuilderName(
	projectName string,
	ensemblerName string,
	ensemblerID models.ID,
	versionID string,
) string {
	// Creates a unique resource name with partial versioning (part of the versionID hash) as max char count is limited
	// by k8s pod name length (63)
	partialVersionID := getPartialVersionID(versionID, 5)
	return fmt.Sprintf("service-%s-%s-%d-%s", projectName, ensemblerName, ensemblerID, partialVersionID)
}

// generateServiceImageName generate the name of docker image of the ensembling service that will be created from given
// model
func (n *ensemblerServiceNameGenerator) generateDockerImageName(projectName string, ensemblerName string) string {
	return fmt.Sprintf("%s/%s/ensembler-services/%s", n.registry, projectName, ensemblerName)
}

func getPartialVersionID(versionID string, numChar int) string {
	if len(versionID) > numChar {
		return versionID[:numChar]
	}
	return versionID
}
