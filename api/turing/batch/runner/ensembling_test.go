package batchrunner

import (
	"testing"
	"time"

	mlp "github.com/gojek/mlp/client"
	batchcontroller "github.com/gojek/turing/api/turing/batch/controller"
	batchcontrollermock "github.com/gojek/turing/api/turing/batch/controller/mocks"
	"github.com/gojek/turing/api/turing/imagebuilder"
	imagebuildermock "github.com/gojek/turing/api/turing/imagebuilder/mocks"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
	servicemock "github.com/gojek/turing/api/turing/service/mocks"
	batchensembler "github.com/gojek/turing/engines/batch-ensembler/pkg/api/proto/v1"
	"github.com/stretchr/testify/mock"
)

const (
	annotationKeyOne     = "spark/spark.jars"
	annotationValueOne   = "https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.0.1.jar"
	annotationKeyTwo     = "spark/spark.jars.packages"
	annotationValueTwo   = "com.google.cloud.spark:spark-bigquery-with-dependencies_2.12:0.19.1"
	annotationKeyThree   = "hadoopConfiguration/fs.gs.impl"
	annotationValueThree = "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFileSystem"
	annotationKeyFour    = "hadoopConfiguration/fs.AbstractFileSystem.gs.impl"
	annotationValueFour  = "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFS"
)

func generateEnsemblingJobFixture() *models.EnsemblingJob {
	return &models.EnsemblingJob{
		Name:            "test-ensembler-1",
		ProjectID:       models.ID(1),
		EnsemblerID:     models.ID(1),
		EnvironmentName: "dev",
		InfraConfig: &models.InfraConfig{
			ArtifactURI:        "gs://bucket/ensembler",
			EnsemblerName:      "ensembler",
			ServiceAccountName: "test-service-account",
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
					Name: "test-batch-ensembling",
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
						Uri: "/home/spark/ensembler",
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

func TestRun(t *testing.T) {
	// Unfortunately this is hard to test as we need Kubernetes integration
	// and a Spark Operator. Testing with an actual cluster is required.
	// Here we just try to run it without throwing an exception.
	var tests = map[string]struct {
		environment          string
		ensemblingController func() batchcontroller.EnsemblingController
		imageBuilder         func() imagebuilder.ImageBuilder
		ensemblingJobService func() service.EnsemblingJobService
		mlpService           func() service.MLPService
	}{
		"success | nominal": {
			environment: "testing",
			ensemblingController: func() batchcontroller.EnsemblingController {
				ctlr := &batchcontrollermock.EnsemblingController{}
				ctlr.On(
					"Create",
					mock.Anything,
				).Return(nil)
				ctlr.On(
					"GetStatus",
					mock.Anything,
					mock.Anything,
				).Return(batchcontroller.SparkApplicationStateCompleted, nil)
				return ctlr
			},
			imageBuilder: func() imagebuilder.ImageBuilder {
				ib := &imagebuildermock.ImageBuilder{}
				ib.On(
					"BuildImage",
					mock.Anything,
					mock.Anything,
				).Return("ghcr.io/test-project/mymodel:1", nil)
				return ib
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &servicemock.EnsemblingJobService{}

				svc.On(
					"List",
					mock.Anything,
				).Return(&service.PaginatedResults{
					Results: []*models.EnsemblingJob{generateEnsemblingJobFixture()},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				}, nil).Once()

				svc.On(
					"Save",
					mock.Anything,
				).Return(nil)

				newFixture := generateEnsemblingJobFixture()
				newFixture.Status = models.JobRunning
				svc.On(
					"List",
					mock.Anything,
				).Return(&service.PaginatedResults{
					Results: []*models.EnsemblingJob{newFixture},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				}, nil)

				return svc
			},
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetProject",
					mock.Anything,
					mock.Anything,
				).Return(&mlp.Project{Id: 1}, nil)
				return svc
			},
		},
		"success | imagebuilding stuck": {
			environment: "testing",
			ensemblingController: func() batchcontroller.EnsemblingController {
				ctlr := &batchcontrollermock.EnsemblingController{}
				ctlr.On(
					"Create",
					mock.Anything,
				).Return(nil)
				return ctlr
			},
			imageBuilder: func() imagebuilder.ImageBuilder {
				ib := &imagebuildermock.ImageBuilder{}
				ib.On(
					"BuildImage",
					mock.Anything,
					mock.Anything,
				).Return("ghcr.io/test-project/mymodel:1", nil)
				ib.On(
					"GetImageBuildingJobStatus",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(imagebuilder.JobStatusFailed, nil)
				return ib
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &servicemock.EnsemblingJobService{}

				svc.On(
					"List",
					mock.Anything,
				).Return(&service.PaginatedResults{
					Results: []*models.EnsemblingJob{generateEnsemblingJobFixture()},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				}, nil).Once()

				svc.On(
					"Save",
					mock.Anything,
				).Return(nil)

				newFixture := generateEnsemblingJobFixture()
				newFixture.Status = models.JobBuildingImage
				svc.On(
					"List",
					mock.Anything,
				).Return(&service.PaginatedResults{
					Results: []*models.EnsemblingJob{newFixture},
					Paging: service.Paging{
						Total: 1,
						Page:  1,
						Pages: 1,
					},
				}, nil)

				return svc
			},
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetProject",
					mock.Anything,
					mock.Anything,
				).Return(&mlp.Project{Id: 1}, nil)
				return svc
			},
		},
		"success | no ensembling jobs": {
			environment: "testing",
			ensemblingController: func() batchcontroller.EnsemblingController {
				ctlr := &batchcontrollermock.EnsemblingController{}
				ctlr.On(
					"Create",
					mock.Anything,
				).Return(nil)
				return ctlr
			},
			imageBuilder: func() imagebuilder.ImageBuilder {
				ib := &imagebuildermock.ImageBuilder{}
				ib.On(
					"BuildImage",
					mock.Anything,
					mock.Anything,
				).Return("ghcr.io/test-project/mymodel:1", nil)
				return ib
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				svc := &servicemock.EnsemblingJobService{}

				svc.On(
					"List",
					mock.Anything,
				).Return(&service.PaginatedResults{
					Results: []*models.EnsemblingJob{},
					Paging: service.Paging{
						Total: 0,
						Page:  1,
						Pages: 1,
					},
				}, nil)

				svc.On(
					"Save",
					mock.Anything,
				).Return(nil)

				return svc
			},
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetProject",
					mock.Anything,
					mock.Anything,
				).Return(&mlp.Project{Id: 1}, nil)
				return svc
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ensemblingController := tt.ensemblingController()
			ensemblingJobService := tt.ensemblingJobService()
			mlpService := tt.mlpService()
			imageBuilder := tt.imageBuilder()

			r := NewBatchEnsemblingJobRunner(
				ensemblingController,
				ensemblingJobService,
				mlpService,
				imageBuilder,
				tt.environment,
				10,
				3,
				10*time.Minute,
			)
			r.Run()
		})
	}
}
