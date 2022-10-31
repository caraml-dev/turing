//go:build integration

package api_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"

	"github.com/caraml-dev/turing/api/turing/api"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/server"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"
	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/api/turing/service/mocks"
)

func generateEnsemblingJobFixtureJSON() string {
	return `{
		"name":"test-ensembler-1",
		"ensembler_id":1,
		"infra_config":{
			"service_account_name":"test-service-account-1",
			"resources":{
				"driver_cpu_request": "1",
				"driver_memory_request": "1Gi",
				"executor_replica": 10,
				"executor_cpu_request": "1",
				"executor_memory_request": "1Gi"
			},
			"environment_variables": {
				"foo": "bar"
			}
		},
		"job_config":{
			"spec":{
				"sink":{
					"type":"BQ",
					"columns":[
						"customer_id as customerId",
						"target_date",
						"results"
					],
					"bq_config":{
						"table":"project.dataset.ensembling_results",
						"options":{
							"partitionField":"target_date"
						},
						"staging_bucket":"bucket-name"
					},
					"save_mode":"OVERWRITE"
				},
				"source":{
					"join_on":[
						"customer_id",
						"target_date"
					],
					"dataset":{
						"bq_config":{
							"query":"select * from helloworld where customer_id = 4",
							"options":{
								"viewsEnabled":"true",
								"materializationDataset":"dataset"
							}
						}
					}
				},
				"ensembler":{
					"result":{
						"type":"FLOAT",
						"item_type":"FLOAT",
						"column_name":"prediction_score"
					}
				},
				"predictions":{
					"model_a":{
						"join_on":[
							"customer_id",
							"target_date"
						],
						"columns":[
							"predictions"
						],
						"dataset":{
							"bq_config":{
								"table":"project.dataset.predictions_model_a",
								"features":[
									"customer_id",
									"target_date",
									"predictions"
								]
							}
						}
					},
					"model_b":{
						"join_on":[
							"customer_id",
							"target_date"
						],
						"columns":[
							"predictions"
						],
						"dataset":{
							"bq_config":{
								"query":"select * from helloworld where customer_id = 3"
							}
						}
					}
				}
			},
			"version":"v1",
			"metadata":{
				"name":"test-batch-ensembling-1",
				"annotations":{
					"spark/spark.jars":"https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.0.1.jar",
					"spark/spark.jars.packages":"com.google.cloud.spark:spark-bigquery-with-dependencies_2.12:0.19.1",
					"hadoopConfiguration/fs.gs.impl":"com.google.cloud.hadoop.fs.gcs.GoogleHadoopFileSystem",
					"hadoopConfiguration/fs.AbstractFileSystem.gs.impl":"com.google.cloud.hadoop.fs.gcs.GoogleHadoopFS"
				}
			}
		}
	}`
}

func TestIntegrationEnsemblingJobController_CreateEnsemblingJob(t *testing.T) {
	var tests = map[string]struct {
		method               string
		path                 string
		expected             *api.Response
		ensemblersService    func() service.EnsemblersService
		ensemblingJobService func() service.EnsemblingJobService
		mlpService           func() service.MLPService
		vars                 api.RequestVars
		body                 string
	}{
		"success | nominal flow": {
			path:     "/projects/1/jobs",
			method:   http.MethodPost,
			expected: api.Accepted(api.GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true)),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(api.CreateEnsembler(1, "pyfunc"), nil)
				return ensemblersSvc
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				ensemblingJobService := &mocks.EnsemblingJobService{}
				ensemblingJobService.On(
					"CreateEnsemblingJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(api.GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true), nil)

				return ensemblingJobService
			},
			mlpService: func() service.MLPService {
				mlpService := &mocks.MLPService{}
				mlpService.On(
					"GetProject",
					models.ID(1),
				).Return(&mlp.Project{Id: 1, Name: "foo"}, nil)
				return mlpService
			},
			vars: api.RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixtureJSON(),
		},
		"failure | non existent ensembler": {
			path:     "/projects/1/jobs",
			method:   http.MethodPost,
			expected: api.NotFound("ensembler not found", errors.New("no exist").Error()),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("no exist"))
				return ensemblersSvc
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				ensemblingJobService := &mocks.EnsemblingJobService{}
				return ensemblingJobService
			},
			mlpService: func() service.MLPService {
				mlpService := &mocks.MLPService{}
				mlpService.On(
					"GetProject",
					models.ID(1),
				).Return(&mlp.Project{Id: 1, Name: "foo"}, nil)
				return mlpService
			},
			vars: api.RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixtureJSON(),
		},
		"failure | wrong type of ensembler": {
			path:     "/projects/1/jobs",
			method:   http.MethodPost,
			expected: api.BadRequest("only pyfunc ensemblers allowed", "ensembler type given: *models.GenericEnsembler"),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(api.CreateEnsembler(1, "generic"), nil)
				return ensemblersSvc
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				ensemblingJobService := &mocks.EnsemblingJobService{}
				return ensemblingJobService
			},
			mlpService: func() service.MLPService {
				mlpService := &mocks.MLPService{}
				mlpService.On(
					"GetEnvironment",
					"dev",
				).Return(&merlin.Environment{}, nil)
				mlpService.On(
					"GetProject",
					models.ID(1),
				).Return(&mlp.Project{Id: 1, Name: "foo"}, nil)
				return mlpService
			},
			vars: api.RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixtureJSON(),
		},
		"failure | non existent project": {
			path:     "/projects/1/jobs",
			method:   http.MethodPost,
			expected: api.NotFound("project not found", errors.New("hello").Error()),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(api.CreateEnsembler(1, "pyfunc"), nil)
				return ensemblersSvc
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				ensemblingJobService := &mocks.EnsemblingJobService{}
				return ensemblingJobService
			},
			mlpService: func() service.MLPService {
				mlpService := &mocks.MLPService{}
				mlpService.On(
					"GetProject",
					models.ID(1),
				).Return(nil, errors.New("hello"))
				return mlpService
			},
			vars: api.RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixtureJSON(),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ensemblersService := tt.ensemblersService()
			ensemblingJobService := tt.ensemblingJobService()
			mlpService := tt.mlpService()

			router := mux.NewRouter()
			appCtx := &api.AppContext{
				EnsemblersService:    ensemblersService,
				EnsemblingJobService: ensemblingJobService,
				MLPService:           mlpService,
			}
			_ = server.AddAPIRoutesHandler(
				router,
				"/",
				appCtx,
				&config.Config{
					BatchEnsemblingConfig: config.BatchEnsemblingConfig{
						Enabled: true,
					},
				},
			)

			actual := httptest.NewRecorder()

			request, err := http.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatalf("unexpected error happened, %v", err)
			}
			router.ServeHTTP(actual, request)

			expected := httptest.NewRecorder()
			tt.expected.WriteTo(expected)

			assert.Equal(t, expected.Code, actual.Code)
			assert.Equal(t, expected.Body.String(), actual.Body.String())
		})
	}
}

func TestIntegrationEnsemblingJobController_GetEnsemblingJob(t *testing.T) {
	tests := map[string]struct {
		method               string
		path                 string
		expected             *api.Response
		ensemblingJobService func() service.EnsemblingJobService
	}{
		"success | nominal": {
			method: http.MethodGet,
			path:   "/projects/1/jobs/1",
			expected: api.Ok(api.GenerateEnsemblingJobFixture(
				1,
				models.ID(1),
				models.ID(1),
				"test-ensembler-1",
				true,
			)),
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("FindByID", mock.Anything, mock.Anything).Return(
					api.GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
					nil,
				)
				return svc
			},
		},
		"failure | not found": {
			method:   http.MethodGet,
			path:     "/projects/1/jobs/1",
			expected: api.NotFound("ensembling job not found", errors.New("no exist").Error()),
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("FindByID", mock.Anything, mock.Anything).Return(
					nil,
					errors.New("no exist"),
				)
				return svc
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			svc := tt.ensemblingJobService()

			router := mux.NewRouter()
			appCtx := &api.AppContext{
				EnsemblingJobService: svc,
			}
			_ = server.AddAPIRoutesHandler(
				router,
				"/",
				appCtx,
				&config.Config{
					BatchEnsemblingConfig: config.BatchEnsemblingConfig{
						Enabled: true,
					},
				},
			)

			actual := httptest.NewRecorder()
			request, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatalf("unexpected error happened, %v", err)
			}
			router.ServeHTTP(actual, request)

			expected := httptest.NewRecorder()
			tt.expected.WriteTo(expected)

			assert.Equal(t, expected.Body.String(), actual.Body.String())
			assert.Equal(t, expected.Code, actual.Code)
		})
	}
}

func TestIntegrationEnsemblingJobController_ListEnsemblingJob(t *testing.T) {
	tests := map[string]struct {
		method               string
		path                 string
		expected             *api.Response
		ensemblingJobService func() service.EnsemblingJobService
	}{
		"success | nominal": {
			method: http.MethodGet,
			path:   "/projects/1/jobs",
			expected: api.Ok(
				&service.PaginatedResults{
					Results: []interface{}{
						api.GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
					},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				},
			),
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("List", mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{
							api.GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
						},
						Paging: service.Paging{
							Total: 1,
							Page:  1,
							Pages: 1,
						},
					},
					nil,
				)
				return svc
			},
		},
		"success | nominal with single status": {
			method: http.MethodGet,
			path:   "/projects/1/jobs?status=pending",
			expected: api.Ok(
				&service.PaginatedResults{
					Results: []interface{}{
						api.GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
					},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				},
			),
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("List", mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{
							api.GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
						},
						Paging: service.Paging{
							Total: 1,
							Page:  1,
							Pages: 1,
						},
					},
					nil,
				)
				return svc
			},
		},
		"success | nominal with multiple statuses": {
			method: http.MethodGet,
			path:   "/projects/1/jobs?status=pending&status=terminated",
			expected: api.Ok(
				&service.PaginatedResults{
					Results: []interface{}{
						api.GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
					},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				},
			),
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("List", mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{
							api.GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
						},
						Paging: service.Paging{
							Total: 1,
							Page:  1,
							Pages: 1,
						},
					},
					nil,
				)
				return svc
			},
		},
		"success | no result": {
			method: http.MethodGet,
			path:   "/projects/1/jobs",
			expected: api.Ok(
				&service.PaginatedResults{
					Results: []interface{}{},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				},
			),
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("List", mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{},
						Paging: service.Paging{
							Total: 1,
							Page:  1,
							Pages: 1,
						},
					},
					nil,
				)
				return svc
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			svc := tt.ensemblingJobService()

			router := mux.NewRouter()
			appCtx := &api.AppContext{
				EnsemblingJobService: svc,
			}
			_ = server.AddAPIRoutesHandler(
				router,
				"/",
				appCtx,
				&config.Config{
					BatchEnsemblingConfig: config.BatchEnsemblingConfig{
						Enabled: true,
					},
				},
			)

			actual := httptest.NewRecorder()

			request, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatalf("unexpected error happened, %v", err)
			}
			router.ServeHTTP(actual, request)

			expected := httptest.NewRecorder()
			tt.expected.WriteTo(expected)

			assert.Equal(t, expected.Body.String(), actual.Body.String())
			assert.Equal(t, expected.Code, actual.Code)
		})
	}
}
