package api

import (
	"errors"
	"fmt"
	"testing"

	"github.com/caraml-dev/turing/api/turing/batch"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	openapi "github.com/caraml-dev/turing/api/turing/generated"
	"github.com/caraml-dev/turing/api/turing/internal/ref"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/api/turing/service/mocks"
	"github.com/caraml-dev/turing/api/turing/validation"
)

var (
	annotationKeyOne   = "spark/spark.jars"
	annotationValueOne = "https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.0.1.jar"

	annotationKeyTwo   = "spark/spark.jars.packages"
	annotationValueTwo = "com.google.cloud.spark:spark-bigquery-with-dependencies_2.12:0.19.1"

	annotationKeyThree   = "hadoopConfiguration/fs.gs.impl"
	annotationValueThree = "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFileSystem"

	annotationKeyFour   = "hadoopConfiguration/fs.AbstractFileSystem.gs.impl"
	annotationValueFour = "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFS"
)

func GenerateEnsemblingJobFixture(
	i int,
	ensemblerID models.ID,
	projectID models.ID,
	name string,
	genExpected bool,
) *models.EnsemblingJob {
	nullableEnsemblingResources := openapi.NullableEnsemblingResources{}
	nullableEnsemblingResources.Set(&openapi.EnsemblingResources{
		DriverCpuRequest:      ref.String("1"),
		DriverMemoryRequest:   ref.String("1Gi"),
		ExecutorReplica:       ref.Int32(10),
		ExecutorCpuRequest:    ref.String("1"),
		ExecutorMemoryRequest: ref.String("1Gi"),
	})
	barString := "bar"
	envVars := []openapi.EnvVar{
		{
			Name:  "foo",
			Value: &barString,
		},
	}
	value := &models.EnsemblingJob{
		Name:        name,
		ProjectID:   projectID,
		EnsemblerID: ensemblerID,
		InfraConfig: &models.InfraConfig{
			EnsemblerInfraConfig: openapi.EnsemblerInfraConfig{
				ArtifactUri:        ref.String("gs://bucket/ensembler"),
				EnsemblerName:      ref.String("ensembler"),
				Resources:          nullableEnsemblingResources,
				Env:                &envVars,
				ServiceAccountName: ref.String(fmt.Sprintf("test-service-account-%d", i)),
			},
		},
		JobConfig: &models.JobConfig{
			Version: "v1",
			Kind:    openapi.ENSEMBLERCONFIGKIND_BATCH_ENSEMBLING_JOB,
			Metadata: &openapi.EnsemblingJobMeta{
				Name: fmt.Sprintf("test-batch-ensembling-%d", i),
				Annotations: map[string]string{
					annotationKeyOne:   annotationValueOne,
					annotationKeyTwo:   annotationValueTwo,
					annotationKeyThree: annotationValueThree,
					annotationKeyFour:  annotationValueFour,
				},
			},
			Spec: openapi.EnsemblingJobSpec{
				Source: openapi.EnsemblingJobSource{
					Dataset: openapi.Dataset{
						BigQueryDataset: &openapi.BigQueryDataset{
							Type: batch.DatasetTypeBQ,
							BqConfig: openapi.BigQueryDatasetConfig{
								Query: ref.String("select * from hello_world where customer_id = 4"),
								Options: map[string]string{
									"viewsEnabled":           "true",
									"materializationDataset": "dataset",
								},
							},
						},
					},
					JoinOn: []string{"customer_id", "target_date"},
				},
				Predictions: map[string]openapi.EnsemblingJobPredictionSource{
					"model_a": {
						Dataset: openapi.Dataset{
							BigQueryDataset: &openapi.BigQueryDataset{
								Type: batch.DatasetTypeBQ,
								BqConfig: openapi.BigQueryDatasetConfig{
									Table: ref.String("project.dataset.predictions_model_a"),
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
						Dataset: openapi.Dataset{
							BigQueryDataset: &openapi.BigQueryDataset{
								Type: batch.DatasetTypeBQ,
								BqConfig: openapi.BigQueryDatasetConfig{
									Query: ref.String("select * from hello_world where customer_id = 3"),
								},
							},
						},
						Columns: []string{"predictions"},
						JoinOn:  []string{"customer_id", "target_date"},
					},
				},
				Ensembler: openapi.EnsemblingJobEnsemblerSpec{
					Result: openapi.EnsemblingJobEnsemblerSpecResult{
						ColumnName: "prediction_score",
						Type:       openapi.ENSEMBLINGJOBRESULTTYPE_FLOAT,
						ItemType:   ref.EnsemblingJobResultType(openapi.ENSEMBLINGJOBRESULTTYPE_FLOAT),
					},
				},
				Sink: openapi.EnsemblingJobSink{
					BigQuerySink: &openapi.BigQuerySink{
						Type: batch.SinkTypeBQ,
						Columns: []string{
							"customer_id as customerId",
							"target_date",
							"results",
						},
						SaveMode: openapi.SAVEMODE_OVERWRITE,
						BqConfig: openapi.BigQuerySinkConfig{
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
	}

	if genExpected {
		value.JobConfig.Spec.Ensembler.Uri = "/home/spark/ensembler"
		value.EnvironmentName = "dev"
		if name == "" {
			value.Name = "test-ensembler-1"
		}
		value.InfraConfig.ArtifactUri = ref.String("gs://bucket/ensembler")
		value.InfraConfig.EnsemblerName = ref.String("ensembler")
	}
	return value
}

func CreateEnsembler(id int, ensemblerType string) models.EnsemblerLike {
	if ensemblerType == "pyfunc" {
		return &models.PyFuncEnsembler{
			GenericEnsembler: &models.GenericEnsembler{
				Name:      "ensembler",
				Model:     models.Model{ID: models.ID(id)},
				Type:      models.EnsemblerPyFuncType,
				ProjectID: 1,
			},
			ArtifactURI: "gs://bucket/ensembler",
		}
	}
	return &models.GenericEnsembler{
		Name:      "ensembler",
		Model:     models.Model{ID: 1},
		Type:      models.EnsemblerPyFuncType,
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
			expected: Accepted(GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "", true)),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(CreateEnsembler(1, "pyfunc"), nil)

				return ensemblersSvc
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				ensemblingJobService := &mocks.EnsemblingJobService{}
				ensemblingJobService.On(
					"CreateEnsemblingJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "", true), nil)

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
			body: GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(0), "", false),
		},
		"success | name provided": {
			expected: Accepted(GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true)),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(CreateEnsembler(1, "pyfunc"), nil)
				return ensemblersSvc
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				ensemblingJobService := &mocks.EnsemblingJobService{}
				ensemblingJobService.On(
					"CreateEnsemblingJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true), nil)
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
			body: GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(0), "test-ensembler-1", false),
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
			body: GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(0), "test-ensembler-1", false),
		},
		"failure | wrong type of ensembler": {
			expected: BadRequest("only pyfunc ensemblers allowed", "ensembler type given: *models.GenericEnsembler"),
			ensemblersService: func() service.EnsemblersService {
				ensemblersSvc := &mocks.EnsemblersService{}
				ensemblersSvc.On(
					"FindByID",
					mock.Anything,
					mock.Anything,
				).Return(CreateEnsembler(1, "generic"), nil)
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
			body: GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(0), "test-ensembler-1", false),
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
			body: GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(0), "test-ensembler-1", false),
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
				"job_id":     {"1"},
				"project_id": {"1"},
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("FindByID", mock.Anything, mock.Anything).Return(
					GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
					nil,
				)
				return svc
			},
			expectedResponseCode: 200,
			expectedBody: Ok(GenerateEnsemblingJobFixture(
				1,
				models.ID(1),
				models.ID(1),
				"test-ensembler-1",
				true,
			)),
		},
		"failure | no such ensembling job": {
			params: RequestVars{
				"job_id":     {"1"},
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
				"job_id": {"1"},
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
							GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
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
						GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
					},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				},
			),
		},
		"success | nominal with single status": {
			params: RequestVars{
				"project_id": {"1"},
				"status":     {"pending"},
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("List", mock.Anything, mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{
							GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
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
						GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
					},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				},
			),
		},
		"success | nominal with multiple statuses": {
			params: RequestVars{
				"project_id": {"1"},
				"status":     {"pending", "terminated"},
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("List", mock.Anything, mock.Anything).Return(
					&service.PaginatedResults{
						Results: []interface{}{
							GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
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
						GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
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
							GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
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
						GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
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
		"success | invalid status, it should still go through": {
			params: RequestVars{
				"project_id": {"1"},
				"status":     {"non_existent_status"},
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

func TestEnsemblingJobController_DeleteEnsemblingJob(t *testing.T) {
	var tests = map[string]struct {
		ensemblingJobService func() service.EnsemblingJobService
		params               RequestVars
		expectedResponseCode int
		expectedBody         *Response
	}{
		"success | job deleted": {
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("FindByID", mock.Anything, mock.Anything).Return(
					GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
					nil,
				)
				svc.On(
					"MarkEnsemblingJobForTermination",
					mock.Anything,
				).Return(nil)
				return svc
			},
			params: RequestVars{
				"job_id":     {"1"},
				"project_id": {"1"},
			},
			expectedResponseCode: 202,
			expectedBody:         Accepted(deleteEnsemblingJobResponse{ID: models.ID(1)}),
		},
		"failure | job not found": {
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("FindByID", mock.Anything, mock.Anything).Return(
					nil,
					fmt.Errorf("hello"),
				)
				return svc
			},
			params: RequestVars{
				"job_id":     {"1"},
				"project_id": {"1"},
			},
			expectedResponseCode: 404,
			expectedBody:         NotFound("ensembling job not found", fmt.Errorf("hello").Error()),
		},
		"failure | internal server error": {
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &mocks.EnsemblingJobService{}
				svc.On("FindByID", mock.Anything, mock.Anything).Return(
					GenerateEnsemblingJobFixture(1, models.ID(1), models.ID(1), "test-ensembler-1", true),
					nil,
				)
				svc.On(
					"MarkEnsemblingJobForTermination",
					mock.Anything,
				).Return(fmt.Errorf("hello"))
				return svc
			},
			params: RequestVars{
				"job_id":     {"1"},
				"project_id": {"1"},
			},
			expectedResponseCode: 500,
			expectedBody:         InternalServerError("unable to delete ensembling job", fmt.Errorf("hello").Error()),
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
			resp := ctrl.DeleteEnsemblingJob(nil, tt.params, nil)
			assert.Equal(t, tt.expectedResponseCode, resp.code)
			if tt.expectedBody != nil {
				assert.Equal(t, tt.expectedBody, resp)
			}
		})
	}
}
