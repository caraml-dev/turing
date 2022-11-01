package batchensembling

// Mocking here is tricky due to cyclic imports.
// To circumvent this, use the --inpackage flag in mockery
// i.e. mockery --name=EnsemblingController --case underscore --inpackage

import (
	"testing"
	"time"

	mlp "github.com/gojek/mlp/api/client"
	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/turing/api/turing/imagebuilder"
	imagebuildermock "github.com/caraml-dev/turing/api/turing/imagebuilder/mocks"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	servicemock "github.com/caraml-dev/turing/api/turing/service/mocks"
)

func TestRun(t *testing.T) {
	// Unfortunately this is hard to test as we need Kubernetes integration
	// and a Spark Operator. Testing with an actual cluster is required.
	// Here we just try to run it without throwing an exception.
	mockEnsemblersService := func() service.EnsemblersService {
		svc := &servicemock.EnsemblersService{}
		svc.On("FindByID", mock.Anything, mock.Anything).Return(&models.PyFuncEnsembler{}, nil)
		return svc
	}
	var tests = map[string]struct {
		ensemblingController func() EnsemblingController
		imageBuilder         func() imagebuilder.ImageBuilder
		ensemblingJobService func() service.EnsemblingJobService
		ensemblersService    func() service.EnsemblersService
		mlpService           func() service.MLPService
	}{
		"success | nominal": {
			ensemblingController: func() EnsemblingController {
				ctlr := &MockEnsemblingController{}
				ctlr.On(
					"Create",
					mock.Anything,
				).Return(nil)
				ctlr.On(
					"GetStatus",
					mock.Anything,
					mock.Anything,
				).Return(SparkApplicationStateCompleted, nil)
				return ctlr
			},
			imageBuilder: func() imagebuilder.ImageBuilder {
				ib := &imagebuildermock.ImageBuilder{}
				ib.On(
					"BuildImage",
					mock.Anything,
					mock.Anything,
				).Return("ghcr.io/test-project/mymodel:1", nil)
				return ib
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &servicemock.EnsemblingJobService{}

				svc.On(
					"List",
					mock.Anything,
				).Return(&service.PaginatedResults{
					Results: []*models.EnsemblingJob{generateEnsemblingJobFixture()},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				}, nil).Once()

				svc.On(
					"Save",
					mock.Anything,
				).Return(nil)

				newFixture := generateEnsemblingJobFixture()
				newFixture.Status = models.JobRunning
				svc.On(
					"List",
					mock.Anything,
				).Return(&service.PaginatedResults{
					Results: []*models.EnsemblingJob{newFixture},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				}, nil)

				svc.On("FindByID", mock.Anything, mock.Anything).Return(
					&models.EnsemblingJob{},
					nil,
				)
				return svc
			},
			ensemblersService: mockEnsemblersService,
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetProject",
					mock.Anything,
					mock.Anything,
				).Return(&mlp.Project{Id: 1}, nil)
				return svc
			},
		},
		"success | imagebuilding stuck": {
			ensemblingController: func() EnsemblingController {
				ctlr := &MockEnsemblingController{}
				ctlr.On(
					"Create",
					mock.Anything,
				).Return(nil)
				return ctlr
			},
			imageBuilder: func() imagebuilder.ImageBuilder {
				ib := &imagebuildermock.ImageBuilder{}
				ib.On(
					"BuildImage",
					mock.Anything,
					mock.Anything,
				).Return("ghcr.io/test-project/mymodel:1", nil)
				ib.On(
					"GetImageBuildingJobStatus",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(imagebuilder.JobStatusFailed, nil)
				return ib
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &servicemock.EnsemblingJobService{}

				svc.On(
					"List",
					mock.Anything,
				).Return(&service.PaginatedResults{
					Results: []*models.EnsemblingJob{generateEnsemblingJobFixture()},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				}, nil).Once()

				svc.On(
					"Save",
					mock.Anything,
				).Return(nil)

				newFixture := generateEnsemblingJobFixture()
				newFixture.Status = models.JobBuildingImage
				svc.On(
					"List",
					mock.Anything,
				).Return(&service.PaginatedResults{
					Results: []*models.EnsemblingJob{newFixture},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				}, nil)

				svc.On("FindByID", mock.Anything, mock.Anything).Return(
					&models.EnsemblingJob{},
					nil,
				)

				return svc
			},
			ensemblersService: mockEnsemblersService,
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetProject",
					mock.Anything,
					mock.Anything,
				).Return(&mlp.Project{Id: 1}, nil)
				return svc
			},
		},
		"success | no ensembling jobs": {
			ensemblingController: func() EnsemblingController {
				ctlr := &MockEnsemblingController{}
				ctlr.On(
					"Create",
					mock.Anything,
				).Return(nil)
				return ctlr
			},
			imageBuilder: func() imagebuilder.ImageBuilder {
				ib := &imagebuildermock.ImageBuilder{}
				ib.On(
					"BuildImage",
					mock.Anything,
					mock.Anything,
				).Return("ghcr.io/test-project/mymodel:1", nil)
				return ib
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &servicemock.EnsemblingJobService{}

				svc.On(
					"List",
					mock.Anything,
				).Return(&service.PaginatedResults{
					Results: []*models.EnsemblingJob{},
					Paging: service.Paging{
						Total: 0,
						Page:  1,
						Pages: 1,
					},
				}, nil)

				svc.On(
					"Save",
					mock.Anything,
				).Return(nil)

				svc.On("FindByID", mock.Anything, mock.Anything).Return(
					&models.EnsemblingJob{},
					nil,
				)

				return svc
			},
			ensemblersService: mockEnsemblersService,
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetProject",
					mock.Anything,
					mock.Anything,
				).Return(&mlp.Project{Id: 1}, nil)
				return svc
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ensemblingController := tt.ensemblingController()
			ensemblingJobService := tt.ensemblingJobService()
			ensemblersService := tt.ensemblersService()
			mlpService := tt.mlpService()
			imageBuilder := tt.imageBuilder()

			r := NewBatchEnsemblingJobRunner(
				ensemblingController,
				ensemblingJobService,
				ensemblersService,
				mlpService,
				imageBuilder,
				10,
				3,
				10*time.Minute,
				10*time.Second,
			)
			r.Run()
		})
	}
}
