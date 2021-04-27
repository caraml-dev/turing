// +build integration

package api

import (
	"bytes"
	"errors"
	"fmt"
	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojek/turing/api/turing/service"
	"github.com/stretchr/testify/assert"
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
		expected     func(models.EnsemblerLike) *Response
	}{
		"failure | project doesn't exist": {
			method: http.MethodPost,
			path:   "/projects/1/ensemblers",
			body: `{
 				"type": "pyfunc",
				"name": "my-ensembler-1"
			}`,
			expected: func(_ models.EnsemblerLike) *Response {
				return NotFound("project not found", "error")
			},
		},
		"failure | unsupported ensembler": {
			method: http.MethodPost,
			path:   "/projects/1/ensemblers",
			body: `{
 				"type": "unknown",
				"name": "ensembler-unrecognized"
			}`,
			expected: func(_ models.EnsemblerLike) *Response {
				return BadRequest(
					"invalid request body",
					"Failed to deserialize request body: unsupported ensembler type: unknown")
			},
		},
		"failure | invalid payload": {
			method: http.MethodPost,
			path:   "/projects/1/ensemblers",
			body: `{
 				"type": "pyfunc",
				"artifact_uri": "gs://unknown"
			}`,
			expected: func(_ models.EnsemblerLike) *Response {
				return BadRequest(
					"invalid request body",
					"Key: 'CreateOrUpdateEnsemblerRequest.Ensembler.GenericEnsembler.TName' Error:Field validation for 'TName' failed on the 'required' tag")
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
					TProjectID: 2,
					TType:      "pyfunc",
					TName:      "new-ensembler",
				},
				ArtifactURI: "gs://unknown",
			},
			ensemblerSvc: func(parsed, _ models.EnsemblerLike) service.EnsemblersService {
				mockSvc := &mocks.EnsemblersService{}
				mockSvc.
					On("Save", parsed).
					Return(nil, fmt.Errorf(`ensembler with name "%s" already exists`, parsed.Name()))
				return mockSvc
			},
			expected: func(_ models.EnsemblerLike) *Response {
				return InternalServerError(
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
					TProjectID: 2,
					TType:      "pyfunc",
					TName:      "my-ensembler-1",
				},
			},
			saved: &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					Model:      models.Model{ID: 1},
					TProjectID: 2,
					TType:      "pyfunc",
					TName:      "my-ensembler-1",
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
			expected: func(ensembler models.EnsemblerLike) *Response {
				return Created(ensembler)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var ensemblerSvc service.EnsemblersService
			if tt.ensemblerSvc != nil {
				ensemblerSvc = tt.ensemblerSvc(tt.parsed, tt.saved)
			}
			router := NewRouter(&AppContext{
				MLPService:        mlpSvc,
				EnsemblersService: ensemblerSvc,
			})

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
