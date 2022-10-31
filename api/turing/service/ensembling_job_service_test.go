//go:build integration

package service_test

import (
	"fmt"
	"testing"

	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/api/turing/service/mocks"

	mlp "github.com/gojek/mlp/api/client"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/caraml-dev/turing/api/turing/batch"
	"github.com/caraml-dev/turing/api/turing/cluster/labeller"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/database"
	openapi "github.com/caraml-dev/turing/api/turing/generated"
	"github.com/caraml-dev/turing/api/turing/internal/ref"
	"github.com/caraml-dev/turing/api/turing/models"
)

const (
	artifactFolder string = "artifact"
	// Actually this var-job=%s has a run_id appended to it, but it's ok since we use assert.Contains
	dashboardURLStringFormat string = "https://a.co/dashboard?var-project=%s&var-job=%s"
	mlpProjectName           string = "foo"
)

var (
	driverCPURequest            = "1"
	driverMemoryRequest         = "1Gi"
	executorReplica       int32 = 2
	executorCPURequest          = "1"
	executorMemoryRequest       = "1Gi"

	imageBuilderNamespace = "image"
	loggingURLFormat      = "http://www.example.com/{{.Namespace}}/{{.PodName}}"
	dashboardURLFormat    = "https://a.co/dashboard?var-project={{.Project}}&var-job={{.Job}}"
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

func createMLPService() service.MLPService {
	mlpService := &mocks.MLPService{}
	mlpService.On(
		"GetProject",
		mock.Anything,
	).Return(&mlp.Project{Id: 1, Name: mlpProjectName}, nil)
	return mlpService
}

func generateEnsemblingJobFixture(
	i int,
	ensemblerID models.ID,
	projectID models.ID,
	genExpected bool,
) *models.EnsemblingJob {
	nullableEnsemblingResources := openapi.NullableEnsemblingResources{}
	nullableEnsemblingResources.Set(&openapi.EnsemblingResources{
		DriverCpuRequest:      &driverCPURequest,
		DriverMemoryRequest:   &driverMemoryRequest,
		ExecutorReplica:       &executorReplica,
		ExecutorCpuRequest:    &executorCPURequest,
		ExecutorMemoryRequest: &executorMemoryRequest,
	})
	barString := "bar"
	envVars := []openapi.EnvVar{
		{
			Name:  "foo",
			Value: &barString,
		},
	}
	value := &models.EnsemblingJob{
		EnsemblerID:     ensemblerID,
		ProjectID:       projectID,
		EnvironmentName: "dev",
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
		value.JobConfig.Spec.Ensembler.Uri = fmt.Sprintf(
			"%s/%s/%s",
			service.SparkHomeFolder,
			artifactFolder,
			service.EnsemblerFolder,
		)
		value.EnvironmentName = "dev"
		artifactURI := fmt.Sprintf("gs://bucket/%s", artifactFolder)
		value.InfraConfig.ArtifactUri = &artifactURI
		value.InfraConfig.EnsemblerName = &service.EnsemblerFolder
		value.MonitoringURL = fmt.Sprintf(dashboardURLStringFormat, mlpProjectName, service.EnsemblerFolder)
	}

	return value
}

func TestSaveAndFindByIDEnsemblingJobIntegration(t *testing.T) {
	t.Run("success | insertion with no errors", func(t *testing.T) {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			ensemblingJobService := service.NewEnsemblingJobService(
				db,
				"dev",
				imageBuilderNamespace,
				&loggingURLFormat,
				&dashboardURLFormat,
				defaultConfigurations,
				createMLPService(),
			)

			projectID := models.ID(1)
			ensemblerID := models.ID(1000)
			ensemblingJob := generateEnsemblingJobFixture(1, ensemblerID, projectID, false)
			ensemblingJob.InfraConfig.EnsemblerName = &service.EnsemblerFolder
			err := ensemblingJobService.Save(ensemblingJob)
			assert.NoError(t, err)
			assert.NotEqual(t, models.ID(0), ensemblingJob.ID)

			found, err := ensemblingJobService.FindByID(
				ensemblingJob.ID,
				service.EnsemblingJobFindByIDOptions{ProjectID: &projectID},
			)
			assert.NoError(t, err)

			assert.NotEqual(t, found.ID, models.ID(0))
			assert.Equal(t, ensemblingJob.Name, found.Name)
			assert.Equal(t, ensemblingJob.EnsemblerID, found.EnsemblerID)
			assert.Equal(t, ensemblingJob.ProjectID, found.ProjectID)
			assert.Equal(t, ensemblingJob.EnvironmentName, found.EnvironmentName)
			assert.Equal(t, models.JobPending, found.Status)
			assert.Equal(t, ensemblingJob.InfraConfig, found.InfraConfig)
			assert.Equal(t, ensemblingJob.JobConfig, found.JobConfig)
			oldRunID := found.RunID
			assert.NotEqual(t, oldRunID, 0)

			expected := generateEnsemblingJobFixture(1, ensemblerID, projectID, true)
			assert.Contains(t, found.MonitoringURL, expected.MonitoringURL)

			// save again to test if RunID has incremented.
			ensemblingJob = generateEnsemblingJobFixture(1, ensemblerID, projectID, false)
			ensemblingJob.InfraConfig.EnsemblerName = &service.EnsemblerFolder
			err = ensemblingJobService.Save(ensemblingJob)
			assert.NoError(t, err)
			assert.NotEqual(t, models.ID(0), ensemblingJob.ID)
			assert.Equal(t, oldRunID+1, ensemblingJob.RunID)
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
				ensemblingJobService := service.NewEnsemblingJobService(
					db,
					"dev",
					imageBuilderNamespace,
					&loggingURLFormat,
					&dashboardURLFormat,
					defaultConfigurations,
					createMLPService(),
				)

				for saveCounter := 0; saveCounter < tt.saveQuantity; saveCounter++ {
					projectID := models.ID(1)
					ensemblerID := models.ID(1000)
					ensemblingJob := generateEnsemblingJobFixture(1, ensemblerID, projectID, false)
					err := ensemblingJobService.Save(ensemblingJob)
					assert.NoError(t, err)
					assert.NotEqual(t, models.ID(0), ensemblingJob.ID)
				}

				// Query pending ensembling jobs
				fetched, err := ensemblingJobService.List(
					service.EnsemblingJobListOptions{
						PaginationOptions: service.PaginationOptions{
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
			ensemblingJobService := service.NewEnsemblingJobService(
				db,
				"dev",
				imageBuilderNamespace,
				&loggingURLFormat,
				&dashboardURLFormat,
				defaultConfigurations,
				createMLPService(),
			)

			// Save job
			projectID := models.ID(1)
			ensemblerID := models.ID(1000)
			ensemblingJob := generateEnsemblingJobFixture(1, ensemblerID, projectID, false)
			ensemblingJob.InfraConfig.EnsemblerName = &service.EnsemblerFolder
			err := ensemblingJobService.Save(ensemblingJob)
			assert.NoError(t, err)
			assert.NotEqual(t, models.ID(0), ensemblingJob.ID)

			// Query pending ensembling jobs
			pageSize := 10
			retryCountLessThan := 3
			fetched, err := ensemblingJobService.List(
				service.EnsemblingJobListOptions{
					PaginationOptions: service.PaginationOptions{
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
				service.EnsemblingJobFindByIDOptions{ProjectID: &projectID},
			)
			assert.NoError(t, err)
			assert.Equal(t, models.JobFailedSubmission, found.Status)

			expected := generateEnsemblingJobFixture(1, ensemblerID, projectID, true)
			assert.Contains(t, found.MonitoringURL, expected.MonitoringURL)
		})
	})
}

func TestCreateEnsemblingJob(t *testing.T) {
	var tests = map[string]struct {
		ensembler              *models.PyFuncEnsembler
		request                *models.EnsemblingJob
		expected               *models.EnsemblingJob
		err                    error
		removeDefaultResources bool
		removeDriverCPURequest bool
	}{
		"success | name provided": {
			ensembler: &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					Name:      service.EnsemblerFolder,
					Model:     models.Model{ID: 1},
					Type:      models.EnsemblerPyFuncType,
					ProjectID: 1,
				},
				ArtifactURI: fmt.Sprintf("gs://bucket/%s", artifactFolder),
			},
			request:                generateEnsemblingJobFixture(1, 1, 1, false),
			expected:               generateEnsemblingJobFixture(1, 1, 1, true),
			removeDefaultResources: false,
			removeDriverCPURequest: false,
		},
		"success | default resources removed": {
			ensembler: &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					Name:      service.EnsemblerFolder,
					Model:     models.Model{ID: 1},
					Type:      models.EnsemblerPyFuncType,
					ProjectID: 1,
				},
				ArtifactURI: fmt.Sprintf("gs://bucket/%s", artifactFolder),
			},
			request:                generateEnsemblingJobFixture(1, 1, 1, false),
			expected:               generateEnsemblingJobFixture(1, 1, 1, true),
			removeDefaultResources: true,
			removeDriverCPURequest: false,
		},
		"success | remove 1 setting from resources": {
			ensembler: &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					Name:      service.EnsemblerFolder,
					Model:     models.Model{ID: 1},
					Type:      models.EnsemblerPyFuncType,
					ProjectID: 1,
				},
				ArtifactURI: fmt.Sprintf("gs://bucket/%s", artifactFolder),
			},
			request:                generateEnsemblingJobFixture(1, 1, 1, false),
			expected:               generateEnsemblingJobFixture(1, 1, 1, true),
			removeDefaultResources: false,
			removeDriverCPURequest: true,
		},
	}

	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				ensemblingJobService := service.NewEnsemblingJobService(
					db,
					"dev",
					imageBuilderNamespace,
					&loggingURLFormat,
					&dashboardURLFormat,
					defaultConfigurations,
					createMLPService(),
				)

				if tt.removeDefaultResources {
					tt.request.InfraConfig.Resources = openapi.NullableEnsemblingResources{}
				}

				if tt.removeDriverCPURequest {
					resources := tt.request.InfraConfig.GetResources()
					resources.DriverCpuRequest = nil
				}

				result, err := ensemblingJobService.CreateEnsemblingJob(
					tt.request,
					models.ID(1),
					tt.ensembler,
				)

				assert.Nil(t, err)
				expected := tt.expected

				assert.NotEqual(t, models.ID(0), result.ID)

				assert.Contains(t, result.Name, service.EnsemblerFolder)
				assert.Equal(t, expected.EnsemblerID, result.EnsemblerID)
				assert.Equal(t, expected.ProjectID, result.ProjectID)
				assert.Equal(t, expected.EnvironmentName, result.EnvironmentName)
				assert.Equal(t, models.JobPending, result.Status)
				assert.Contains(t, result.MonitoringURL, expected.MonitoringURL)

				assert.Equal(
					t,
					result.JobConfig.Spec.Ensembler.Uri,
					expected.JobConfig.Spec.Ensembler.Uri,
				)

				assert.Equal(t, expected.InfraConfig.ArtifactUri, result.InfraConfig.ArtifactUri)
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
						*defaultConfigurations.BatchEnsemblingJobResources.DriverCpuRequest,
						*result.InfraConfig.GetResources().DriverCpuRequest,
					)
				}
			})
		}
	})
}

func TestMarkEnsemblingJobForTermination(t *testing.T) {
	t.Run("success | delete ensembling job", func(t *testing.T) {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			ensemblingJobService := service.NewEnsemblingJobService(
				db,
				"dev",
				imageBuilderNamespace,
				&loggingURLFormat,
				&dashboardURLFormat,
				defaultConfigurations,
				createMLPService(),
			)

			// Save job
			projectID := models.ID(1)
			ensemblerID := models.ID(1000)
			ensemblingJob := generateEnsemblingJobFixture(1, ensemblerID, projectID, false)
			ensemblingJob.InfraConfig.EnsemblerName = &service.EnsemblerFolder
			err := ensemblingJobService.Save(ensemblingJob)
			assert.NoError(t, err)
			assert.NotEqual(t, models.ID(0), ensemblingJob.ID)

			// Delete job
			err = ensemblingJobService.MarkEnsemblingJobForTermination(ensemblingJob)
			assert.NoError(t, err)

			// Query back job to check if terminated
			found, err := ensemblingJobService.FindByID(
				ensemblingJob.ID,
				service.EnsemblingJobFindByIDOptions{ProjectID: &projectID},
			)
			assert.NoError(t, err)
			assert.Equal(t, models.JobTerminating, found.Status)

			expected := generateEnsemblingJobFixture(1, ensemblerID, projectID, true)
			assert.Contains(t, found.MonitoringURL, expected.MonitoringURL)
		})
	})
}

func TestPhysicalDeleteEnsemblingJob(t *testing.T) {
	t.Run("success | delete ensembling job", func(t *testing.T) {
		database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
			ensemblingJobService := service.NewEnsemblingJobService(
				db,
				"dev",
				imageBuilderNamespace,
				&loggingURLFormat,
				&dashboardURLFormat,
				defaultConfigurations,
				createMLPService(),
			)

			// Save job
			projectID := models.ID(1)
			ensemblerID := models.ID(1000)
			ensemblingJob := generateEnsemblingJobFixture(1, ensemblerID, projectID, false)
			err := ensemblingJobService.Save(ensemblingJob)
			assert.NoError(t, err)
			assert.NotEqual(t, models.ID(0), ensemblingJob.ID)

			// Delete job
			err = ensemblingJobService.Delete(ensemblingJob)
			assert.NoError(t, err)

			// Query back job to check if job is no longer there
			found, err := ensemblingJobService.FindByID(
				ensemblingJob.ID,
				service.EnsemblingJobFindByIDOptions{ProjectID: &projectID},
			)
			assert.NotNil(t, err)
			assert.Nil(t, found)
		})
	})
}

func TestGetNamespaceByComponent(t *testing.T) {
	tests := map[string]struct {
		componentType string
		project       *mlp.Project
		expected      string
	}{
		"success | image builder type": {
			componentType: batch.ImageBuilderPodType,
			project:       nil,
			expected:      imageBuilderNamespace,
		},
		"success | any other type": {
			componentType: batch.DriverPodType,
			project: &mlp.Project{
				Id:   1,
				Name: "hello",
			},
			expected: "hello",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewEnsemblingJobService(
				nil,
				"dev",
				imageBuilderNamespace,
				&loggingURLFormat,
				&dashboardURLFormat,
				defaultConfigurations,
				createMLPService(),
			)
			got := svc.GetNamespaceByComponent(tt.componentType, tt.project)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCreatePodLabelSelector(t *testing.T) {
	labeller.InitKubernetesLabeller("prefix/", "dev")
	defer labeller.InitKubernetesLabeller("", "dev")

	ensemblerName := "name"
	tests := map[string]struct {
		componentType string
		expected      []service.LabelSelector
	}{
		"success | image builder": {
			componentType: batch.ImageBuilderPodType,
			expected: []service.LabelSelector{
				{
					Key:   fmt.Sprintf("prefix/%s", labeller.AppLabel),
					Value: ensemblerName,
				},
			},
		},
		"success | driver": {
			componentType: batch.DriverPodType,
			expected: []service.LabelSelector{
				{
					Key:   fmt.Sprintf("prefix/%s", labeller.AppLabel),
					Value: ensemblerName,
				},
				{
					Key:   "spark-role",
					Value: "driver",
				},
			},
		},
		"success | executor": {
			componentType: batch.ExecutorPodType,
			expected: []service.LabelSelector{
				{
					Key:   fmt.Sprintf("prefix/%s", labeller.AppLabel),
					Value: ensemblerName,
				},
				{
					Key:   "spark-role",
					Value: "executor",
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewEnsemblingJobService(
				nil,
				"dev",
				imageBuilderNamespace,
				&loggingURLFormat,
				&dashboardURLFormat,
				defaultConfigurations,
				createMLPService(),
			)
			got := svc.CreatePodLabelSelector(ensemblerName, tt.componentType)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestFormatLoggingURL(t *testing.T) {
	tests := map[string]struct {
		ensemblerName string
		namespace     string
		componentType string
		format        string
		expected      string
	}{
		"success | nominal": {
			ensemblerName: "fooname",
			namespace:     "barspace",
			componentType: batch.ImageBuilderPodType,
			format:        "http://www.example.com/{{.Namespace}}/{{.PodName}}",
			expected:      "http://www.example.com/barspace/fooname",
		},
		"success | not initialised with format": {
			ensemblerName: "fooname",
			namespace:     "barspace",
			componentType: batch.ImageBuilderPodType,
			format:        "",
			expected:      "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			svc := service.NewEnsemblingJobService(
				nil,
				"dev",
				imageBuilderNamespace,
				&tt.format,
				&dashboardURLFormat,
				defaultConfigurations,
				createMLPService(),
			)
			got, err := svc.FormatLoggingURL(tt.ensemblerName, tt.namespace, tt.componentType)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
