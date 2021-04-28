package api

import (
	"errors"
	"testing"

	"github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/gojek/turing/api/turing/validation"

	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
	"github.com/gojek/turing/api/turing/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestEnsemblersController_ListEnsemblers(t *testing.T) {
	empty := &service.PaginatedResults{
		Results: []*models.GenericEnsembler{},
		Paging:  service.Paging{Total: 0, Page: 1, Pages: 1},
	}
	ensemblers := &service.PaginatedResults{
		Results: []models.EnsemblerLike{
			&models.GenericEnsembler{
				Model:      models.Model{ID: 1},
				TProjectID: 3,
				TType:      models.EnsemblerTypePyFunc,
				TName:      "test-ensembler-1",
			},
			&models.GenericEnsembler{
				Model:      models.Model{ID: 2},
				TProjectID: 3,
				TType:      models.EnsemblerTypePyFunc,
				TName:      "test-ensembler-2",
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
							Page:     testutils.NullableInt(1),
							PageSize: testutils.NullableInt(10),
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
			Model:      models.Model{ID: 2},
			TType:      models.EnsemblerTypePyFunc,
			TProjectID: 1,
		},
	}

	tests := map[string]struct {
		vars         RequestVars
		ensemblerSvc func() service.EnsemblersService
		expected     *Response
	}{
		"failure | bad request": {
			vars: RequestVars{"project_id": {"1"}},
			ensemblerSvc: func() service.EnsemblersService {
				return nil
			},
			expected: BadRequest(
				"failed to fetch ensembler",
				"failed to parse query string: Key: 'EnsemblerID' "+
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
			ctrl := &EnsemblersController{
				NewBaseController(
					&AppContext{
						EnsemblersService: tt.ensemblerSvc(),
					},
					validator,
				),
			}
			response := ctrl.GetEnsembler(nil, tt.vars, nil)
			assert.Equal(t, tt.expected, response)
		})
	}
}
