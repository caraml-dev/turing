//go:build integration

package api_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mlp "github.com/gojek/mlp/api/client"
	"github.com/gorilla/mux"

	"github.com/caraml-dev/turing/api/turing/api"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/server"
	"github.com/caraml-dev/turing/api/turing/service/mocks"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/api/turing/service"
)

func TestEnsemblersController_CreateEnsembler(t *testing.T) {
	mlpSvc := &mocks.MLPService{}
	mlpSvc.
		On("GetProject", models.ID(1)).
		Return(nil, errors.New("error"))
	mlpSvc.
		On("GetProject", models.ID(2)).
		Return(&mlp.Project{Id: 2}, nil)

	tests := map[string]struct {
		method       string
		path         string
		body         string
		parsed       models.EnsemblerLike
		saved        models.EnsemblerLike
		ensemblerSvc func(models.EnsemblerLike, models.EnsemblerLike) service.EnsemblersService
		expected     func(models.EnsemblerLike) *api.Response
	}{
		"failure | project doesn't exist": {
			method: http.MethodPost,
			path:   "/projects/1/ensemblers",
			body: `{
 				"type": "pyfunc",
				"name": "my-ensembler-1"
			}`,
			expected: func(_ models.EnsemblerLike) *api.Response {
				return api.NotFound("project not found", "error")
			},
		},

		"failure | invalid payload": {
			method: http.MethodPost,
			path:   "/projects/1/ensemblers",
			body:   "string",
			expected: func(_ models.EnsemblerLike) *api.Response {
				return api.BadRequest(
					"invalid request body",
					"Failed to deserialize request body: invalid character 's' looking for beginning of value")
			},
		},
		"failure | unsupported ensembler": {
			method: http.MethodPost,
			path:   "/projects/1/ensemblers",
			body: `{
 				"type": "unknown",
				"name": "ensembler-unrecognized"
			}`,
			expected: func(_ models.EnsemblerLike) *api.Response {
				return api.BadRequest(
					"invalid request body",
					"Failed to deserialize request body: unsupported ensembler type: unknown")
			},
		},
		"failure | payload failed validation": {
			method: http.MethodPost,
			path:   "/projects/1/ensemblers",
			body: `{
 				"type": "pyfunc",
				"artifact_uri": "gs://unknown"
			}`,
			expected: func(_ models.EnsemblerLike) *api.Response {
				return api.BadRequest(
					"invalid request body",
					"Key: 'CreateOrUpdateEnsemblerRequest.EnsemblerLike.GenericEnsembler.Name' Error:Field validation for 'Name' failed on the 'required' tag")
			},
		},
		"failure | unable to save": {
			method: http.MethodPost,
			path:   "/projects/2/ensemblers",
			body: `{
 				"type": "pyfunc",
				"name": "new-ensembler",
				"artifact_uri": "gs://unknown"
			}`,
			parsed: &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					ProjectID: 2,
					Type:      "pyfunc",
					Name:      "new-ensembler",
				},
				ArtifactURI: "gs://unknown",
			},
			ensemblerSvc: func(parsed, _ models.EnsemblerLike) service.EnsemblersService {
				mockSvc := &mocks.EnsemblersService{}
				mockSvc.
					On("Save", parsed).
					Return(nil, fmt.Errorf(`ensembler with name "%s" already exists`, parsed.GetName()))
				return mockSvc
			},
			expected: func(_ models.EnsemblerLike) *api.Response {
				return api.InternalServerError(
					"unable to save an ensembler",
					`ensembler with name "new-ensembler" already exists`)
			},
		},
		"success | save pyfunc ensembler": {
			method: http.MethodPost,
			path:   "/projects/2/ensemblers",
			body: `{
 				"type": "pyfunc",
				"name": "my-ensembler-1"
			}`,
			parsed: &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					ProjectID: 2,
					Type:      "pyfunc",
					Name:      "my-ensembler-1",
				},
			},
			saved: &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					Model:     models.Model{ID: 1},
					ProjectID: 2,
					Type:      "pyfunc",
					Name:      "my-ensembler-1",
				},
				ExperimentID: 1,
				RunID:        "abcd-efghij-klmn",
				ArtifactURI:  "gs://bucket-name/mlflow/1/abcd-efghij-klmn/artifacts/ensembler",
			},
			ensemblerSvc: func(parsed, saved models.EnsemblerLike) service.EnsemblersService {
				mockSvc := &mocks.EnsemblersService{}
				mockSvc.
					On("Save", parsed).
					Return(saved, nil)
				return mockSvc
			},
			expected: func(ensembler models.EnsemblerLike) *api.Response {
				return api.Created(ensembler)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var ensemblerSvc service.EnsemblersService
			if tt.ensemblerSvc != nil {
				ensemblerSvc = tt.ensemblerSvc(tt.parsed, tt.saved)
			}

			router := mux.NewRouter()

			appCtx := &api.AppContext{
				MLPService:        mlpSvc,
				EnsemblersService: ensemblerSvc,
			}
			_ = server.AddAPIRoutesHandler(router, "/", appCtx, &config.Config{})

			actual := httptest.NewRecorder()

			request, err := http.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatalf("unexpected error happened, %v", err)
			}
			router.ServeHTTP(actual, request)

			expected := httptest.NewRecorder()
			tt.expected(tt.saved).WriteTo(expected)

			assert.Equal(t, expected.Code, actual.Code)
			assert.Equal(t, expected.Body.String(), actual.Body.String())
		})
	}
}
