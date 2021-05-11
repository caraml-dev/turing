package imagebuilder

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	fakebatchv1 "k8s.io/client-go/kubernetes/typed/batch/v1/fake"
	ktesting "k8s.io/client-go/testing"
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
		existingJob       *batchv1.Job
		wantCreateJob     *batchv1.Job
		inputDependencies []string
		namespace         string
		imageConfig       ImageConfig
		kanikoConfig      KanikoConfig
		buildLabels       map[string]string
	}{
		"success | no existing job": {
			expected:    fmt.Sprintf("%s/%s-%s-job:%d", dockerRegistry, projectName, modelName, modelVersion),
			projectName: projectName,
			modelName:   modelName,
			versionID:   modelVersion,
			artifactURI: artifactURI,
			existingJob: nil,
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, modelVersion),
					Namespace: buildNamespace,
					Labels: map[string]string{
						"gojek.com/team": "dsp",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers: []corev1.Container{
								{
									Name:  "kaniko-builder",
									Image: kanikoImageRef,
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", dockerfilePath),
										fmt.Sprintf("--context=%s", buildContext),
										fmt.Sprintf("--build-arg=MODEL_URL=%s", artifactURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", baseImageRef),
										fmt.Sprintf("--build-arg=FOLDER_NAME=%s", folderName),
										fmt.Sprintf(
											"--destination=%s/%s-%s-job:%d",
											dockerRegistry,
											projectName,
											modelName,
											modelVersion,
										),
										"--cache=true",
										"--single-snapshot",
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "kaniko-secret",
											MountPath: "/secret",
										},
									},
									Env: []corev1.EnvVar{
										{
											Name:  "GOOGLE_APPLICATION_CREDENTIALS",
											Value: "/secret/kaniko-secret.json",
										},
									},
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "kaniko-secret",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "kaniko-secret",
										},
									},
								},
							},
						},
					},
				},
				Status: batchv1.JobStatus{},
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
			existingJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, modelVersion),
					Namespace: buildNamespace,
					Labels: map[string]string{
						"gojek.com/team": "dsp",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers: []corev1.Container{
								{
									Name:  "kaniko-builder",
									Image: kanikoImageRef,
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", dockerfilePath),
										fmt.Sprintf("--context=%s", buildContext),
										fmt.Sprintf("--build-arg=MODEL_URL=%s", artifactURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", baseImageRef),
										fmt.Sprintf("--build-arg=FOLDER_NAME=%s", folderName),
										fmt.Sprintf(
											"--destination=%s/%s-%s-job:%d",
											dockerRegistry,
											projectName,
											modelName,
											modelVersion,
										),
										"--cache=true",
										"--single-snapshot",
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "kaniko-secret",
											MountPath: "/secret",
										},
									},
									Env: []corev1.EnvVar{
										{
											Name:  "GOOGLE_APPLICATION_CREDENTIALS",
											Value: "/secret/kaniko-secret.json",
										},
									},
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "kaniko-secret",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "kaniko-secret",
										},
									},
								},
							},
						},
					},
				},
				Status: batchv1.JobStatus{},
			},
			wantCreateJob: nil,
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
		"success: existing job has already completed successfully": {
			expected:    fmt.Sprintf("%s/%s-%s-job:%d", dockerRegistry, projectName, modelName, modelVersion),
			projectName: projectName,
			modelName:   modelName,
			versionID:   modelVersion,
			artifactURI: artifactURI,
			existingJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, modelVersion),
					Namespace: buildNamespace,
					Labels: map[string]string{
						"gojek.com/team": "dsp",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers: []corev1.Container{
								{
									Name:  "kaniko-builder",
									Image: kanikoImageRef,
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", dockerfilePath),
										fmt.Sprintf("--context=%s", buildContext),
										fmt.Sprintf("--build-arg=MODEL_URL=%s", artifactURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", baseImageRef),
										fmt.Sprintf("--build-arg=FOLDER_NAME=%s", folderName),
										fmt.Sprintf(
											"--destination=%s/%s-%s-job:%d",
											dockerRegistry,
											projectName,
											modelName,
											modelVersion,
										),
										"--cache=true",
										"--single-snapshot",
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "kaniko-secret",
											MountPath: "/secret",
										},
									},
									Env: []corev1.EnvVar{
										{
											Name:  "GOOGLE_APPLICATION_CREDENTIALS",
											Value: "/secret/kaniko-secret.json",
										},
									},
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "kaniko-secret",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "kaniko-secret",
										},
									},
								},
							},
						},
					},
				},
				Status: batchv1.JobStatus{
					Succeeded: 1,
				},
			},
			wantCreateJob: nil,
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
		"success: existing job failed": {
			expected:    fmt.Sprintf("%s/%s-%s-job:%d", dockerRegistry, projectName, modelName, modelVersion),
			projectName: projectName,
			modelName:   modelName,
			versionID:   modelVersion,
			artifactURI: artifactURI,
			existingJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, modelVersion),
					Namespace: buildNamespace,
					Labels: map[string]string{
						"gojek.com/team": "dsp",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers: []corev1.Container{
								{
									Name:  "kaniko-builder",
									Image: kanikoImageRef,
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", dockerfilePath),
										fmt.Sprintf("--context=%s", buildContext),
										fmt.Sprintf("--build-arg=MODEL_URL=%s", artifactURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", baseImageRef),
										fmt.Sprintf("--build-arg=FOLDER_NAME=%s", folderName),
										fmt.Sprintf(
											"--destination=%s/%s-%s-job:%d",
											dockerRegistry,
											projectName,
											modelName,
											modelVersion,
										),
										"--cache=true",
										"--single-snapshot",
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "kaniko-secret",
											MountPath: "/secret",
										},
									},
									Env: []corev1.EnvVar{
										{
											Name:  "GOOGLE_APPLICATION_CREDENTIALS",
											Value: "/secret/kaniko-secret.json",
										},
									},
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "kaniko-secret",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "kaniko-secret",
										},
									},
								},
							},
						},
					},
				},
				Status: batchv1.JobStatus{
					Failed: 1,
				},
			},
			wantCreateJob: &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("batch-%s-%s-%d", projectName, modelName, modelVersion),
					Namespace: buildNamespace,
					Labels: map[string]string{
						"gojek.com/team": "dsp",
					},
				},
				Spec: batchv1.JobSpec{
					Completions:             &jobCompletions,
					BackoffLimit:            &jobBackOffLimit,
					TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyNever,
							Containers: []corev1.Container{
								{
									Name:  "kaniko-builder",
									Image: kanikoImageRef,
									Args: []string{
										fmt.Sprintf("--dockerfile=%s", dockerfilePath),
										fmt.Sprintf("--context=%s", buildContext),
										fmt.Sprintf("--build-arg=MODEL_URL=%s", artifactURI),
										fmt.Sprintf("--build-arg=BASE_IMAGE=%s", baseImageRef),
										fmt.Sprintf("--build-arg=FOLDER_NAME=%s", folderName),
										fmt.Sprintf(
											"--destination=%s/%s-%s-job:%d",
											dockerRegistry,
											projectName,
											modelName,
											modelVersion,
										),
										"--cache=true",
										"--single-snapshot",
									},
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "kaniko-secret",
											MountPath: "/secret",
										},
									},
									Env: []corev1.EnvVar{
										{
											Name:  "GOOGLE_APPLICATION_CREDENTIALS",
											Value: "/secret/kaniko-secret.json",
										},
									},
									Resources: corev1.ResourceRequirements{
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("2Gi"),
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "kaniko-secret",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "kaniko-secret",
										},
									},
								},
							},
						},
					},
				},
				Status: batchv1.JobStatus{},
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
			// Configure Fake Kubernetes Client
			kubeClient := fake.NewSimpleClientset()
			client := kubeClient.BatchV1().Jobs(tt.imageConfig.BuildNamespace).(*fakebatchv1.FakeJobs)

			client.Fake.PrependReactor(
				"get",
				"jobs",
				func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					client.Fake.PrependReactor(
						"get",
						"jobs",
						func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
							if tt.existingJob != nil {
								successfulJob := tt.existingJob.DeepCopy()
								successfulJob.Status.Succeeded = 1
								return true, successfulJob, nil
							} else if tt.wantCreateJob != nil {
								successfulJob := tt.wantCreateJob.DeepCopy()
								successfulJob.Status.Succeeded = 1
								return true, successfulJob, nil
							} else {
								assert.Fail(t, "either existingJob or wantCreateJob must be not nil")
								panic("should not reach this code")
							}
						},
					)
					if tt.existingJob != nil {
						return true, tt.existingJob, nil
					}
					return true, nil, kerrors.NewNotFound(schema.ParseGroupResource("v1"), action.(ktesting.GetAction).GetName())
				},
			)

			client.Fake.PrependReactor(
				"create",
				"jobs",
				func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					po := action.(ktesting.CreateAction).GetObject().(*batchv1.Job)
					return true, &batchv1.Job{
						ObjectMeta: metav1.ObjectMeta{
							Name: po.Name,
						},
					}, nil
				},
			)

			client.Fake.PrependReactor(
				"delete",
				"jobs",
				func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, nil
				},
			)

			ib, err := NewEnsemberJobImageBuilder(kubeClient, tt.imageConfig, tt.kanikoConfig)
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
		"failure |cpu_limit_error": {
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
