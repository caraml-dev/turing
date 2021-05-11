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

func generateEnsemblingJobFixture(i int, ensemblerID models.ID, projectID models.ID) *models.EnsemblingJob {
	return &models.EnsemblingJob{
		Name:            fmt.Sprintf("test-ensembler-%d", i),
		VersionID:       models.ID(i),
		ProjectID:       models.ID(projectID),
		EnsemblerID:     ensemblerID,
		EnvironmentName: "gods-dev",
		InfraConfig: &models.InfraConfig{
			ServiceAccountName: fmt.Sprintf("test-service-account-%d", i),
			ImageRef:           "gcr.io/hello/world:123",
			Resources: &models.BatchEnsemblingJobResources{
				Requests: &models.Resource{
					CPU:    "2",
					Memory: "2Gi",
				},
				Limits: &models.Resource{
					CPU:    "2",
					Memory: "2Gi",
				},
			},
		},
		EnsemblerConfig: &models.EnsemblerConfig{
			EnsemblerConfig: batchensembler.BatchEnsemblingJob{
				Version: "v1",
				Kind:    batchensembler.BatchEnsemblingJob_BatchEnsemblingJob,
				Metadata: &batchensembler.BatchEnsemblingJobMetadata{
					Name: fmt.Sprintf("test-batch-ensembling-%d", i),
					Annotations: map[string]string{
						"spark/spark.jars":                                  "https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.0.1.jar",
						"spark/spark.jars.packages":                         "com.google.cloud.spark:spark-bigquery-with-dependencies_2.12:0.19.1",
						"hadoopConfiguration/fs.gs.impl":                    "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFileSystem",
						"hadoopConfiguration/fs.AbstractFileSystem.gs.impl": "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFS",
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
						"model_a": &batchensembler.PredictionSource{
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
						"model_b": &batchensembler.PredictionSource{
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
						Uri: "gs://bucket-name/my-ensembler/artifacts/ensembler",
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
}

func TestEnsemblingJobController_CreateEnsemblingJob(t *testing.T) {
	var tests = []struct {
		name                 string
		expected             *Response
		ensemblersService    func() service.EnsemblersService
		ensemblingJobService func() service.EnsemblingJobService
		mlpService           func() service.MLPService
		vars                 RequestVars
		body                 interface{}
	}{
		{
			name:     "nominal flow",
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
			body: generateEnsemblingJobFixture(1, models.ID(1), models.ID(0)),
		},
		{
			name:     "non existent ensembler",
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
			body: generateEnsemblingJobFixture(1, models.ID(1), models.ID(0)),
		},
		{
			name:     "invalid mlp environment",
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
			body: generateEnsemblingJobFixture(1, models.ID(1), models.ID(0)),
		},
		{
			name:     "non existent project",
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
			body: generateEnsemblingJobFixture(1, models.ID(1), models.ID(0)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
