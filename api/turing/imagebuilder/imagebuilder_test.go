package imagebuilder

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apibatchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/caraml-dev/mlp/api/pkg/artifact"
	"github.com/caraml-dev/mlp/api/pkg/artifact/mocks"
	"github.com/caraml-dev/turing/api/turing/cluster"
	clustermock "github.com/caraml-dev/turing/api/turing/cluster/mocks"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/models"
)

var (
	timeout, _            = time.ParseDuration("10s")
	testArtifactURISuffix = "://bucket-name/mlflow/3069/e130c40703ee424da97b9ecee7b874b7/artifacts"
	testArtifactURI       = "gs://bucket-name/mlflow/3069/e130c40703ee424da97b9ecee7b874b7/artifacts"
	testArtifactGsutilURL = &artifact.URL{
		Bucket: "bucket-name",
		Object: "mlflow/3069/e130c40703ee424da97b9ecee7b874b7/artifacts",
	}
	testCondaEnvContent = `dependencies:
- python=3.9.*
- pip:
  - mlflow`
	testCondaEnvURLSuffix = testArtifactURISuffix + "/ensembler/conda.yaml"
)

func getHashedModelDependenciesURL() string {
	hash := sha256.New()
	hash.Write([]byte(testCondaEnvContent))
	hashEnv := hash.Sum(nil)

	return fmt.Sprintf("gs://%s/turing/model_dependencies/%x", testArtifactGsutilURL.Bucket, hashEnv)
}

const (
	projectName                          = "test-project"
	modelName                            = "mymodel"
	modelVersion                         = models.ID(1)
	runID                                = "abc123"
	dockerRegistry                       = "ghcr.io"
	pyFuncEnsemblerJobDockerfilePath     = "engines/pyfunc-ensembler-job/app.Dockerfile"
	pyFuncEnsemblerServiceDockerfilePath = "engines/pyfunc-ensembler-service/app.Dockerfile"
	buildContext                         = "git://github.com/caraml-dev/turing.git#refs/heads/master"
	pyFuncEnsemblerJobBaseImageRef       = "ghcr.io/caraml-dev/turing/pyfunc-ensembler-job:v0.0.0-build.154-e108820"
	pyFuncEnsemblerServiceBaseImageRef   = "ghcr.io/caraml-dev/turing/pyfunc-ensembler-service:v0.0.0-build.154-e102280"
	buildNamespace                       = "mlp"
	ensemblerFolder                      = "ensembler"
)

func TestBuildPyFuncEnsemblerJobImage(t *testing.T) {
	imageBuildingConfig := config.ImageBuildingConfig{
		BuildNamespace:       buildNamespace,
		BuildTimeoutDuration: timeout,
		DestinationRegistry:  dockerRegistry,
		BaseImage:            pyFuncEnsemblerJobBaseImageRef,
		KanikoConfig: config.KanikoConfig{
			BuildContextURI:    buildContext,
			DockerfileFilePath: pyFuncEnsemblerJobDockerfilePath,
			Image:              "gcr.io/kaniko-project/executor",
			ImageVersion:       "v1.5.2",
			ResourceRequestsLimits: config.ResourceRequestsLimits{
				Requests: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
				Limits: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
			},
		},
	}
	tests := map[string]struct {
		name                string
		expected            string
		projectName         string
		modelName           string
		modelID             models.ID
		versionID           string
		inputDependencies   []string
		namespace           string
		imageBuildingConfig config.ImageBuildingConfig
		buildLabels         map[string]string
		clusterController   func() cluster.Controller
		ensemblerFolder     string
		imageTag            string
		artifactServiceMock func(*mocks.Service)
	}{
		"success | no existing job": {
			expected:    fmt.Sprintf("%s/%s/ensembler-jobs/%s:%s", dockerRegistry, projectName, modelName, runID),
			projectName: projectName,
			modelName:   modelName,
			modelID:     modelVersion,
			versionID:   runID,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				// First time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					k8serrors.NewNotFound(
						schema.GroupResource{},
						fmt.Sprintf("batch-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
					),
				).Once()

				// Second time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Succeeded: 1,
						},
					},
					nil,
				).Once()

				ctlr.On(
					"CreateJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			buildLabels: map[string]string{
				"gojek.io/team": "dsp",
			},
			imageBuildingConfig: imageBuildingConfig,
			ensemblerFolder:     ensemblerFolder,
			imageTag:            "3.7.*",
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
		"success: existing job is running": {
			expected:    fmt.Sprintf("%s/%s/ensembler-jobs/%s:%s", dockerRegistry, projectName, modelName, runID),
			projectName: projectName,
			modelName:   modelName,
			modelID:     modelVersion,
			versionID:   runID,
			buildLabels: map[string]string{
				"gojek.io/team": "dsp",
			},
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				// First time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
						},
					},
					nil,
				).Once()

				// Second time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Succeeded: 1,
						},
					},
					nil,
				).Once()

				ctlr.On(
					"CreateJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			imageBuildingConfig: imageBuildingConfig,
			ensemblerFolder:     ensemblerFolder,
			imageTag:            "3.7.*",
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
		"success: existing job failed": {
			expected:    fmt.Sprintf("%s/%s/ensembler-jobs/%s:%s", dockerRegistry, projectName, modelName, runID),
			projectName: projectName,
			modelName:   modelName,
			modelID:     modelVersion,
			versionID:   runID,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				// First time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
						},
						Status: apibatchv1.JobStatus{
							Failed: 1,
						},
					},
					nil,
				).Once()

				// Second time GetJob is called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					k8serrors.NewNotFound(
						schema.GroupResource{},
						fmt.Sprintf("batch-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
					),
				).Once()

				// Third time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Succeeded: 1,
						},
					},
					nil,
				).Once()

				ctlr.On(
					"DeleteJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).Once()

				ctlr.On(
					"CreateJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			buildLabels: map[string]string{
				"gojek.io/team": "dsp",
			},
			imageBuildingConfig: imageBuildingConfig,
			ensemblerFolder:     ensemblerFolder,
			imageTag:            "3.7.*",
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			artifactServiceMock := &mocks.Service{}
			tt.artifactServiceMock(artifactServiceMock)

			clusterController := tt.clusterController()

			ib, err := NewEnsemblerJobImageBuilder(
				clusterController,
				tt.imageBuildingConfig,
				googleCloudStorageArtifactServiceType,
				artifactServiceMock,
			)
			assert.Nil(t, err)

			buildImageRequest := BuildImageRequest{
				tt.projectName,
				tt.modelName,
				tt.modelID,
				tt.versionID,
				testArtifactURI,
				tt.buildLabels,
				tt.ensemblerFolder,
			}
			actual, err := ib.BuildImage(buildImageRequest)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestBuildPyFuncEnsemblerServiceImage(t *testing.T) {
	imageBuildingConfig := config.ImageBuildingConfig{
		BuildNamespace:       buildNamespace,
		BuildTimeoutDuration: timeout,
		DestinationRegistry:  dockerRegistry,
		BaseImage:            pyFuncEnsemblerServiceBaseImageRef,
		KanikoConfig: config.KanikoConfig{
			BuildContextURI:    buildContext,
			DockerfileFilePath: pyFuncEnsemblerServiceDockerfilePath,
			Image:              "gcr.io/kaniko-project/executor",
			ImageVersion:       "v1.5.2",
			ResourceRequestsLimits: config.ResourceRequestsLimits{
				Requests: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
				Limits: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
			},
		},
	}
	tests := map[string]struct {
		name                       string
		expectedImage              string
		expectedImageBuildingError string
		projectName                string
		modelName                  string
		modelID                    models.ID
		versionID                  string
		inputDependencies          []string
		namespace                  string
		imageBuildingConfig        config.ImageBuildingConfig
		buildLabels                map[string]string
		clusterController          func() cluster.Controller
		ensemblerFolder            string
		imageTag                   string
		artifactServiceMock        func(*mocks.Service)
	}{
		"success | no existing job": {
			expectedImage: fmt.Sprintf("%s/%s/ensembler-services/%s:%s", dockerRegistry, projectName, modelName, runID),
			projectName:   projectName,
			modelName:     modelName,
			modelID:       modelVersion,
			versionID:     runID,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				// First time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					k8serrors.NewNotFound(
						schema.GroupResource{},
						fmt.Sprintf("service-builder-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
					),
				).Once()

				// Second time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Succeeded: 1,
						},
					},
					nil,
				).Once()

				ctlr.On(
					"CreateJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("service-builder-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			buildLabels: map[string]string{
				"gojek.io/team": "dsp",
			},
			imageBuildingConfig: imageBuildingConfig,
			ensemblerFolder:     ensemblerFolder,
			imageTag:            "3.7.*",
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
		"success: existing job is running": {
			expectedImage: fmt.Sprintf("%s/%s/ensembler-services/%s:%s", dockerRegistry, projectName, modelName, runID),
			projectName:   projectName,
			modelName:     modelName,
			modelID:       modelVersion,
			versionID:     runID,
			buildLabels: map[string]string{
				"gojek.io/team": "dsp",
			},
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				// First time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("service-builder-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
						},
					},
					nil,
				).Once()

				// Second time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Succeeded: 1,
						},
					},
					nil,
				).Once()

				ctlr.On(
					"CreateJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("service-builder-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			imageBuildingConfig: imageBuildingConfig,
			ensemblerFolder:     ensemblerFolder,
			imageTag:            "3.7.*",
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
		"success: existing job failed": {
			expectedImage: fmt.Sprintf("%s/%s/ensembler-services/%s:%s", dockerRegistry, projectName, modelName, runID),
			projectName:   projectName,
			modelName:     modelName,
			modelID:       modelVersion,
			versionID:     runID,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				// First time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("service-builder-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
						},
						Status: apibatchv1.JobStatus{
							Failed: 1,
						},
					},
					nil,
				).Once()

				// Second time GetJob is called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					k8serrors.NewNotFound(
						schema.GroupResource{},
						fmt.Sprintf("service-builder-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
					),
				).Once()

				// Third time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Succeeded: 1,
						},
					},
					nil,
				).Once()

				ctlr.On(
					"DeleteJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil).Once()

				ctlr.On(
					"CreateJob",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("service-builder-%s-%s-%d-%s", projectName, modelName, modelVersion, runID[:5]),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			buildLabels: map[string]string{
				"gojek.io/team": "dsp",
			},
			imageBuildingConfig: imageBuildingConfig,
			ensemblerFolder:     ensemblerFolder,
			imageTag:            "3.7.*",
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			artifactServiceMock := &mocks.Service{}
			tt.artifactServiceMock(artifactServiceMock)

			clusterController := tt.clusterController()

			ib, err := NewEnsemblerServiceImageBuilder(
				clusterController,
				tt.imageBuildingConfig,
				googleCloudStorageArtifactServiceType,
				artifactServiceMock,
			)
			assert.Nil(t, err)

			buildImageRequest := BuildImageRequest{
				tt.projectName,
				tt.modelName,
				tt.modelID,
				tt.versionID,
				testArtifactURI,
				tt.buildLabels,
				tt.ensemblerFolder,
			}
			actual, err := ib.BuildImage(buildImageRequest)
			if tt.expectedImageBuildingError == "" {
				assert.Nil(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedImageBuildingError)
			}

			assert.Equal(t, tt.expectedImage, actual)
		})
	}
}

func TestParseResources(t *testing.T) {
	tests := map[string]struct {
		name                   string
		expected               bool
		resourceRequestsLimits config.ResourceRequestsLimits
	}{
		"success | parsable": {
			expected: true,
			resourceRequestsLimits: config.ResourceRequestsLimits{
				Requests: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
				Limits: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
			},
		},
		"failure | cpu_request_error": {
			expected: false,
			resourceRequestsLimits: config.ResourceRequestsLimits{
				Requests: config.Resource{
					CPU:    "Chicken",
					Memory: "2Gi",
				},
				Limits: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
			},
		},
		"failure | cpu_limit_error": {
			expected: false,
			resourceRequestsLimits: config.ResourceRequestsLimits{
				Requests: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
				Limits: config.Resource{
					CPU:    "Vegetable",
					Memory: "2Gi",
				},
			},
		},
		"failure | memory_request_error": {
			expected: false,
			resourceRequestsLimits: config.ResourceRequestsLimits{
				Requests: config.Resource{
					CPU:    "1",
					Memory: "Brains",
				},
				Limits: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
			},
		},
		"failure | memory_limit_error": {
			expected: false,
			resourceRequestsLimits: config.ResourceRequestsLimits{
				Requests: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
				Limits: config.Resource{
					CPU:    "1",
					Memory: "DownloadMoreRam",
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := checkParseResources(tt.resourceRequestsLimits)
			if tt.expected == true {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}

func TestGetEnsemblerJobImageBuildingJobStatus(t *testing.T) {
	imageBuildingConfig := config.ImageBuildingConfig{
		BuildNamespace:       buildNamespace,
		BuildTimeoutDuration: timeout,
		DestinationRegistry:  dockerRegistry,
		BaseImage:            pyFuncEnsemblerJobBaseImageRef,
		KanikoConfig: config.KanikoConfig{
			BuildContextURI:    buildContext,
			DockerfileFilePath: pyFuncEnsemblerJobDockerfilePath,
			Image:              "gcr.io/kaniko-project/executor",
			ImageVersion:       "v1.5.2",
			ResourceRequestsLimits: config.ResourceRequestsLimits{
				Requests: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
				Limits: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
			},
		},
	}
	tests := map[string]struct {
		clusterController   func() cluster.Controller
		imageBuildingConfig config.ImageBuildingConfig
		hasErr              bool
		expected            JobStatus
		artifactServiceMock func(*mocks.Service)
	}{
		"success | active": {
			imageBuildingConfig: imageBuildingConfig,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Active: 1,
						},
					},
					nil,
				)
				return ctlr
			},
			hasErr: false,
			expected: JobStatus{
				State: JobStateActive,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
		"success | succeeded": {
			imageBuildingConfig: imageBuildingConfig,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Succeeded: 1,
						},
					},
					nil,
				)
				return ctlr
			},
			hasErr: false,
			expected: JobStatus{
				State: JobStateSucceeded,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
		"success | Failed": {
			imageBuildingConfig: imageBuildingConfig,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "batch-test-project-mymodel-1-abc12",
							Namespace: imageBuildingConfig.BuildNamespace,
						},
						Status: apibatchv1.JobStatus{
							Failed: 1,
							Conditions: []apibatchv1.JobCondition{
								{
									LastProbeTime: metav1.Date(2024, 4, 29, 0o0, 0o0, 0o0, 0, time.UTC),
									Type:          apibatchv1.JobFailed,
									Reason:        "BackoffLimitExceeded",
									Message:       "Job has reached the specified backoff limit",
								},
							},
						},
					},
					nil,
				)
				ctlr.On("ListPods", mock.Anything, mock.Anything, mock.Anything).Return(
					&v1.PodList{
						Items: []v1.Pod{
							{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "batch-test-project-mymodel-1-abc12-123",
									Namespace: imageBuildingConfig.BuildNamespace,
									Labels: map[string]string{
										"job-name": "batch-test-project-mymodel-1-abc12",
									},
								},
								Status: v1.PodStatus{
									Phase: v1.PodFailed,
									ContainerStatuses: []v1.ContainerStatus{
										{
											Name: "kaniko-builder",
											State: v1.ContainerState{
												Terminated: &v1.ContainerStateTerminated{
													ExitCode: 1,
													Reason:   "Error",
													Message:  "CondaEnvException: Pip failed",
												},
											},
										},
									},
								},
							},
						},
					},
					nil,
				)
				return ctlr
			},
			hasErr: false,
			expected: JobStatus{
				State: JobStateFailed,
				Message: `Error

Job conditions:
┌───────────────────────────────┬────────┬──────────────────────┬─────────────────────────────────────────────┐
│ TIMESTAMP                     │ TYPE   │ REASON               │ MESSAGE                                     │
├───────────────────────────────┼────────┼──────────────────────┼─────────────────────────────────────────────┤
│ Mon, 29 Apr 2024 00:00:00 UTC │ Failed │ BackoffLimitExceeded │ Job has reached the specified backoff limit │
└───────────────────────────────┴────────┴──────────────────────┴─────────────────────────────────────────────┘

Pod container status:
┌────────────────┬────────────┬───────────┬────────┐
│ CONTAINER NAME │ STATUS     │ EXIT CODE │ REASON │
├────────────────┼────────────┼───────────┼────────┤
│ kaniko-builder │ Terminated │ 1         │ Error  │
└────────────────┴────────────┴───────────┴────────┘

Pod last termination message:
CondaEnvException: Pip failed`,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
		"success | Unknown": {
			imageBuildingConfig: imageBuildingConfig,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{},
					nil,
				)
				ctlr.On("ListPods", mock.Anything, mock.Anything, mock.Anything).Return(
					&v1.PodList{
						Items: []v1.Pod{},
					},
					nil,
				)
				return ctlr
			},
			hasErr: false,
			expected: JobStatus{
				State: JobStateUnknown,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
		"failure | Unknown": {
			imageBuildingConfig: imageBuildingConfig,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
					fmt.Errorf("hello"),
				)
				return ctlr
			},
			hasErr: true,
			expected: JobStatus{
				State:   JobStateUnknown,
				Message: "hello",
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			artifactServiceMock := &mocks.Service{}
			tt.artifactServiceMock(artifactServiceMock)

			clusterController := tt.clusterController()
			ib, _ := NewEnsemblerJobImageBuilder(
				clusterController,
				tt.imageBuildingConfig,
				googleCloudStorageArtifactServiceType,
				artifactServiceMock,
			)
			status := ib.GetImageBuildingJobStatus(projectName, modelName, models.ID(1), runID)
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestGetEnsemblerServiceImageBuildingJobStatus(t *testing.T) {
	imageBuildingConfig := config.ImageBuildingConfig{
		BuildNamespace:       buildNamespace,
		BuildTimeoutDuration: timeout,
		DestinationRegistry:  dockerRegistry,
		BaseImage:            pyFuncEnsemblerJobBaseImageRef,
		KanikoConfig: config.KanikoConfig{
			BuildContextURI:    buildContext,
			DockerfileFilePath: pyFuncEnsemblerServiceDockerfilePath,
			Image:              "gcr.io/kaniko-project/executor",
			ImageVersion:       "v1.5.2",
			ResourceRequestsLimits: config.ResourceRequestsLimits{
				Requests: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
				Limits: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
			},
		},
	}
	tests := map[string]struct {
		clusterController   func() cluster.Controller
		imageBuildingConfig config.ImageBuildingConfig
		hasErr              bool
		expected            JobStatus
		artifactServiceMock func(*mocks.Service)
	}{
		"success | active": {
			imageBuildingConfig: imageBuildingConfig,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Active: 1,
						},
					},
					nil,
				)
				return ctlr
			},
			hasErr: false,
			expected: JobStatus{
				State: JobStateActive,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
		"success | succeeded": {
			imageBuildingConfig: imageBuildingConfig,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Succeeded: 1,
						},
					},
					nil,
				)
				return ctlr
			},
			hasErr: false,
			expected: JobStatus{
				State: JobStateSucceeded,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
		"success | Failed": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImage:            pyFuncEnsemblerJobBaseImageRef,
				KanikoConfig: config.KanikoConfig{
					BuildContextURI:    buildContext,
					DockerfileFilePath: pyFuncEnsemblerServiceDockerfilePath,
					Image:              "gcr.io/kaniko-project/executor",
					ImageVersion:       "v1.5.2",
					ResourceRequestsLimits: config.ResourceRequestsLimits{
						Requests: config.Resource{
							CPU:    "1",
							Memory: "2Gi",
						},
						Limits: config.Resource{
							CPU:    "1",
							Memory: "2Gi",
						},
					},
				},
			},
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Failed: 1,
						},
					},
					nil,
				)
				ctlr.On("ListPods", mock.Anything, mock.Anything, mock.Anything).Return(
					&v1.PodList{
						Items: []v1.Pod{},
					},
					nil,
				)
				return ctlr
			},
			hasErr: false,
			expected: JobStatus{
				State: JobStateFailed,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
		"success | Unknown": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImage:            pyFuncEnsemblerJobBaseImageRef,
				KanikoConfig: config.KanikoConfig{
					BuildContextURI:    buildContext,
					DockerfileFilePath: pyFuncEnsemblerServiceDockerfilePath,
					Image:              "gcr.io/kaniko-project/executor",
					ImageVersion:       "v1.5.2",
					ResourceRequestsLimits: config.ResourceRequestsLimits{
						Requests: config.Resource{
							CPU:    "1",
							Memory: "2Gi",
						},
						Limits: config.Resource{
							CPU:    "1",
							Memory: "2Gi",
						},
					},
				},
			},
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{},
					nil,
				)
				ctlr.On("ListPods", mock.Anything, mock.Anything, mock.Anything).Return(
					&v1.PodList{
						Items: []v1.Pod{},
					},
					nil,
				)
				return ctlr
			},
			hasErr: false,
			expected: JobStatus{
				State: JobStateUnknown,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
		"failure | Unknown": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImage:            pyFuncEnsemblerJobBaseImageRef,
				KanikoConfig: config.KanikoConfig{
					BuildContextURI:    buildContext,
					DockerfileFilePath: pyFuncEnsemblerServiceDockerfilePath,
					Image:              "gcr.io/kaniko-project/executor",
					ImageVersion:       "v1.5.2",
					ResourceRequestsLimits: config.ResourceRequestsLimits{
						Requests: config.Resource{
							CPU:    "1",
							Memory: "2Gi",
						},
						Limits: config.Resource{
							CPU:    "1",
							Memory: "2Gi",
						},
					},
				},
			},
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
					fmt.Errorf("hello"),
				)
				return ctlr
			},
			hasErr: true,
			expected: JobStatus{
				State:   JobStateUnknown,
				Message: "hello",
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			artifactServiceMock := &mocks.Service{}
			tt.artifactServiceMock(artifactServiceMock)

			clusterController := tt.clusterController()
			ib, _ := NewEnsemblerServiceImageBuilder(
				clusterController,
				tt.imageBuildingConfig,
				googleCloudStorageArtifactServiceType,
				artifactServiceMock,
			)
			status := ib.GetImageBuildingJobStatus("", "", models.ID(1), runID)
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestDeleteEnsemblerJobImageBuildingJob(t *testing.T) {
	imageBuildingConfig := config.ImageBuildingConfig{
		BuildNamespace:       buildNamespace,
		BuildTimeoutDuration: timeout,
		DestinationRegistry:  dockerRegistry,
		BaseImage:            pyFuncEnsemblerJobBaseImageRef,
		KanikoConfig: config.KanikoConfig{
			BuildContextURI:    buildContext,
			DockerfileFilePath: pyFuncEnsemblerServiceDockerfilePath,
			Image:              "gcr.io/kaniko-project/executor",
			ImageVersion:       "v1.5.2",
			ResourceRequestsLimits: config.ResourceRequestsLimits{
				Requests: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
				Limits: config.Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
			},
		},
	}
	tests := map[string]struct {
		clusterController   func() cluster.Controller
		imageBuildingConfig config.ImageBuildingConfig
		hasErr              bool
		artifactServiceMock func(*mocks.Service)
	}{
		"success | no error": {
			imageBuildingConfig: imageBuildingConfig,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: "bicycle",
						},
					},
					nil,
				)
				ctlr.On("DeleteJob", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				return ctlr
			},
			hasErr: false,
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			artifactServiceMock := &mocks.Service{}
			tt.artifactServiceMock(artifactServiceMock)

			clusterController := tt.clusterController()
			ib, _ := NewEnsemblerJobImageBuilder(
				clusterController,
				tt.imageBuildingConfig,
				googleCloudStorageArtifactServiceType,
				artifactServiceMock,
			)
			err := ib.DeleteImageBuildingJob("", "", models.ID(1), runID)

			if tt.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestDeleteEnsemblerServiceImageBuildingJob(t *testing.T) {
	tests := map[string]struct {
		clusterController   func() cluster.Controller
		imageBuildingConfig config.ImageBuildingConfig
		hasErr              bool
		artifactServiceMock func(*mocks.Service)
	}{
		"success | no error": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImage:            pyFuncEnsemblerJobBaseImageRef,
				KanikoConfig: config.KanikoConfig{
					BuildContextURI:    buildContext,
					DockerfileFilePath: pyFuncEnsemblerServiceDockerfilePath,
					Image:              "gcr.io/kaniko-project/executor",
					ImageVersion:       "v1.5.2",
					ResourceRequestsLimits: config.ResourceRequestsLimits{
						Requests: config.Resource{
							CPU:    "1",
							Memory: "2Gi",
						},
						Limits: config.Resource{
							CPU:    "1",
							Memory: "2Gi",
						},
					},
				},
			},
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: "bicycle",
						},
					},
					nil,
				)
				ctlr.On("DeleteJob", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				return ctlr
			},
			hasErr: false,
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				modelDependenciesURL := getHashedModelDependenciesURL()
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			artifactServiceMock := &mocks.Service{}
			tt.artifactServiceMock(artifactServiceMock)

			clusterController := tt.clusterController()
			ib, _ := NewEnsemblerJobImageBuilder(
				clusterController,
				tt.imageBuildingConfig,
				googleCloudStorageArtifactServiceType,
				artifactServiceMock,
			)
			err := ib.DeleteImageBuildingJob("", "", models.ID(1), runID)

			if tt.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_imageBuilder_getHashedModelDependenciesURL(t *testing.T) {
	modelDependenciesURL := getHashedModelDependenciesURL()
	type args struct {
		ctx         context.Context
		artifactURI string
	}
	tests := []struct {
		name                string
		args                args
		artifactServiceMock func(*mocks.Service)
		want                string
		wantErr             bool
	}{
		{
			name: "hash dependencies is already exist",
			args: args{
				ctx:         context.Background(),
				artifactURI: testArtifactURI,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return([]byte(testCondaEnvContent), nil)
			},
			want:    modelDependenciesURL,
			wantErr: false,
		},
		{
			name: "hash dependencies is not exist yet",
			args: args{
				ctx:         context.Background(),
				artifactURI: testArtifactURI,
			},
			artifactServiceMock: func(artifactServiceMock *mocks.Service) {
				artifactServiceMock.On("ParseURL", fmt.Sprintf("gs%s", testArtifactURISuffix)).Return(testArtifactGsutilURL, nil)
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("GetURLScheme").Return("gs")
				artifactServiceMock.On("ReadArtifact", mock.Anything, fmt.Sprintf("gs%s", testCondaEnvURLSuffix)).
					Return([]byte(testCondaEnvContent), nil)
				artifactServiceMock.On("ReadArtifact", mock.Anything, modelDependenciesURL).Return(nil, artifact.ErrObjectNotExist)
				artifactServiceMock.On("WriteArtifact", mock.Anything, modelDependenciesURL, []byte(testCondaEnvContent)).
					Return(nil)
			},
			want:    modelDependenciesURL,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			artifactServiceMock := &mocks.Service{}
			tt.artifactServiceMock(artifactServiceMock)

			c := &imageBuilder{
				artifactService: artifactServiceMock,
			}

			got, err := c.getHashedModelDependenciesURL(tt.args.ctx, tt.args.artifactURI)
			if (err != nil) != tt.wantErr {
				t.Errorf("imageBuilder.getHashedModelDependenciesURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("imageBuilder.getHashedModelDependenciesURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
