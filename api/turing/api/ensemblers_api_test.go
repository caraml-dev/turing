package api

import (
	"errors"
	"testing"

	"github.com/caraml-dev/mlp/api/pkg/client/mlflow"
	mlflowMock "github.com/caraml-dev/mlp/api/pkg/client/mlflow/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/internal/ref"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/api/turing/service/mocks"
	"github.com/caraml-dev/turing/api/turing/validation"
)

func TestEnsemblersController_ListEnsemblers(t *testing.T) {
	empty := &service.PaginatedResults{
		Results: []*models.GenericEnsembler{},
		Paging:  service.Paging{Total: 0, Page: 1, Pages: 1},
	}
	ensemblers := &service.PaginatedResults{
		Results: []models.EnsemblerLike{
			&models.GenericEnsembler{
				Model:     models.Model{ID: 1},
				ProjectID: 3,
				Type:      models.EnsemblerPyFuncType,
				Name:      "test-ensembler-1",
			},
			&models.GenericEnsembler{
				Model:     models.Model{ID: 2},
				ProjectID: 3,
				Type:      models.EnsemblerPyFuncType,
				Name:      "test-ensembler-2",
			},
		},
		Paging: service.Paging{Total: 3, Page: 1, Pages: 1},
	}

	tests := map[string]struct {
		vars         RequestVars
		ensemblerSvc func() service.EnsemblersService
		expected     *Response
	}{
		"failure | bad request": {
			vars: RequestVars{},
			expected: BadRequest(
				"unable to list ensemblers",
				"failed to parse query string: Key: 'EnsemblersListOptions.ProjectID' "+
					"Error:Field validation for 'ProjectID' failed on the 'required' tag"),
		},
		"failure | internal server error": {
			vars: RequestVars{"project_id": {"2"}},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.
					On("List", service.EnsemblersListOptions{
						ProjectID: models.NewID(2),
					}).
					Return(nil, errors.New("test ensemblers_service error"))
				return ensemblersSvc
			},
			expected: InternalServerError(
				"unable to list ensemblers",
				"test ensemblers_service error"),
		},
		"failure | invalid pagination parameters [page]": {
			vars: RequestVars{
				"project_id": {"3"},
				"page":       {"first"},
				"page_size":  {"10"},
			},
			expected: BadRequest(
				"unable to list ensemblers",
				`failed to parse query string: schema: error converting value for "page"`),
		},
		"failure | invalid pagination parameters [page_size] ": {
			vars: RequestVars{
				"project_id": {"3"},
				"page":       {"1"},
				"page_size":  {"0"},
			},
			expected: BadRequest(
				"unable to list ensemblers",
				"failed to parse query string: Key: 'EnsemblersListOptions.PaginationOptions.PageSize' "+
					"Error:Field validation for 'PageSize' failed on the 'min' tag"),
		},
		"success | no ensemblers found": {
			vars: RequestVars{"project_id": {"1"}},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.
					On("List", service.EnsemblersListOptions{ProjectID: models.NewID(1)}).
					Return(empty, nil)
				return ensemblersSvc
			},
			expected: Ok(empty),
		},
		"success": {
			vars: RequestVars{
				"project_id": {"3"},
				"page":       {"1"},
				"page_size":  {"10"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.
					On("List", service.EnsemblersListOptions{
						PaginationOptions: service.PaginationOptions{
							Page:     ref.Int(1),
							PageSize: ref.Int(10),
						},
						ProjectID: models.NewID(3),
					}).
					Return(ensemblers, nil)
				return ensemblersSvc
			},
			expected: Ok(ensemblers),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var ensemblersSvc service.EnsemblersService
			if tt.ensemblerSvc != nil {
				ensemblersSvc = tt.ensemblerSvc()
			}
			validator, _ := validation.NewValidator(nil)
			ctrl := &EnsemblersController{
				NewBaseController(
					&AppContext{
						EnsemblersService: ensemblersSvc,
					}, validator,
				),
			}
			response := ctrl.ListEnsemblers(nil, tt.vars, nil)
			assert.Equal(t, tt.expected, response)
		})
	}
}

func TestEnsemblersController_GetEnsembler(t *testing.T) {
	ensembler := &models.PyFuncEnsembler{
		GenericEnsembler: &models.GenericEnsembler{
			Model:     models.Model{ID: 2},
			Type:      models.EnsemblerPyFuncType,
			ProjectID: 1,
		},
	}

	tests := map[string]struct {
		vars         RequestVars
		ensemblerSvc func() service.EnsemblersService
		expected     *Response
	}{
		"failure | bad request": {
			vars: RequestVars{"project_id": {"1"}},
			expected: BadRequest(
				"failed to fetch ensembler",
				"failed to parse query string: Key: 'EnsemblersPathOptions.EnsemblerID' "+
					"Error:Field validation for 'EnsemblerID' failed on the 'required' tag",
			),
		},
		"failure | not found": {
			vars: RequestVars{
				"project_id":   {"1"},
				"ensembler_id": {"1"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(1), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(1),
					}).
					Return(nil, errors.New("test ensembler error"))
				return ensemblerSvc
			},
			expected: NotFound("ensembler not found", "test ensembler error"),
		},
		"failure | not found in project": {
			vars: RequestVars{
				"project_id":   {"2"},
				"ensembler_id": {"2"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(2),
					}).
					Return(nil, errors.New("ensembler with ID 2 doesn't belong to this project"))
				return ensemblerSvc
			},
			expected: NotFound(
				"ensembler not found",
				"ensembler with ID 2 doesn't belong to this project",
			),
		},
		"success": {
			vars: RequestVars{
				"project_id":   {"1"},
				"ensembler_id": {"2"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(1),
					}).
					Return(ensembler, nil)
				return ensemblerSvc
			},
			expected: Ok(ensembler),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validator, _ := validation.NewValidator(nil)
			var ensemblerSvc service.EnsemblersService
			if tt.ensemblerSvc != nil {
				ensemblerSvc = tt.ensemblerSvc()
			}
			ctrl := &EnsemblersController{
				NewBaseController(
					&AppContext{
						EnsemblersService: ensemblerSvc,
					},
					validator,
				),
			}
			response := ctrl.GetEnsembler(nil, tt.vars, nil)
			assert.Equal(t, tt.expected, response)
		})
	}
}

func TestEnsemblersController_UpdateEnsembler(t *testing.T) {
	original := &models.PyFuncEnsembler{
		GenericEnsembler: &models.GenericEnsembler{
			Model:     models.Model{ID: 2},
			ProjectID: 2,
			Type:      models.EnsemblerPyFuncType,
			Name:      "original-ensembler",
		},
		MlflowURL:    "http://localhost:5000/experiemnts/0/runs/1",
		ExperimentID: 0,
		RunID:        "1",
		ArtifactURI:  "gs://bucket-name/mlflow/0/1/artifacts",
	}

	updated := &models.PyFuncEnsembler{
		GenericEnsembler: &models.GenericEnsembler{
			Model:     models.Model{ID: 2},
			ProjectID: 2,
			Type:      models.EnsemblerPyFuncType,
			Name:      "updated-ensembler",
		},
		MlflowURL:    "http://localhost:5000/experiemnts/0/runs/2",
		ExperimentID: 0,
		RunID:        "2",
		ArtifactURI:  "gs://bucket-name/mlflow/0/2/artifacts",
	}

	tests := map[string]struct {
		vars         RequestVars
		ensemblerSvc func() service.EnsemblersService
		mlflowSvc    func() mlflow.Service
		body         interface{}
		expected     *Response
	}{
		"failure | bad request": {
			vars: RequestVars{"project_id": {"unknown"}},
			expected: BadRequest(
				"failed to fetch ensembler",
				`failed to parse query string: schema: error converting value for "project_id"`,
			),
		},
		"failure | ensembler not found": {
			vars: RequestVars{
				"project_id":   {"1"},
				"ensembler_id": {"2"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(1),
					}).
					Return(nil, errors.New("ensembler with ID 2 doesn't belong to this project"))
				return ensemblerSvc
			},
			expected: NotFound(
				"ensembler not found",
				"ensembler with ID 2 doesn't belong to this project",
			),
		},
		"failure | invalid payload": {
			vars: RequestVars{
				"project_id":   {"2"},
				"ensembler_id": {"2"},
			},
			body: &request.CreateOrUpdateEnsemblerRequest{
				EnsemblerLike: &models.GenericEnsembler{
					Model:     models.Model{},
					ProjectID: 2,
					Type:      "unknown",
					Name:      "updated-ensembler",
				},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(2),
					}).
					Return(original, nil)
				return ensemblerSvc
			},
			expected: BadRequest(
				"invalid ensembler configuration",
				"Ensembler type cannot be changed after creation",
			),
		},
		"failure | incompatible types": {
			vars: RequestVars{
				"project_id":   {"2"},
				"ensembler_id": {"2"},
			},
			body: &request.CreateOrUpdateEnsemblerRequest{
				EnsemblerLike: &models.GenericEnsembler{
					Model:     models.Model{},
					ProjectID: 2,
					Type:      "pyfunc",
					Name:      "updated-ensembler",
				},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(2),
					}).
					Return(original, nil)
				return ensemblerSvc
			},
			expected: BadRequest(
				"invalid ensembler configuration",
				"update must be of the same type as as the receiver",
			),
		},
		"failure | failed to save": {
			vars: RequestVars{
				"project_id":   {"2"},
				"ensembler_id": {"2"},
			},
			body: &request.CreateOrUpdateEnsemblerRequest{
				EnsemblerLike: &models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model:     models.Model{},
						ProjectID: 2,
						Name:      updated.Name,
					},
					MlflowURL:    "http://localhost:5000/experiemnts/0/runs/2",
					ExperimentID: 0,
					RunID:        "2",
					ArtifactURI:  "gs://bucket-name/mlflow/0/2/artifacts",
				},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(2),
					}).
					Return(original, nil)
				ensemblerSvc.
					On("Save", updated).
					Return(nil, errors.New("failed to save"))
				return ensemblerSvc
			},
			mlflowSvc: func() mlflow.Service {
				mlflowSvc := &mlflowMock.Service{}
				mlflowSvc.On("DeleteRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				return mlflowSvc
			},
			expected: InternalServerError(
				"failed to update an ensembler", "failed to save"),
		},
		"success": {
			vars: RequestVars{
				"project_id":   {"2"},
				"ensembler_id": {"2"},
			},
			body: &request.CreateOrUpdateEnsemblerRequest{
				EnsemblerLike: &models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model:     models.Model{},
						ProjectID: 2,
						Name:      updated.Name,
					},
					MlflowURL:    "http://localhost:5000/experiemnts/0/runs/2",
					ExperimentID: 0,
					RunID:        "2",
					ArtifactURI:  "gs://bucket-name/mlflow/0/2/artifacts",
				},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(2),
					}).
					Return(original, nil)
				ensemblerSvc.
					On("Save", updated).
					Return(updated, nil)
				return ensemblerSvc
			},
			mlflowSvc: func() mlflow.Service {
				mlflowSvc := &mlflowMock.Service{}
				mlflowSvc.On("DeleteRun", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				return mlflowSvc
			},
			expected: Ok(updated),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validator, _ := validation.NewValidator(nil)
			var ensemblerSvc service.EnsemblersService
			if tt.ensemblerSvc != nil {
				ensemblerSvc = tt.ensemblerSvc()
			}
			var mlflowSvc mlflow.Service
			if tt.mlflowSvc != nil {
				mlflowSvc = tt.mlflowSvc()
			}

			ctrl := &EnsemblersController{
				NewBaseController(
					&AppContext{
						EnsemblersService: ensemblerSvc,
						MlflowService:     mlflowSvc,
					},
					validator,
				),
			}
			response := ctrl.UpdateEnsembler(nil, tt.vars, tt.body)
			assert.Equal(t, tt.expected, response)
		})
	}
}

func TestEnsemblerController_DeleteEnsembler(t *testing.T) {
	original := &models.PyFuncEnsembler{
		GenericEnsembler: &models.GenericEnsembler{
			Model:     models.Model{ID: 2},
			ProjectID: 2,
			Type:      models.EnsemblerPyFuncType,
			Name:      "original-ensembler",
		},
		MlflowURL:    "http://localhost:5000/experiemnts/0/runs/1",
		ExperimentID: 1,
		RunID:        "1",
		ArtifactURI:  "gs://bucket-name/mlflow/0/1/artifacts",
	}
	routerVersionStatusInactive := []models.RouterVersionStatus{
		models.RouterVersionStatusFailed,
		models.RouterVersionStatusUndeployed,
	}

	routerVersionStatusActive := []models.RouterVersionStatus{
		models.RouterVersionStatusDeployed,
		models.RouterVersionStatusPending,
	}

	routerVersion := &models.RouterVersion{
		Model: models.Model{
			ID: 1,
		},
		Status: "deployed",
	}

	ensemblerID := models.ID(2)

	ensemblingJobActiveOption := service.EnsemblingJobListOptions{
		EnsemblerID: &ensemblerID,
		Statuses: []models.Status{
			models.JobPending,
			models.JobBuildingImage,
			models.JobRunning,
		},
	}

	ensemblingJobInactiveOption := service.EnsemblingJobListOptions{
		EnsemblerID: &ensemblerID,
		Statuses: []models.Status{
			models.JobFailed,
			models.JobCompleted,
			models.JobFailedBuildImage,
			models.JobFailedSubmission,
		},
	}

	dummyEnsemblingJob := GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "", true)

	tests := map[string]struct {
		vars              RequestVars
		ensemblerSvc      func() service.EnsemblersService
		mlflowSvc         func() mlflow.Service
		routerVersionsSvc func() service.RouterVersionsService
		ensemblingJobSvc  func() service.EnsemblingJobService
		expected          *Response
	}{
		"failure | bad request": {
			vars: RequestVars{"project_id": {"unknown"}},
			expected: BadRequest(
				"failed to fetch ensembler",
				`failed to parse query string: schema: error converting value for "project_id"`,
			),
		},
		"failure | ensembler not found": {
			vars: RequestVars{
				"project_id":   {"1"},
				"ensembler_id": {"2"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(1),
					}).
					Return(nil, errors.New("ensembler with ID 2 doesn't belong to this project"))
				return ensemblerSvc
			},
			expected: NotFound(
				"ensembler not found",
				"ensembler with ID 2 doesn't belong to this project",
			),
		},
		"failure | there is active router version": {
			vars: RequestVars{
				"project_id":   {"2"},
				"ensembler_id": {"2"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(2),
					}).
					Return(original, nil)
				ensemblerSvc.
					On("Delete", original).
					Return(nil)
				return ensemblerSvc
			},
			routerVersionsSvc: func() service.RouterVersionsService {
				routerVersionSvc := &mocks.RouterVersionsService{}
				routerVersionSvc.On("FindRouterVersionsByEnsembler", mock.Anything).
					Return([]*models.RouterVersion{routerVersion}, nil)

				return routerVersionSvc
			},
			ensemblingJobSvc: func() service.EnsemblingJobService {
				ensemblingJobSvc := &mocks.EnsemblingJobService{}
				ensemblingJobSvc.On("List", mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{},
						Paging: service.Paging{
							Total: 0,
							Page:  1,
							Pages: 1,
						},
					},
					nil)
				ensemblingJobSvc.On("Delete", mock.Anything).Return(nil)
				return ensemblingJobSvc
			},
			expected: BadRequest("failed to delete an ensembler", "There are active router version using this ensembler"),
		},
		"failure | there is active ensembling job": {
			vars: RequestVars{
				"project_id":   {"2"},
				"ensembler_id": {"2"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(2),
					}).
					Return(original, nil)
				ensemblerSvc.
					On("Delete", original).
					Return(nil)
				return ensemblerSvc
			},
			routerVersionsSvc: func() service.RouterVersionsService {
				routerVersionSvc := &mocks.RouterVersionsService{}
				routerVersionSvc.On("FindRouterVersionsByEnsembler", mock.Anything).
					Return([]*models.RouterVersion{}, nil)

				return routerVersionSvc
			},
			ensemblingJobSvc: func() service.EnsemblingJobService {
				ensemblingJobSvc := &mocks.EnsemblingJobService{}
				ensemblingJobSvc.On("List", mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{
							dummyEnsemblingJob,
						},
						Paging: service.Paging{
							Total: 1,
							Page:  1,
							Pages: 1,
						},
					},
					nil)
				ensemblingJobSvc.On("Delete", mock.Anything).Return(nil)
				return ensemblingJobSvc
			},
			expected: BadRequest("failed to delete an ensembler", "there are active ensembling job using this ensembler"),
		},
		"failure | failed to delete router version": {
			vars: RequestVars{
				"project_id":   {"2"},
				"ensembler_id": {"2"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(2),
					}).
					Return(original, nil)
				ensemblerSvc.
					On("Delete", original).
					Return(nil)
				return ensemblerSvc
			},
			routerVersionsSvc: func() service.RouterVersionsService {
				routerVersionSvc := &mocks.RouterVersionsService{}
				routerVersionSvc.On("FindRouterVersionsByEnsembler", service.RouterVersionByEnsemblerListOptions{
					ProjectID:   models.NewID(2),
					EnsemblerID: models.NewID(2),
					Statuses:    routerVersionStatusActive,
				}).
					Return([]*models.RouterVersion{}, nil)
				routerVersionSvc.On("FindRouterVersionsByEnsembler", service.RouterVersionByEnsemblerListOptions{
					ProjectID:   models.NewID(2),
					EnsemblerID: models.NewID(2),
					Statuses:    routerVersionStatusInactive,
				}).
					Return([]*models.RouterVersion{routerVersion}, nil)
				routerVersionSvc.On("Delete", mock.Anything).Return(errors.New("failed to delete router version"))
				return routerVersionSvc
			},
			ensemblingJobSvc: func() service.EnsemblingJobService {
				ensemblingJobSvc := &mocks.EnsemblingJobService{}
				ensemblingJobSvc.On("List", mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{},
						Paging: service.Paging{
							Total: 0,
							Page:  1,
							Pages: 1,
						},
					},
					nil)
				ensemblingJobSvc.On("Delete", mock.Anything).Return(nil)
				return ensemblingJobSvc
			},
			expected: InternalServerError("unable to delete router version", "failed to delete router version"),
		},
		"failure | failed to delete ensembling job": {
			vars: RequestVars{
				"project_id":   {"2"},
				"ensembler_id": {"2"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(2),
					}).
					Return(original, nil)
				ensemblerSvc.
					On("Delete", original).
					Return(nil)
				return ensemblerSvc
			},
			routerVersionsSvc: func() service.RouterVersionsService {
				routerVersionSvc := &mocks.RouterVersionsService{}
				routerVersionSvc.On("FindRouterVersionsByEnsembler", mock.Anything).Return([]*models.RouterVersion{}, nil)

				return routerVersionSvc
			},
			ensemblingJobSvc: func() service.EnsemblingJobService {
				ensemblingJobSvc := &mocks.EnsemblingJobService{}
				ensemblingJobSvc.On("List", ensemblingJobActiveOption).Return(
					&service.PaginatedResults{
						Results: []interface{}{},
						Paging: service.Paging{
							Total: 0,
							Page:  1,
							Pages: 1,
						},
					},
					nil)
				ensemblingJobSvc.On("List", ensemblingJobInactiveOption).Return(
					&service.PaginatedResults{
						Results: []*models.EnsemblingJob{
							dummyEnsemblingJob,
						},
						Paging: service.Paging{
							Total: 1,
							Page:  1,
							Pages: 1,
						},
					},
					nil)
				ensemblingJobSvc.On("Delete", dummyEnsemblingJob).Return(errors.New("failed to delete ensembling job"))
				ensemblingJobSvc.On("MarkEnsemblingJobForTermination", dummyEnsemblingJob).
					Return(errors.New("failed to delete ensembling job"))
				return ensemblingJobSvc
			},
			mlflowSvc: func() mlflow.Service {
				mlflowSvc := &mlflowMock.Service{}
				mlflowSvc.On("DeleteExperiment", mock.Anything, "1", true).Return(nil)
				return mlflowSvc
			},
			expected: InternalServerError("unable to delete ensembling job", "failed to delete ensembling job"),
		},
		"failure | failed to delete mlflow experiment": {
			vars: RequestVars{
				"project_id":   {"2"},
				"ensembler_id": {"2"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(2),
					}).
					Return(original, nil)
				ensemblerSvc.
					On("Delete", original).
					Return(nil)
				return ensemblerSvc
			},
			routerVersionsSvc: func() service.RouterVersionsService {
				routerVersionSvc := &mocks.RouterVersionsService{}
				routerVersionSvc.On("FindRouterVersionsByEnsembler", mock.Anything).Return([]*models.RouterVersion{}, nil)

				return routerVersionSvc
			},
			ensemblingJobSvc: func() service.EnsemblingJobService {
				ensemblingJobSvc := &mocks.EnsemblingJobService{}
				ensemblingJobSvc.On("List", mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{},
						Paging: service.Paging{
							Total: 0,
							Page:  1,
							Pages: 1,
						},
					},
					nil)
				ensemblingJobSvc.On("Delete", mock.Anything).Return(nil)
				return ensemblingJobSvc
			},
			mlflowSvc: func() mlflow.Service {
				mlflowSvc := &mlflowMock.Service{}
				mlflowSvc.On("DeleteExperiment", mock.Anything, "1", true).Return(errors.New("failed to delete mlflow experiment"))
				return mlflowSvc
			},
			expected: InternalServerError("failed to delete an ensembler", "failed to delete mlflow experiment"),
		},
		"failure | failed to delete": {
			vars: RequestVars{
				"project_id":   {"2"},
				"ensembler_id": {"2"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(2),
					}).
					Return(original, nil)
				ensemblerSvc.
					On("Delete", original).
					Return(errors.New("failed to delete"))
				return ensemblerSvc
			},
			routerVersionsSvc: func() service.RouterVersionsService {
				routerVersionSvc := &mocks.RouterVersionsService{}
				routerVersionSvc.On("FindRouterVersionsByEnsembler", mock.Anything).Return([]*models.RouterVersion{}, nil)

				return routerVersionSvc
			},
			ensemblingJobSvc: func() service.EnsemblingJobService {
				ensemblingJobSvc := &mocks.EnsemblingJobService{}
				ensemblingJobSvc.On("List", mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{},
						Paging: service.Paging{
							Total: 0,
							Page:  1,
							Pages: 1,
						},
					},
					nil)
				ensemblingJobSvc.On("Delete", mock.Anything).Return(nil)
				return ensemblingJobSvc
			},
			mlflowSvc: func() mlflow.Service {
				mlflowSvc := &mlflowMock.Service{}
				mlflowSvc.On("DeleteExperiment", mock.Anything, "1", true).Return(nil)
				return mlflowSvc
			},
			expected: InternalServerError("failed to delete an ensembler", "failed to delete"),
		},
		"success": {
			vars: RequestVars{
				"project_id":   {"2"},
				"ensembler_id": {"2"},
			},
			ensemblerSvc: func() service.EnsemblersService {
				ensemblerSvc := &mocks.EnsemblersService{}
				ensemblerSvc.
					On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
						ProjectID: models.NewID(2),
					}).
					Return(original, nil)
				ensemblerSvc.
					On("Delete", original).
					Return(nil)
				return ensemblerSvc
			},
			routerVersionsSvc: func() service.RouterVersionsService {
				routerVersionSvc := &mocks.RouterVersionsService{}
				routerVersionSvc.On("FindRouterVersionsByEnsembler", mock.Anything).Return([]*models.RouterVersion{}, nil)

				return routerVersionSvc
			},
			ensemblingJobSvc: func() service.EnsemblingJobService {
				ensemblingJobSvc := &mocks.EnsemblingJobService{}
				ensemblingJobSvc.On("List", mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{},
						Paging: service.Paging{
							Total: 0,
							Page:  1,
							Pages: 1,
						},
					},
					nil)
				ensemblingJobSvc.On("Delete", mock.Anything).Return(nil)
				return ensemblingJobSvc
			},
			mlflowSvc: func() mlflow.Service {
				mlflowSvc := &mlflowMock.Service{}
				mlflowSvc.On("DeleteExperiment", mock.Anything, "1", true).Return(nil)
				return mlflowSvc
			},
			expected: Ok(models.ID(2)),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validator, _ := validation.NewValidator(nil)
			var ensemblerSvc service.EnsemblersService
			if tt.ensemblerSvc != nil {
				ensemblerSvc = tt.ensemblerSvc()
			}
			var mlflowSvc mlflow.Service
			if tt.mlflowSvc != nil {
				mlflowSvc = tt.mlflowSvc()
			}
			var ensemblingJobSvc service.EnsemblingJobService
			if tt.ensemblingJobSvc != nil {
				ensemblingJobSvc = tt.ensemblingJobSvc()
			}
			var routerVersionsSvc service.RouterVersionsService
			if tt.ensemblingJobSvc != nil {
				routerVersionsSvc = tt.routerVersionsSvc()
			}

			ctrl := &EnsemblersController{
				NewBaseController(
					&AppContext{
						EnsemblersService:     ensemblerSvc,
						MlflowService:         mlflowSvc,
						EnsemblingJobService:  ensemblingJobSvc,
						RouterVersionsService: routerVersionsSvc,
					},
					validator,
				),
			}
			response := ctrl.DeleteEnsembler(nil, tt.vars, nil)
			assert.Equal(t, tt.expected, response)
		})
	}
}
