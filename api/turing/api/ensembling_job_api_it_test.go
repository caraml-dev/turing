// +build integration

package api

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
	"github.com/gojek/turing/api/turing/service/mocks"
	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
)

func generateEnsemblingJobFixtureJSON() string {
	return `{
		"name":"test-ensembler-1",
		"version_id":1,
		"ensembler_id":1,
		"environment_name":"gods-dev",
		"infra_config":{
			"service_account_name":"test-service-account-1",
			"image_ref":"gcr.io/hello/world:123",
			"resources":{
				"requests":{
					"cpu":"2",
					"memory":"2Gi"
				},
				"limits":{
					"cpu":"2",
					"memory":"2Gi"
				}
			}
		},
		"ensembler_config":{
			"spec":{
				"sink":{
					"type":"BQ",
					"columns":[
						"customer_id as customerId",
						"target_date",
						"results"
					],
					"bqConfig":{
						"table":"project.dataset.ensembling_results",
						"options":{
							"partitionField":"target_date"
						},
						"stagingBucket":"bucket-name"
					},
					"saveMode":"OVERWRITE"
				},
				"source":{
					"joinOn":[
						"customer_id",
						"target_date"
					],
					"dataset":{
						"bqConfig":{
							"query":"select * from helloworld where customer_id = 4",
							"options":{
								"viewsEnabled":"true",
								"materializationDataset":"dataset"
							}
						}
					}
				},
				"ensembler":{
					"uri":"gs://bucket-name/my-ensembler/artifacts/ensembler",
					"result":{
						"type":"FLOAT",
						"itemType":"FLOAT",
						"columnName":"prediction_score"
					}
				},
				"predictions":{
					"model_a":{
						"joinOn":[
							"customer_id",
							"target_date"
						],
						"columns":[
							"predictions"
						],
						"dataset":{
							"bqConfig":{
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
						"joinOn":[
							"customer_id",
							"target_date"
						],
						"columns":[
							"predictions"
						],
						"dataset":{
							"bqConfig":{
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
	var tests = []struct {
		name                 string
		method               string
		path                 string
		expected             *Response
		ensemblersService    func() service.EnsemblersService
		ensemblingJobService func() service.EnsemblingJobService
		mlpService           func() service.MLPService
		vars                 RequestVars
		body                 string
	}{
		{
			name:     "nominal flow",
			path:     "/projects/1/jobs",
			method:   http.MethodPost,
			expected: Accepted(generateEnsemblingJobFixture(1, models.ID(1), models.ID(1))),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)
				return ensemblersSvc
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				ensemblingJobService := &mocks.EnsemblingJobService{}
				ensemblingJobService.On(
					"Save",
					mock.Anything,
				).Return(nil)
				return ensemblingJobService
			},
			mlpService: func() service.MLPService {
				mlpService := &mocks.MLPService{}
				mlpService.On(
					"GetEnvironment",
					"gods-dev",
				).Return(&merlin.Environment{}, nil)
				mlpService.On(
					"GetProject",
					models.ID(1),
				).Return(&mlp.Project{Id: 1}, nil)
				return mlpService
			},
			vars: RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixtureJSON(),
		},
		{
			name:     "non existent ensembler",
			path:     "/projects/1/jobs",
			method:   http.MethodPost,
			expected: NotFound("ensembler not found", errors.New("no exist").Error()),
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
				ensemblingJobService.On(
					"Save",
					generateEnsemblingJobFixture(1, models.ID(1), models.ID(1)),
				).Return(nil)
				return ensemblingJobService
			},
			mlpService: func() service.MLPService {
				mlpService := &mocks.MLPService{}
				mlpService.On(
					"GetEnvironment",
					"gods-dev",
				).Return(&merlin.Environment{}, nil)
				mlpService.On(
					"GetProject",
					models.ID(1),
				).Return(&mlp.Project{Id: 1}, nil)
				return mlpService
			},
			vars: RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixtureJSON(),
		},
		{
			name:     "invalid mlp environment",
			path:     "/projects/1/jobs",
			method:   http.MethodPost,
			expected: BadRequest("invalid environment", fmt.Sprintf("environment %s does not exist", "gods-dev")),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)
				return ensemblersSvc
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				ensemblingJobService := &mocks.EnsemblingJobService{}
				ensemblingJobService.On(
					"Save",
					generateEnsemblingJobFixture(1, models.ID(1), models.ID(1)),
				).Return(nil)
				return ensemblingJobService
			},
			mlpService: func() service.MLPService {
				mlpService := &mocks.MLPService{}
				mlpService.On(
					"GetEnvironment",
					"gods-dev",
				).Return(&merlin.Environment{}, errors.New("error mlp"))
				mlpService.On(
					"GetProject",
					models.ID(1),
				).Return(&mlp.Project{Id: 1}, nil)
				return mlpService
			},
			vars: RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixtureJSON(),
		},
		{
			name:     "non existent project",
			path:     "/projects/1/jobs",
			method:   http.MethodPost,
			expected: NotFound("project not found", errors.New("hello").Error()),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)
				return ensemblersSvc
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				ensemblingJobService := &mocks.EnsemblingJobService{}
				ensemblingJobService.On(
					"Save",
					generateEnsemblingJobFixture(1, models.ID(1), models.ID(1)),
				).Return(nil)
				return ensemblingJobService
			},
			mlpService: func() service.MLPService {
				mlpService := &mocks.MLPService{}
				mlpService.On(
					"GetEnvironment",
					"gods-dev",
				).Return(&merlin.Environment{}, nil)
				mlpService.On(
					"GetProject",
					models.ID(1),
				).Return(nil, errors.New("hello"))
				return mlpService
			},
			vars: RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixtureJSON(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ensemblersService := tt.ensemblersService()
			ensemblingJobService := tt.ensemblingJobService()
			mlpService := tt.mlpService()

			router := NewRouter(
				&AppContext{
					EnsemblersService:    ensemblersService,
					EnsemblingJobService: ensemblingJobService,
					MLPService:           mlpService,
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

			assert.Equal(t, expected.Body.String(), actual.Body.String())
			assert.Equal(t, expected.Code, actual.Code)
		})
	}
}
