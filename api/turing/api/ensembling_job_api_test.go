package api

import (
	"errors"
	"fmt"
	"testing"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
	"github.com/gojek/turing/api/turing/service/mocks"
	"github.com/gojek/turing/api/turing/validation"
	batchensembler "github.com/gojek/turing/engines/batch-ensembler/pkg/api/proto/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	annotationKeyOne   string = "spark/spark.jars"
	annotationValueOne string = "https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.0.1.jar"

	annotationKeyTwo   string = "spark/spark.jars.packages"
	annotationValueTwo string = "com.google.cloud.spark:spark-bigquery-with-dependencies_2.12:0.19.1"

	annotationKeyThree   string = "hadoopConfiguration/fs.gs.impl"
	annotationValueThree string = "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFileSystem"

	annotationKeyFour   string = "hadoopConfiguration/fs.AbstractFileSystem.gs.impl"
	annotationValueFour string = "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFS"
)

func generateEnsemblingJobFixture(
	i int,
	ensemblerID models.ID,
	projectID models.ID,
	name string,
	genExpected bool,
) *models.EnsemblingJob {
	value := &models.EnsemblingJob{
		Name:        name,
		ProjectID:   models.ID(projectID),
		EnsemblerID: ensemblerID,
		InfraConfig: &models.InfraConfig{
			ServiceAccountName: fmt.Sprintf("test-service-account-%d", i),
			Resources: &models.BatchEnsemblingJobResources{
				DriverCPURequest:      "1",
				DriverMemoryRequest:   "1Gi",
				ExecutorReplica:       10,
				ExecutorCPURequest:    "1",
				ExecutorMemoryRequest: "1Gi",
			},
		},
		JobConfig: &models.JobConfig{
			JobConfig: batchensembler.BatchEnsemblingJob{
				Version: "v1",
				Kind:    batchensembler.BatchEnsemblingJob_BatchEnsemblingJob,
				Metadata: &batchensembler.BatchEnsemblingJobMetadata{
					Name: fmt.Sprintf("test-batch-ensembling-%d", i),
					Annotations: map[string]string{
						annotationKeyOne:   annotationValueOne,
						annotationKeyTwo:   annotationValueTwo,
						annotationKeyThree: annotationValueThree,
						annotationKeyFour:  annotationValueFour,
					},
				},
				Spec: &batchensembler.BatchEnsemblingJobSpec{
					Source: &batchensembler.Source{
						Dataset: &batchensembler.Dataset{
							Type: batchensembler.Dataset_DatasetType(
								batchensembler.Dataset_BQ,
							),
							Config: &batchensembler.Dataset_BqConfig{
								BqConfig: &batchensembler.Dataset_BigQueryDatasetConfig{
									Query: "select * from helloworld where customer_id = 4",
									Options: map[string]string{
										"viewsEnabled":           "true",
										"materializationDataset": "dataset",
									},
								},
							},
						},
						JoinOn: []string{"customer_id", "target_date"},
					},
					Predictions: map[string]*batchensembler.PredictionSource{
						"model_a": {
							Dataset: &batchensembler.Dataset{
								Type: batchensembler.Dataset_DatasetType(
									batchensembler.Dataset_BQ,
								),
								Config: &batchensembler.Dataset_BqConfig{
									BqConfig: &batchensembler.Dataset_BigQueryDatasetConfig{
										Table: "project.dataset.predictions_model_a",
										Features: []string{
											"customer_id",
											"target_date",
											"predictions",
										},
									},
								},
							},
							Columns: []string{"predictions"},
							JoinOn:  []string{"customer_id", "target_date"},
						},
						"model_b": {
							Dataset: &batchensembler.Dataset{
								Type: batchensembler.Dataset_DatasetType(
									batchensembler.Dataset_BQ,
								),
								Config: &batchensembler.Dataset_BqConfig{
									BqConfig: &batchensembler.Dataset_BigQueryDatasetConfig{
										Query: "select * from helloworld where customer_id = 3",
									},
								},
							},
							Columns: []string{"predictions"},
							JoinOn:  []string{"customer_id", "target_date"},
						},
					},
					Ensembler: &batchensembler.Ensembler{
						Result: &batchensembler.Ensembler_Result{
							ColumnName: "prediction_score",
							Type:       batchensembler.Ensembler_FLOAT,
							ItemType:   batchensembler.Ensembler_FLOAT,
						},
					},
					Sink: &batchensembler.Sink{
						Type: batchensembler.Sink_BQ,
						Columns: []string{
							"customer_id as customerId",
							"target_date",
							"results",
						},
						SaveMode: batchensembler.Sink_OVERWRITE,
						Config: &batchensembler.Sink_BqConfig{
							BqConfig: &batchensembler.Sink_BigQuerySinkConfig{
								Table:         "project.dataset.ensembling_results",
								StagingBucket: "bucket-name",
								Options: map[string]string{
									"partitionField": "target_date",
								},
							},
						},
					},
				},
			},
		},
	}

	if genExpected {
		value.JobConfig.JobConfig.Spec.Ensembler.Uri = "/home/spark/ensembler"
		value.EnvironmentName = "dev"
		if name == "" {
			value.Name = "test-ensembler-1"
		}
		value.InfraConfig.ArtifactURI = "gs://bucket/ensembler"
		value.InfraConfig.EnsemblerName = "ensembler"
	}
	return value
}

func createEnsembler(id int, ensemblerType string) models.EnsemblerLike {
	if ensemblerType == "pyfunc" {
		return &models.PyFuncEnsembler{
			GenericEnsembler: &models.GenericEnsembler{
				Name:      "ensembler",
				Model:     models.Model{ID: 1},
				Type:      models.EnsemblerTypePyFunc,
				ProjectID: 1,
			},
			ArtifactURI: "gs://bucket/ensembler",
		}
	}
	return &models.GenericEnsembler{
		Name:      "ensembler",
		Model:     models.Model{ID: 1},
		Type:      models.EnsemblerTypePyFunc,
		ProjectID: 1,
	}
}

func TestEnsemblingJobController_CreateEnsemblingJob(t *testing.T) {
	var tests = map[string]struct {
		expected             *Response
		ensemblersService    func() service.EnsemblersService
		ensemblingJobService func() service.EnsemblingJobService
		mlpService           func() service.MLPService
		vars                 RequestVars
		body                 interface{}
	}{
		"success | name not provided": {
			expected: Accepted(generateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "", true)),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(createEnsembler(1, "pyfunc"), nil)

				return ensemblersSvc
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				ensemblingJobService := &mocks.EnsemblingJobService{}
				ensemblingJobService.On(
					"CreateEnsemblingJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(generateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "", true), nil)

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
				).Return(&mlp.Project{Id: 1}, nil)
				return mlpService
			},
			vars: RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixture(1, models.ID(1), models.ID(0), "", false),
		},
		"success | name provided": {
			expected: Accepted(generateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true)),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(createEnsembler(1, "pyfunc"), nil)
				return ensemblersSvc
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				ensemblingJobService := &mocks.EnsemblingJobService{}
				ensemblingJobService.On(
					"CreateEnsemblingJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(generateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true), nil)
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
				).Return(&mlp.Project{Id: 1}, nil)
				return mlpService
			},
			vars: RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixture(1, models.ID(1), models.ID(0), "test-ensembler-1", false),
		},
		"failure | non existent ensembler": {
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
				).Return(&mlp.Project{Id: 1}, nil)
				return mlpService
			},
			vars: RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixture(1, models.ID(1), models.ID(0), "test-ensembler-1", false),
		},
		"failure | wrong type of ensembler": {
			expected: BadRequest("only pyfunc ensemblers allowed", "ensembler type given: *models.GenericEnsembler"),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(createEnsembler(1, "generic"), nil)
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
				).Return(&mlp.Project{Id: 1}, nil)
				return mlpService
			},
			vars: RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixture(1, models.ID(1), models.ID(0), "test-ensembler-1", false),
		},
		"failure | non existent project": {
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
				).Return(nil, errors.New("hello"))
				return mlpService
			},
			vars: RequestVars{
				"project_id": {"1"},
			},
			body: generateEnsemblingJobFixture(1, models.ID(1), models.ID(0), "test-ensembler-1", false),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ensemblersService := tt.ensemblersService()
			ensemblingJobService := tt.ensemblingJobService()
			mlpService := tt.mlpService()

			validator, _ := validation.NewValidator(nil)
			ctrl := &EnsemblingJobController{
				NewBaseController(
					&AppContext{
						EnsemblersService:    ensemblersService,
						EnsemblingJobService: ensemblingJobService,
						MLPService:           mlpService,
					},
					validator,
				),
			}
			response := ctrl.Create(nil, tt.vars, tt.body)
			assert.Equal(t, tt.expected, response)
		})
	}
}

func TestEnsemblingJobController_GetEnsemblingJob(t *testing.T) {
	tests := map[string]struct {
		params               RequestVars
		ensemblingJobService func() service.EnsemblingJobService
		expectedResponseCode int
		expectedBody         *Response
	}{
		"success | nominal": {
			params: RequestVars{
				"id":         {"1"},
				"project_id": {"1"},
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("FindByID", mock.Anything, mock.Anything).Return(
					generateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
					nil,
				)
				return svc
			},
			expectedResponseCode: 200,
			expectedBody: Ok(generateEnsemblingJobFixture(
				1,
				models.ID(1),
				models.ID(1),
				"test-ensembler-1",
				true,
			)),
		},
		"failure | no such ensembling job": {
			params: RequestVars{
				"id":         {"1"},
				"project_id": {"1"},
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("FindByID", mock.Anything, mock.Anything).Return(
					nil,
					errors.New("hello"),
				)
				return svc
			},
			expectedResponseCode: 404,
			expectedBody:         nil,
		},
		"failure | missing project_id": {
			params: RequestVars{
				"id": {"1"},
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				return svc
			},
			expectedResponseCode: 400,
			expectedBody:         nil,
		},
		"failure | missing ensembling id": {
			params: RequestVars{
				"project_id": {"1"},
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				return svc
			},
			expectedResponseCode: 400,
			expectedBody:         nil,
		},
		"failure | missing all params": {
			params: RequestVars{},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				return svc
			},
			expectedResponseCode: 400,
			expectedBody:         nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			svc := tt.ensemblingJobService()
			validator, _ := validation.NewValidator(nil)
			ctrl := &EnsemblingJobController{
				NewBaseController(
					&AppContext{
						EnsemblingJobService: svc,
					},
					validator,
				),
			}
			resp := ctrl.GetEnsemblingJob(nil, tt.params, nil)
			assert.Equal(t, tt.expectedResponseCode, resp.code)
			if tt.expectedBody != nil {
				assert.Equal(t, tt.expectedBody, resp)
			}
		})
	}
}

func TestEnsemblingJobController_ListEnsemblingJob(t *testing.T) {
	tests := map[string]struct {
		params               RequestVars
		ensemblingJobService func() service.EnsemblingJobService
		expectedResponseCode int
		expectedBody         *Response
	}{
		"success | nominal": {
			params: RequestVars{
				"project_id": {"1"},
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("List", mock.Anything, mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{
							generateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
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
			expectedResponseCode: 200,
			expectedBody: Ok(
				&service.PaginatedResults{
					Results: []interface{}{
						generateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
					},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				},
			),
		},
		"success | nominal with paging": {
			params: RequestVars{
				"project_id": {"1"},
				"page":       {"1"},
				"pageSize":   {"20"},
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("List", mock.Anything, mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{
							generateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
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
			expectedResponseCode: 200,
			expectedBody: Ok(
				&service.PaginatedResults{
					Results: []interface{}{
						generateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
					},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				},
			),
		},
		"success | no result": {
			params: RequestVars{
				"project_id": {"1"},
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("List", mock.Anything, mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{},
						Paging: service.Paging{
							Total: 0,
							Page:  1,
							Pages: 1,
						},
					},
					nil,
				)
				return svc
			},
			expectedResponseCode: 200,
			expectedBody: Ok(
				&service.PaginatedResults{
					Results: []interface{}{},
					Paging: service.Paging{
						Total: 0,
						Page:  1,
						Pages: 1,
					},
				},
			),
		},
		"failure | missing value": {
			params: RequestVars{},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				return svc
			},
			expectedResponseCode: 400,
			expectedBody:         nil,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			svc := tt.ensemblingJobService()
			validator, _ := validation.NewValidator(nil)
			ctrl := &EnsemblingJobController{
				NewBaseController(
					&AppContext{
						EnsemblingJobService: svc,
					},
					validator,
				),
			}
			resp := ctrl.ListEnsemblingJobs(nil, tt.params, nil)
			assert.Equal(t, tt.expectedResponseCode, resp.code)
			if tt.expectedBody != nil {
				assert.Equal(t, tt.expectedBody, resp)
			}
		})
	}
}
