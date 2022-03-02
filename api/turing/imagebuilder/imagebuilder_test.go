package imagebuilder

import (
	"fmt"
	"testing"
	"time"

	"github.com/gojek/turing/api/turing/cluster"
	clustermock "github.com/gojek/turing/api/turing/cluster/mocks"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apibatchv1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	timeout, _ = time.ParseDuration("10s")
)

const (
	projectName                          = "test-project"
	modelName                            = "mymodel"
	modelVersion                         = models.ID(1)
	dockerRegistry                       = "ghcr.io"
	artifactURI                          = "gs://bucket/ensembler"
	pyFuncEnsemblerJobDockerfilePath     = "engines/pyfunc-ensembler-job/app.Dockerfile"
	pyFuncEnsemblerServiceDockerfilePath = "engines/pyfunc-ensembler-service/app.Dockerfile"
	buildContext                         = "git://github.com/gojek/turing.git#refs/heads/master"
	pyFuncEnsemblerJobBaseImageRef       = "ghcr.io/gojek/turing/pyfunc-ensembler-job:v0.0.0-build.154-e108820"
	pyFuncEnsemblerServiceBaseImageRef   = "ghcr.io/gojek/turing/pyfunc-ensembler-service:v0.0.0-build.154-e102280"
	buildNamespace                       = "mlp"
)

func TestBuildPyFuncEnsemblerJobImage(t *testing.T) {
	var tests = map[string]struct {
		name                string
		expected            string
		projectName         string
		modelName           string
		artifactURI         string
		versionID           models.ID
		inputDependencies   []string
		namespace           string
		imageBuildingConfig config.ImageBuildingConfig
		buildLabels         map[string]string
		clusterController   func() cluster.Controller
		ensemblerFolder     string
	}{
		"success | no existing job": {
			expected:    fmt.Sprintf("%s/%s-%s-job:%d", dockerRegistry, projectName, modelName, modelVersion),
			projectName: projectName,
			modelName:   modelName,
			versionID:   modelVersion,
			artifactURI: artifactURI,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				// First time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					k8serrors.NewNotFound(
						schema.GroupResource{},
						fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, modelVersion),
					),
				).Once()

				// Second time it's called
				ctlr.On(
					"GetJob",
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
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-builder-%s-%s-%d", projectName, modelName, modelVersion),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			buildLabels: map[string]string{
				"gojek.io/team": "dsp",
			},
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerJobBaseImageRef,
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
			},
			ensemblerFolder: "ensembler",
		},
		"success: existing job is running": {
			expected:    fmt.Sprintf("%s/%s-%s-job:%d", dockerRegistry, projectName, modelName, modelVersion),
			projectName: projectName,
			modelName:   modelName,
			versionID:   modelVersion,
			artifactURI: artifactURI,
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
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-builder-%s-%s-%d", projectName, modelName, modelVersion),
						},
					},
					nil,
				).Once()

				// Second time it's called
				ctlr.On(
					"GetJob",
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
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-builder-%s-%s-%d", projectName, modelName, modelVersion),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerJobBaseImageRef,
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
			},
			ensemblerFolder: "ensembler",
		},
		"success: existing job failed": {
			expected:    fmt.Sprintf("%s/%s-%s-job:%d", dockerRegistry, projectName, modelName, modelVersion),
			projectName: projectName,
			modelName:   modelName,
			versionID:   modelVersion,
			artifactURI: artifactURI,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				// First time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-builder-%s-%s-%d", projectName, modelName, modelVersion),
						},
						Status: apibatchv1.JobStatus{
							Failed: 1,
						},
					},
					nil,
				).Once()

				// Second time it's called
				ctlr.On(
					"GetJob",
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
				).Return(nil).Once()

				ctlr.On(
					"CreateJob",
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-builder-%s-%s-%d", projectName, modelName, modelVersion),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			buildLabels: map[string]string{
				"gojek.io/team": "dsp",
			},
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerJobBaseImageRef,
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
			},
			ensemblerFolder: "ensembler",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clusterController := tt.clusterController()

			ib, err := NewEnsemblerJobImageBuilder(clusterController, tt.imageBuildingConfig)
			assert.Nil(t, err)

			buildImageRequest := BuildImageRequest{
				tt.projectName,
				tt.modelName,
				tt.versionID,
				tt.artifactURI,
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
	var tests = map[string]struct {
		name                string
		expected            string
		projectName         string
		modelName           string
		artifactURI         string
		versionID           models.ID
		inputDependencies   []string
		namespace           string
		imageBuildingConfig config.ImageBuildingConfig
		buildLabels         map[string]string
		clusterController   func() cluster.Controller
		ensemblerFolder     string
	}{
		"success | no existing job": {
			expected:    fmt.Sprintf("%s/%s-%s-service:%d", dockerRegistry, projectName, modelName, modelVersion),
			projectName: projectName,
			modelName:   modelName,
			versionID:   modelVersion,
			artifactURI: artifactURI,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				// First time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
				).Return(
					nil,
					k8serrors.NewNotFound(
						schema.GroupResource{},
						fmt.Sprintf("service-builder-%s-%s-%d", projectName, modelName, modelVersion),
					),
				).Once()

				// Second time it's called
				ctlr.On(
					"GetJob",
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
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("service-builder-%s-%s-%d", projectName, modelName, modelVersion),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			buildLabels: map[string]string{
				"gojek.io/team": "dsp",
			},
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerServiceBaseImageRef,
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
			ensemblerFolder: "ensembler",
		},
		"success: existing job is running": {
			expected:    fmt.Sprintf("%s/%s-%s-service:%d", dockerRegistry, projectName, modelName, modelVersion),
			projectName: projectName,
			modelName:   modelName,
			versionID:   modelVersion,
			artifactURI: artifactURI,
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
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("service-builder-%s-%s-%d", projectName, modelName, modelVersion),
						},
					},
					nil,
				).Once()

				// Second time it's called
				ctlr.On(
					"GetJob",
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
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("service-builder-%s-%s-%d", projectName, modelName, modelVersion),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerServiceBaseImageRef,
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
			ensemblerFolder: "ensembler",
		},
		"success: existing job failed": {
			expected:    fmt.Sprintf("%s/%s-%s-service:%d", dockerRegistry, projectName, modelName, modelVersion),
			projectName: projectName,
			modelName:   modelName,
			versionID:   modelVersion,
			artifactURI: artifactURI,
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				// First time it's called
				ctlr.On(
					"GetJob",
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("service-builder-%s-%s-%d", projectName, modelName, modelVersion),
						},
						Status: apibatchv1.JobStatus{
							Failed: 1,
						},
					},
					nil,
				).Once()

				// Second time it's called
				ctlr.On(
					"GetJob",
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
				).Return(nil).Once()

				ctlr.On(
					"CreateJob",
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("service-builder-%s-%s-%d", projectName, modelName, modelVersion),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			buildLabels: map[string]string{
				"gojek.io/team": "dsp",
			},
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerServiceBaseImageRef,
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
			ensemblerFolder: "ensembler",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clusterController := tt.clusterController()

			ib, err := NewEnsemblerServiceImageBuilder(clusterController, tt.imageBuildingConfig)
			assert.Nil(t, err)

			buildImageRequest := BuildImageRequest{
				tt.projectName,
				tt.modelName,
				tt.versionID,
				tt.artifactURI,
				tt.buildLabels,
				tt.ensemblerFolder,
			}
			actual, err := ib.BuildImage(buildImageRequest)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestParseResources(t *testing.T) {
	var tests = map[string]struct {
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
	tests := map[string]struct {
		clusterController   func() cluster.Controller
		imageBuildingConfig config.ImageBuildingConfig
		hasErr              bool
		expected            JobStatus
	}{
		"success | active": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerJobBaseImageRef,
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
			},
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Active: 1,
						},
					},
					nil,
				)
				return ctlr
			},
			hasErr:   false,
			expected: JobStatusActive,
		},
		"success | succeeded": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerJobBaseImageRef,
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
			},
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Succeeded: 1,
						},
					},
					nil,
				)
				return ctlr
			},
			hasErr:   false,
			expected: JobStatusSucceeded,
		},
		"success | Failed": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerJobBaseImageRef,
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
			},
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Failed: 1,
						},
					},
					nil,
				)
				return ctlr
			},
			hasErr:   false,
			expected: JobStatusFailed,
		},
		"success | Unknown": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerJobBaseImageRef,
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
			},
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{},
					nil,
				)
				return ctlr
			},
			hasErr:   false,
			expected: JobStatusUnknown,
		},
		"failure | Unknown": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerJobBaseImageRef,
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
			},
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything).Return(
					nil,
					fmt.Errorf("hello"),
				)
				return ctlr
			},
			hasErr:   true,
			expected: JobStatusUnknown,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clusterController := tt.clusterController()
			ib, _ := NewEnsemblerJobImageBuilder(clusterController, tt.imageBuildingConfig)
			status, err := ib.GetImageBuildingJobStatus("", "", models.ID(1))

			if tt.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestGetEnsemblerServiceImageBuildingJobStatus(t *testing.T) {
	tests := map[string]struct {
		clusterController   func() cluster.Controller
		imageBuildingConfig config.ImageBuildingConfig
		hasErr              bool
		expected            JobStatus
	}{
		"success | active": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerServiceBaseImageRef,
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
				ctlr.On("GetJob", mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Active: 1,
						},
					},
					nil,
				)
				return ctlr
			},
			hasErr:   false,
			expected: JobStatusActive,
		},
		"success | succeeded": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerServiceBaseImageRef,
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
				ctlr.On("GetJob", mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Succeeded: 1,
						},
					},
					nil,
				)
				return ctlr
			},
			hasErr:   false,
			expected: JobStatusSucceeded,
		},
		"success | Failed": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerServiceBaseImageRef,
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
				ctlr.On("GetJob", mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						Status: apibatchv1.JobStatus{
							Failed: 1,
						},
					},
					nil,
				)
				return ctlr
			},
			hasErr:   false,
			expected: JobStatusFailed,
		},
		"success | Unknown": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerServiceBaseImageRef,
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
				ctlr.On("GetJob", mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{},
					nil,
				)
				return ctlr
			},
			hasErr:   false,
			expected: JobStatusUnknown,
		},
		"failure | Unknown": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerServiceBaseImageRef,
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
				ctlr.On("GetJob", mock.Anything, mock.Anything).Return(
					nil,
					fmt.Errorf("hello"),
				)
				return ctlr
			},
			hasErr:   true,
			expected: JobStatusUnknown,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clusterController := tt.clusterController()
			ib, _ := NewEnsemblerServiceImageBuilder(clusterController, tt.imageBuildingConfig)
			status, err := ib.GetImageBuildingJobStatus("", "", models.ID(1))

			if tt.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestDeleteEnsemblerJobImageBuildingJob(t *testing.T) {
	tests := map[string]struct {
		clusterController   func() cluster.Controller
		imageBuildingConfig config.ImageBuildingConfig
		hasErr              bool
	}{
		"success | no error": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerJobBaseImageRef,
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
			},
			clusterController: func() cluster.Controller {
				ctlr := &clustermock.Controller{}
				ctlr.On("GetJob", mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: "bicycle",
						},
					},
					nil,
				)
				ctlr.On("DeleteJob", mock.Anything, mock.Anything).Return(nil)
				return ctlr
			},
			hasErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clusterController := tt.clusterController()
			ib, _ := NewEnsemblerJobImageBuilder(clusterController, tt.imageBuildingConfig)
			err := ib.DeleteImageBuildingJob("", "", models.ID(1))

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
	}{
		"success | no error": {
			imageBuildingConfig: config.ImageBuildingConfig{
				BuildNamespace:       buildNamespace,
				BuildTimeoutDuration: timeout,
				DestinationRegistry:  dockerRegistry,
				BaseImageRef:         pyFuncEnsemblerServiceBaseImageRef,
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
				ctlr.On("GetJob", mock.Anything, mock.Anything).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: "bicycle",
						},
					},
					nil,
				)
				ctlr.On("DeleteJob", mock.Anything, mock.Anything).Return(nil)
				return ctlr
			},
			hasErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clusterController := tt.clusterController()
			ib, _ := NewEnsemblerJobImageBuilder(clusterController, tt.imageBuildingConfig)
			err := ib.DeleteImageBuildingJob("", "", models.ID(1))

			if tt.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
