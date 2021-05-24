package imagebuilder

import (
	"fmt"
	"testing"
	"time"

	"github.com/gojek/turing/api/turing/cluster"
	clustermock "github.com/gojek/turing/api/turing/cluster/mocks"
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
	projectName    = "test-project"
	modelName      = "mymodel"
	modelVersion   = 1
	dockerRegistry = "ghcr.io"
	artifactURI    = "gs://bucket/ensembler"
	dockerfilePath = "engines/batch-ensembler/app.Dockerfile"
	buildContext   = "git://github.com/gojek/turing.git#refs/heads/master"
	baseImageRef   = "ghcr.io/gojek/turing/batch-ensembler:0.0.0-build.1-98b071d"
	folderName     = "ensembler"
	buildNamespace = "mlp"
	kanikoImageRef = "gcr.io/kaniko-project/executor:v1.5.2"
)

func TestBuildEnsemblerImage(t *testing.T) {
	var tests = map[string]struct {
		name              string
		expected          string
		projectName       string
		modelName         string
		artifactURI       string
		versionID         int
		inputDependencies []string
		namespace         string
		imageConfig       ImageConfig
		kanikoConfig      KanikoConfig
		buildLabels       map[string]string
		clusterController func() cluster.Controller
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
					"CreateKanikoJob",
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, modelVersion),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			buildLabels: map[string]string{
				"gojek.io/team": "dsp",
			},
			imageConfig: ImageConfig{
				Registry:             dockerRegistry,
				BaseImageRef:         baseImageRef,
				BuildNamespace:       buildNamespace,
				BuildContextURI:      buildContext,
				DockerfileFilePath:   dockerfilePath,
				BuildTimeoutDuration: timeout,
			},
			kanikoConfig: KanikoConfig{
				Image:        "gcr.io/kaniko-project/executor",
				ImageVersion: "v1.5.2",
				ResourceRequestsLimits: ResourceRequestsLimits{
					Requests: Resource{
						CPU:    "1",
						Memory: "2Gi",
					},
					Limits: Resource{
						CPU:    "1",
						Memory: "2Gi",
					},
				},
			},
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
							Name: fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, modelVersion),
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
					"CreateKanikoJob",
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, modelVersion),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			imageConfig: ImageConfig{
				Registry:             dockerRegistry,
				BaseImageRef:         baseImageRef,
				BuildNamespace:       buildNamespace,
				BuildContextURI:      buildContext,
				DockerfileFilePath:   dockerfilePath,
				BuildTimeoutDuration: timeout,
			},
			kanikoConfig: KanikoConfig{
				Image:        "gcr.io/kaniko-project/executor",
				ImageVersion: "v1.5.2",
				ResourceRequestsLimits: ResourceRequestsLimits{
					Requests: Resource{
						CPU:    "1",
						Memory: "2Gi",
					},
					Limits: Resource{
						CPU:    "1",
						Memory: "2Gi",
					},
				},
			},
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
							Name: fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, modelVersion),
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
					"CreateKanikoJob",
					mock.Anything,
					mock.Anything,
				).Return(
					&apibatchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, modelVersion),
						},
					},
					nil,
				).Once()

				return ctlr
			},
			buildLabels: map[string]string{
				"gojek.io/team": "dsp",
			},
			imageConfig: ImageConfig{
				Registry:             dockerRegistry,
				BaseImageRef:         baseImageRef,
				BuildNamespace:       buildNamespace,
				BuildContextURI:      buildContext,
				DockerfileFilePath:   dockerfilePath,
				BuildTimeoutDuration: timeout,
			},
			kanikoConfig: KanikoConfig{
				Image:        "gcr.io/kaniko-project/executor",
				ImageVersion: "v1.5.2",
				ResourceRequestsLimits: ResourceRequestsLimits{
					Requests: Resource{
						CPU:    "1",
						Memory: "2Gi",
					},
					Limits: Resource{
						CPU:    "1",
						Memory: "2Gi",
					},
				},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clusterController := tt.clusterController()

			ib, err := NewEnsemberJobImageBuilder(clusterController, tt.imageConfig, tt.kanikoConfig)
			assert.Nil(t, err)

			buildImageRequest := BuildImageRequest{
				tt.projectName,
				tt.modelName,
				tt.versionID,
				tt.artifactURI,
				tt.buildLabels,
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
		resourceRequestsLimits ResourceRequestsLimits
	}{
		"success | parsable": {
			expected: true,
			resourceRequestsLimits: ResourceRequestsLimits{
				Requests: Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
				Limits: Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
			},
		},
		"failure | cpu_request_error": {
			expected: false,
			resourceRequestsLimits: ResourceRequestsLimits{
				Requests: Resource{
					CPU:    "Chicken",
					Memory: "2Gi",
				},
				Limits: Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
			},
		},
		"failure | cpu_limit_error": {
			expected: false,
			resourceRequestsLimits: ResourceRequestsLimits{
				Requests: Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
				Limits: Resource{
					CPU:    "Vegetable",
					Memory: "2Gi",
				},
			},
		},
		"failure | memory_request_error": {
			expected: false,
			resourceRequestsLimits: ResourceRequestsLimits{
				Requests: Resource{
					CPU:    "1",
					Memory: "Brains",
				},
				Limits: Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
			},
		},
		"failure | memory_limit_error": {
			expected: false,
			resourceRequestsLimits: ResourceRequestsLimits{
				Requests: Resource{
					CPU:    "1",
					Memory: "2Gi",
				},
				Limits: Resource{
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
