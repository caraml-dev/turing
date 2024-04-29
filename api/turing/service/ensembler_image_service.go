package service

import (
	"fmt"

	mlp "github.com/caraml-dev/mlp/api/client"
	"github.com/caraml-dev/turing/api/turing/cluster/labeller"
	"github.com/caraml-dev/turing/api/turing/imagebuilder"
	"github.com/caraml-dev/turing/api/turing/models"
)

type EnsemblerImagesService interface {
	ListImages(project *mlp.Project, ensembler *models.PyFuncEnsembler, runnerType models.EnsemblerRunnerType) ([]imagebuilder.EnsemblerImage, error)
	BuildImage(project *mlp.Project, ensembler *models.PyFuncEnsembler, runnerType models.EnsemblerRunnerType) error
}

type EnsemblerImagesListOptions struct {
	ProjectID           models.ID                  `schema:"project_id" validate:"required"`
	EnsemblerID         models.ID                  `schema:"ensembler_id" validate:"required"`
	EnsemblerRunnerType models.EnsemblerRunnerType `schema:"runner_type"`
}

type ensemblerImagesService struct {
	ensemblerJobImageBuilder     imagebuilder.ImageBuilder
	ensemblerServiceImageBuilder imagebuilder.ImageBuilder
}

func NewEnsemblerImagesService(ensemblerJobImageBuilder, ensemblerServiceImageBuilder imagebuilder.ImageBuilder) EnsemblerImagesService {
	return &ensemblerImagesService{
		ensemblerJobImageBuilder:     ensemblerJobImageBuilder,
		ensemblerServiceImageBuilder: ensemblerServiceImageBuilder,
	}
}

func (s *ensemblerImagesService) ListImages(project *mlp.Project, ensembler *models.PyFuncEnsembler, runnerType models.EnsemblerRunnerType) ([]imagebuilder.EnsemblerImage, error) {
	builders := []imagebuilder.ImageBuilder{}

	if runnerType == models.EnsemblerRunnerTypeJob {
		builders = append(builders, s.ensemblerJobImageBuilder)
	} else if runnerType == models.EnsemblerRunnerTypeService {
		builders = append(builders, s.ensemblerServiceImageBuilder)
	} else {
		builders = append(builders, s.ensemblerJobImageBuilder, s.ensemblerServiceImageBuilder)
	}

	images := []imagebuilder.EnsemblerImage{}
	for _, builder := range builders {
		image, err := builder.GetEnsemblerImage(project, ensembler)
		if err != nil {
			return nil, err
		}
		images = append(images, image)
	}

	return images, nil
}

func (s *ensemblerImagesService) BuildImage(project *mlp.Project, ensembler *models.PyFuncEnsembler, runnerType models.EnsemblerRunnerType) error {
	ib, err := s.getImageBuilder(runnerType)
	if err != nil {
		return err
	}

	request := imagebuilder.BuildImageRequest{
		ProjectName:  project.Name,
		ResourceName: ensembler.Name,
		ResourceID:   ensembler.ID,
		VersionID:    ensembler.RunID,
		ArtifactURI:  ensembler.ArtifactURI,
		BuildLabels: labeller.BuildLabels(
			labeller.KubernetesLabelsRequest{
				Stream: project.Stream,
				Team:   project.Team,
				App:    ensembler.Name,
				Labels: project.Labels,
			},
		),
		EnsemblerFolder: EnsemblerFolder,
		BaseImageRefTag: ensembler.PythonVersion,
	}

	if _, err := ib.BuildImage(request); err != nil {
		return err
	}

	return nil
}

func (s *ensemblerImagesService) getImageBuilder(runnerType models.EnsemblerRunnerType) (imagebuilder.ImageBuilder, error) {
	switch runnerType {
	case models.EnsemblerRunnerTypeJob:
		return s.ensemblerJobImageBuilder, nil
	case models.EnsemblerRunnerTypeService:
		return s.ensemblerServiceImageBuilder, nil
	default:
		return nil, fmt.Errorf("runner type %s is not supported", runnerType)
	}
}
