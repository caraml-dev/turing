// +build integration

package service

import (
	"fmt"
	"testing"

	"github.com/gojek/turing/api/turing/batch"
	"github.com/gojek/turing/api/turing/config"
	openapi "github.com/gojek/turing/api/turing/generated"
	"github.com/gojek/turing/api/turing/internal/ref"
	"github.com/gojek/turing/api/turing/it/database"
	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

var (
	driverCPURequest            = "100"
	driverMemoryRequest         = "1Gi"
	executorReplica       int32 = 2
	executorCPURequest          = "1"
	executorMemoryRequest       = "1Gi"
)
var defaultConfigurations = config.DefaultEnsemblingJobConfigurations{
	BatchEnsemblingJobResources: openapi.EnsemblingResources{
		DriverCpuRequest:      &driverCPURequest,
		DriverMemoryRequest:   &driverMemoryRequest,
		ExecutorReplica:       &executorReplica,
		ExecutorCpuRequest:    &executorCPURequest,
		ExecutorMemoryRequest: &executorMemoryRequest,
	},
	SparkConfigAnnotations: map[string]string{
		"spark/spark.sql.execution.arrow.pyspark.enabled": "true",
	},
}

func generateEnsemblingJobFixture(
	i int,
	ensemblerID models.ID,
	projectID models.ID,
	name string,
	genExpected bool,
) *models.EnsemblingJob {
	value := &models.EnsemblingJob{
		Name:            name,
		EnsemblerID:     ensemblerID,
		ProjectID:       projectID,
		EnvironmentName: "dev",
		InfraConfig: &models.InfraConfig{
			ServiceAccountName: fmt.Sprintf("test-service-account-%d", i),
			Resources: &openapi.EnsemblingResources{
				DriverCpuRequest:      ref.String("1"),
				DriverMemoryRequest:   ref.String("1Gi"),
				ExecutorReplica:       ref.Int32(10),
				ExecutorCpuRequest:    ref.String("1"),
				ExecutorMemoryRequest: ref.String("1Gi"),
			},
		},
		JobConfig: &models.JobConfig{
			Version: "v1",
			Kind:    openapi.ENSEMBLERCONFIGKIND_BATCH_ENSEMBLING_JOB,
			Metadata: &openapi.EnsemblingJobMeta{
				Name: fmt.Sprintf("test-batch-ensembling-%d", i),
				Annotations: map[string]string{
					"spark/spark.jars":                                  "https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.0.1.jar",
					"spark/spark.jars.packages":                         "com.google.cloud.spark:spark-bigquery-with-dependencies_2.12:0.19.1",
					"hadoopConfiguration/fs.gs.impl":                    "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFileSystem",
					"hadoopConfiguration/fs.AbstractFileSystem.gs.impl": "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFS",
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
					Uri: "gs://bucket-name/my-ensembler/artifacts/ensembler",
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
		value.InfraConfig.ArtifactURI = "gs://bucket/ensembler"
		value.InfraConfig.EnsemblerName = "ensembler"
	}

	return value
}

func TestSaveAndFindByIDEnsemblingJobIntegration(t *testing.T) {
	t.Run("success | insertion with no errors", func(t *testing.T) {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			ensemblingJobService := NewEnsemblingJobService(db, "dev", defaultConfigurations)

			projectID := models.ID(1)
			ensemblerID := models.ID(1000)
			ensemblingJob := generateEnsemblingJobFixture(1, ensemblerID, projectID, "test-ensembler", false)
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
			assert.Equal(t, found.EnsemblerID, ensemblingJob.EnsemblerID)
			assert.Equal(t, found.ProjectID, ensemblingJob.ProjectID)
			assert.Equal(t, found.EnvironmentName, ensemblingJob.EnvironmentName)
			assert.Equal(t, models.JobPending, ensemblingJob.Status)
			assert.Equal(t, found.InfraConfig, ensemblingJob.InfraConfig)
			assert.Equal(t, found.JobConfig, ensemblingJob.JobConfig)
		})
	})
}

func TestListEnsemblingJobIntegration(t *testing.T) {
	tests := map[string]struct {
		saveQuantity  int
		queryQuantity int
		pageNumber    int
		expectedCount int
	}{
		"success | first page nominal": {
			saveQuantity:  5,
			queryQuantity: 5,
			pageNumber:    1,
			expectedCount: 5,
		},
		"success | first page nominal, over query": {
			saveQuantity:  5,
			queryQuantity: 10,
			pageNumber:    1,
			expectedCount: 5,
		},
		"success | first page nominal, under query": {
			saveQuantity:  5,
			queryQuantity: 3,
			pageNumber:    1,
			expectedCount: 3,
		},
		"success | second page nominal": {
			saveQuantity:  10,
			queryQuantity: 5,
			pageNumber:    2,
			expectedCount: 5,
		},
		"success | second page nominal, under query": {
			saveQuantity:  9,
			queryQuantity: 5,
			pageNumber:    2,
			expectedCount: 4,
		},
		"success | second page nominal, over query": {
			saveQuantity:  6,
			queryQuantity: 5,
			pageNumber:    2,
			expectedCount: 1,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
				ensemblingJobService := NewEnsemblingJobService(db, "dev", defaultConfigurations)

				for saveCounter := 0; saveCounter < tt.saveQuantity; saveCounter++ {
					projectID := models.ID(1)
					ensemblerID := models.ID(1000)
					ensemblingJob := generateEnsemblingJobFixture(1, ensemblerID, projectID, "test-ensembler", false)
					err := ensemblingJobService.Save(ensemblingJob)
					assert.NoError(t, err)
					assert.NotEqual(t, models.ID(0), ensemblingJob.ID)
				}

				// Query pending ensembling jobs
				fetched, err := ensemblingJobService.List(
					EnsemblingJobListOptions{
						PaginationOptions: PaginationOptions{
							Page:     ref.Int(tt.pageNumber),
							PageSize: &tt.queryQuantity,
						},
					},
				)
				assert.Nil(t, err)
				assert.Equal(t, tt.saveQuantity, fetched.Paging.Total)
				assert.Equal(t, tt.pageNumber, fetched.Paging.Page)

				ensemblingJobs := fetched.Results.([]*models.EnsemblingJob)
				assert.Equal(t, tt.expectedCount, len(ensemblingJobs))
			})
		})
	}
}

func TestFindPendingJobsAndUpdateIntegration(t *testing.T) {
	t.Run("success | find pending jobs and update with no errors", func(t *testing.T) {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			ensemblingJobService := NewEnsemblingJobService(db, "dev", defaultConfigurations)

			// Save job
			projectID := models.ID(1)
			ensemblerID := models.ID(1000)
			ensemblingJob := generateEnsemblingJobFixture(1, ensemblerID, projectID, "test-ensembler", false)
			err := ensemblingJobService.Save(ensemblingJob)
			assert.NoError(t, err)
			assert.NotEqual(t, models.ID(0), ensemblingJob.ID)

			// Query pending ensembling jobs
			pageSize := 10
			retryCountLessThan := 3
			fetched, err := ensemblingJobService.List(
				EnsemblingJobListOptions{
					PaginationOptions: PaginationOptions{
						Page:     ref.Int(1),
						PageSize: &pageSize,
					},
					Statuses:           []models.Status{models.JobPending},
					RetryCountLessThan: &retryCountLessThan,
				},
			)
			assert.NoError(t, err)
			assert.Equal(t, 1, fetched.Paging.Total)
			ensemblingJobs := fetched.Results.([]*models.EnsemblingJob)

			queriedEnsemblingJob := ensemblingJobs[0]
			assert.Equal(t, models.Status("pending"), queriedEnsemblingJob.Status)

			// Update pending job
			queriedEnsemblingJob.Status = models.JobFailedSubmission
			queriedEnsemblingJob.Error = "error"
			err = ensemblingJobService.Save(queriedEnsemblingJob)
			assert.NoError(t, err)

			// Query back
			found, err := ensemblingJobService.FindByID(
				ensemblingJob.ID,
				EnsemblingJobFindByIDOptions{ProjectID: &projectID},
			)
			assert.NoError(t, err)
			assert.Equal(t, models.JobFailedSubmission, found.Status)
		})
	})
}

func TestCreateEnsemblingJob(t *testing.T) {
	var tests = map[string]struct {
		ensembler              *models.PyFuncEnsembler
		request                *models.EnsemblingJob
		expected               *models.EnsemblingJob
		removeDefaultResources bool
		removeDriverCPURequest bool
	}{
		"success | name provided": {
			ensembler: &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					Name:      "ensembler",
					Model:     models.Model{ID: 1},
					Type:      models.EnsemblerTypePyFunc,
					ProjectID: 1,
				},
				ArtifactURI: "gs://bucket/ensembler",
			},
			request:                generateEnsemblingJobFixture(1, 1, 1, "test-ensembler", false),
			expected:               generateEnsemblingJobFixture(1, 1, 1, "test-ensembler", true),
			removeDefaultResources: false,
			removeDriverCPURequest: false,
		},
		"success | name not provided": {
			ensembler: &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					Name:      "ensembler",
					Model:     models.Model{ID: 1},
					Type:      models.EnsemblerTypePyFunc,
					ProjectID: 1,
				},
				ArtifactURI: "gs://bucket/ensembler",
			},
			request:                generateEnsemblingJobFixture(1, 1, 1, "", false),
			expected:               generateEnsemblingJobFixture(1, 1, 1, "", true),
			removeDefaultResources: false,
			removeDriverCPURequest: false,
		},
		"success | default resources removed": {
			ensembler: &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					Name:      "ensembler",
					Model:     models.Model{ID: 1},
					Type:      models.EnsemblerTypePyFunc,
					ProjectID: 1,
				},
				ArtifactURI: "gs://bucket/ensembler",
			},
			request:                generateEnsemblingJobFixture(1, 1, 1, "test-ensembler", false),
			expected:               generateEnsemblingJobFixture(1, 1, 1, "test-ensembler", true),
			removeDefaultResources: true,
			removeDriverCPURequest: false,
		},
		"success | remove 1 setting from resources": {
			ensembler: &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					Name:      "ensembler",
					Model:     models.Model{ID: 1},
					Type:      models.EnsemblerTypePyFunc,
					ProjectID: 1,
				},
				ArtifactURI: "gs://bucket/ensembler",
			},
			request:                generateEnsemblingJobFixture(1, 1, 1, "test-ensembler", false),
			expected:               generateEnsemblingJobFixture(1, 1, 1, "test-ensembler", true),
			removeDefaultResources: false,
			removeDriverCPURequest: true,
		},
	}

	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				ensemblingJobService := NewEnsemblingJobService(db, "dev", defaultConfigurations)

				if tt.removeDefaultResources {
					tt.request.InfraConfig.Resources = nil
				}

				if tt.removeDriverCPURequest {
					tt.request.InfraConfig.Resources.DriverCpuRequest = nil
				}

				result, err := ensemblingJobService.CreateEnsemblingJob(
					tt.request,
					1,
					tt.ensembler,
				)
				assert.Nil(t, err)
				expected := tt.expected

				assert.NotEqual(t, models.ID(0), result.ID)

				assert.NotEqual(t, result.Name, "")
				assert.Equal(t, expected.EnsemblerID, result.EnsemblerID)
				assert.Equal(t, expected.ProjectID, result.ProjectID)
				assert.Equal(t, expected.EnvironmentName, result.EnvironmentName)
				assert.Equal(t, models.JobPending, result.Status)

				assert.Equal(
					t,
					result.JobConfig.Spec.Ensembler.Uri,
					expected.JobConfig.Spec.Ensembler.Uri,
				)

				assert.Equal(t, expected.InfraConfig.ArtifactURI, result.InfraConfig.ArtifactURI)
				assert.Equal(t, expected.InfraConfig.EnsemblerName, result.InfraConfig.EnsemblerName)

				// Check if merging of spark config is done properly
				sparkMap := result.JobConfig.Metadata.Annotations
				expectedKeys := map[string]string{
					"spark/spark.jars":                                  "https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.0.1.jar",
					"spark/spark.jars.packages":                         "com.google.cloud.spark:spark-bigquery-with-dependencies_2.12:0.19.1",
					"hadoopConfiguration/fs.gs.impl":                    "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFileSystem",
					"hadoopConfiguration/fs.AbstractFileSystem.gs.impl": "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFS",
					"spark/spark.sql.execution.arrow.pyspark.enabled":   "true",
				}
				for key, expectedValue := range expectedKeys {
					resultValue, ok := sparkMap[key]
					assert.Equal(t, true, ok)
					assert.Equal(t, expectedValue, resultValue)
				}

				if tt.removeDriverCPURequest {
					assert.Equal(
						t,
						defaultConfigurations.BatchEnsemblingJobResources.DriverCpuRequest,
						result.InfraConfig.Resources.DriverCpuRequest,
					)
				}
			})
		}
	})
}

func TestMarkEnsemblingJobForTermination(t *testing.T) {
	t.Run("success | delete ensembling job", func(t *testing.T) {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			ensemblingJobService := NewEnsemblingJobService(db, "dev", defaultConfigurations)

			// Save job
			projectID := models.ID(1)
			ensemblerID := models.ID(1000)
			ensemblingJob := generateEnsemblingJobFixture(1, ensemblerID, projectID, "test-ensembler", false)
			err := ensemblingJobService.Save(ensemblingJob)
			assert.NoError(t, err)
			assert.NotEqual(t, models.ID(0), ensemblingJob.ID)

			// Delete job
			err = ensemblingJobService.MarkEnsemblingJobForTermination(ensemblingJob)
			assert.NoError(t, err)

			// Query back job to check if terminated
			found, err := ensemblingJobService.FindByID(
				ensemblingJob.ID,
				EnsemblingJobFindByIDOptions{ProjectID: &projectID},
			)
			assert.NoError(t, err)
			assert.Equal(t, models.JobTerminating, found.Status)
		})
	})
}

func TestPhysicalDeleteEnsemblingJob(t *testing.T) {
	t.Run("success | delete ensembling job", func(t *testing.T) {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			ensemblingJobService := NewEnsemblingJobService(db, "dev", defaultConfigurations)

			// Save job
			projectID := models.ID(1)
			ensemblerID := models.ID(1000)
			ensemblingJob := generateEnsemblingJobFixture(1, ensemblerID, projectID, "test-ensembler", false)
			err := ensemblingJobService.Save(ensemblingJob)
			assert.NoError(t, err)
			assert.NotEqual(t, models.ID(0), ensemblingJob.ID)

			// Delete job
			err = ensemblingJobService.Delete(ensemblingJob)
			assert.NoError(t, err)

			// Query back job to check if job is no longer there
			found, err := ensemblingJobService.FindByID(
				ensemblingJob.ID,
				EnsemblingJobFindByIDOptions{ProjectID: &projectID},
			)
			assert.NotNil(t, err)
			assert.Nil(t, found)
		})
	})
}
