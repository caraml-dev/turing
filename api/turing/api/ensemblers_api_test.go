package api

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
	"github.com/gojek/turing/api/turing/service/mocks"
	"github.com/stretchr/testify/assert"
)

func TestEnsemblersController_ListEnsemblers(t *testing.T) {
	mlpSvc := &mocks.MLPService{}
	mlpSvc.
		On("GetProject", models.ID(1)).
		Return(nil, errors.New("test project error"))
	mlpSvc.On("GetProject", models.ID(2)).Return(&mlp.Project{Id: 2}, nil)
	mlpSvc.On("GetProject", models.ID(3)).Return(&mlp.Project{Id: 3}, nil)

	ensemblers := &service.PaginatedResults{
		Results: []models.EnsemblerLike{
			&models.GenericEnsembler{
				Model:     models.Model{ID: 1},
				ProjectID: 3,
				Type:      models.EnsemblerTypePyFunc,
				Name:      "test-ensembler-1",
			},
			&models.GenericEnsembler{
				Model:     models.Model{ID: 2},
				ProjectID: 3,
				Type:      models.EnsemblerTypePyFunc,
				Name:      "test-ensembler-2",
			},
		},
		Paging: service.Paging{Total: 3, Page: 1, Pages: 1},
	}
	ensemblersSvc := &mocks.EnsemblersService{}
	ensemblersSvc.
		On("List", models.ID(2), service.ListEnsemblersQuery{}).
		Return(nil, errors.New("test ensemblers_service error"))
	ensemblersSvc.
		On("List", models.ID(3), service.NewListEnsemblersQuery(1, 10)).
		Return(ensemblers, nil)

	tests := map[string]struct {
		vars     map[string]string
		query    string
		expected *Response
	}{
		"failure | bad request": {
			vars:     map[string]string{},
			expected: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		"failure | not found": {
			vars:     map[string]string{"project_id": "1"},
			expected: NotFound("project not found", "test project error"),
		},
		"failure | internal server error": {
			vars:     map[string]string{"project_id": "2"},
			expected: InternalServerError("unable to list ensemblers", "test ensemblers_service error"),
		},
		"success": {
			vars:     map[string]string{"project_id": "3"},
			query:    "page=1&page_size=10",
			expected: Ok(ensemblers),
		},
	}

	// Run tests
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &EnsemblersController{
				NewBaseController(
					&AppContext{
						MLPService:        mlpSvc,
						EnsemblersService: ensemblersSvc,
					},
				),
			}
			response := ctrl.ListEnsemblers(&http.Request{URL: &url.URL{RawQuery: tt.query}}, tt.vars, nil)
			assert.Equal(t, tt.expected, response)
		})
	}
}
