// +build integration

package service

import (
	"fmt"
	"testing"

	"github.com/gojek/turing/api/turing/it/database"
	"github.com/gojek/turing/api/turing/models"
	batchensembler "github.com/gojek/turing/engines/batch-ensembler/pkg/api/proto/v1"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func generateEnsemblingJobFixture(i int, ensemblerID models.ID, projectID models.ID) *models.EnsemblingJob {
	return &models.EnsemblingJob{
		Name:            fmt.Sprintf("test-ensembler-%d", i),
		VersionID:       models.ID(i),
		EnsemblerID:     ensemblerID,
		ProjectID:       projectID,
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

func TestSaveAndFindByIDEnsemblingJobIntegration(t *testing.T) {
	t.Run("insertion with no errors", func(t *testing.T) {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			ensemblingJobService := NewEnsemblingJobService(db)

			projectID := models.ID(1)
			ensemblerID := models.ID(1000)
			ensemblingJob := generateEnsemblingJobFixture(1, ensemblerID, projectID)
			err := ensemblingJobService.Save(ensemblingJob)
			assert.NoError(t, err)
			assert.NotEqual(t, models.ID(0), ensemblingJob.ID)

			found, err := ensemblingJobService.FindByID(
				ensemblingJob.ID,
				EnsemblingJobFindByIDOptions{ProjectID: &projectID},
			)
			assert.NoError(t, err)

			assert.NotEqual(t, found.ID, models.ID(0))
			assert.Equal(t, found.Name, ensemblingJob.Name)
			assert.Equal(t, found.VersionID, ensemblingJob.VersionID)
			assert.Equal(t, found.EnsemblerID, ensemblingJob.EnsemblerID)
			assert.Equal(t, found.ProjectID, ensemblingJob.ProjectID)
			assert.Equal(t, found.EnvironmentName, ensemblingJob.EnvironmentName)
			assert.Equal(t, models.State("pending"), ensemblingJob.Status)
			assert.Equal(t, found.InfraConfig, ensemblingJob.InfraConfig)
			assert.Equal(t, found.EnsemblerConfig, ensemblingJob.EnsemblerConfig)
		})
	})
}

func TestFindPendingJobsAndUpdateIntegration(t *testing.T) {
	t.Run("insertion with no errors", func(t *testing.T) {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			ensemblingJobService := NewEnsemblingJobService(db)

			// Save job
			projectID := models.ID(1)
			ensemblerID := models.ID(1000)
			ensemblingJob := generateEnsemblingJobFixture(1, ensemblerID, projectID)
			err := ensemblingJobService.Save(ensemblingJob)
			assert.NoError(t, err)
			assert.NotEqual(t, models.ID(0), ensemblingJob.ID)

			// Query pending ensembling jobs
			ensemblingJobs, err := ensemblingJobService.FindPendingJobs(10)
			assert.NoError(t, err)
			assert.Equal(t, 1, len(ensemblingJobs))

			queriedEnsemblingJob := ensemblingJobs[0]
			assert.Equal(t, models.State("pending"), queriedEnsemblingJob.Status)

			// Update pending job
			err = ensemblingJobService.UpdateJobStatus(
				queriedEnsemblingJob.ID,
				models.JobFailedSubmission,
				"error",
			)
			assert.NoError(t, err)

			// Query back
			found, err := ensemblingJobService.FindByID(
				ensemblingJob.ID,
				EnsemblingJobFindByIDOptions{ProjectID: &projectID},
			)
			assert.NoError(t, err)
			assert.Equal(t, models.State("failed_submission"), found.Status)
		})
	})
}
