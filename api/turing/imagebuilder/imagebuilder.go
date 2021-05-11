package imagebuilder

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/gojek/turing/api/turing/log"
)

const (
	googleApplicationEnvVarName = "GOOGLE_APPLICATION_CREDENTIALS"
	kanikoSecretName            = "kaniko-secret"
	kanikoSecretMountpath       = "/secret"
	kanikoSecretFileName        = "kaniko-secret.json"
	imageBuilderContainerName   = "kaniko-builder"
	tickDurationInSeconds       = 5
)

var (
	jobCompletions            int32 = 1
	jobBackOffLimit           int32 = 3
	jobTTLSecondAfterComplete int32 = 3600 * 24
)

// BuildImageRequest contains the information needed to build the OCI image
type BuildImageRequest struct {
	ProjectName string
	ModelName   string
	VersionID   int
	ArtifactURI string
	BuildLabels map[string]string
}

// ImageBuilder defines the operations on building and publishing OCI images.
type ImageBuilder interface {
	// Build OCI image based on a Dockerfile
	BuildImage(request BuildImageRequest) (string, error)
}

type nameGenerator interface {
	// generateBuilderJobName generate kaniko job name that will be used to build a docker image
	generateBuilderJobName(projectName string, modelName string, versionID int) string
	// generateDockerImageName generate image name based on project and model
	generateDockerImageName(projectName string, modelName string) string
}

type imageBuilder struct {
	kubeClient    kubernetes.Interface
	imageConfig   ImageConfig
	kanikoConfig  KanikoConfig
	nameGenerator nameGenerator
}

// NewImageBuilder creates a new ImageBuilder
func newImageBuilder(
	kubeClient kubernetes.Interface,
	imageConfig ImageConfig,
	kanikoConfig KanikoConfig,
	nameGenerator nameGenerator,
) (ImageBuilder, error) {
	err := checkParseResources(kanikoConfig.ResourceRequestsLimits)
	if err != nil {
		return nil, ErrUnableToParseKanikoResource
	}

	return &imageBuilder{
		kubeClient:    kubeClient,
		imageConfig:   imageConfig,
		kanikoConfig:  kanikoConfig,
		nameGenerator: nameGenerator,
	}, nil
}

func (ib *imageBuilder) BuildImage(request BuildImageRequest) (string, error) {
	imageName := ib.nameGenerator.generateDockerImageName(request.ProjectName, request.ModelName)
	imageExists, err := ib.checkIfImageExists(imageName, strconv.Itoa(request.VersionID))
	imageRef := fmt.Sprintf("%s:%d", imageName, request.VersionID)
	if err != nil {
		log.Errorf("Unable to check existing image ref: %v", err)
		return "", ErrUnableToGetImageRef
	}

	if imageExists {
		log.Infof("Image %s already exists. Skipping build.", imageName)
		return imageRef, nil
	}

	// Check if there is an existing build job
	jobClient := ib.kubeClient.BatchV1().Jobs(ib.imageConfig.BuildNamespace)
	kanikoPodName := ib.nameGenerator.generateBuilderJobName(request.ProjectName, request.ModelName, request.VersionID)
	job, err := jobClient.Get(kanikoPodName, metav1.GetOptions{})

	if err != nil {
		if !kerrors.IsNotFound(err) {
			log.Errorf("error retrieving job status: %v", err)
			return "", ErrUnableToGetJobStatus
		}

		jobSpec := ib.createKanikoJobSpec(kanikoPodName, imageRef, request.ArtifactURI, request.BuildLabels)
		job, err = jobClient.Create(jobSpec)
		if err != nil {
			log.Errorf("unable to build image %s, error: %v", imageRef, err)
			return "", ErrUnableToBuildImage
		}
	} else {
		// Only recreate when job has failed too many times, else no action required and just wait for it to finish
		if job.Status.Failed != 0 {
			// job already created before so we have to delete it first if it failed
			err = jobClient.Delete(job.Name, &metav1.DeleteOptions{})
			if err != nil {
				log.Errorf("error deleting job: %v", err)
				return "", ErrDeleteFailedJob
			}

			jobSpec := ib.createKanikoJobSpec(kanikoPodName, imageRef, request.ArtifactURI, request.BuildLabels)
			job, err = jobClient.Create(jobSpec)
			if err != nil {
				log.Errorf("unable to build image %s, error: %v", imageRef, err)
				return "", ErrUnableToBuildImage
			}
		}
	}

	err = ib.waitForJobToFinish(job)
	if err != nil {
		return "", err
	}

	return imageRef, nil
}

func (ib *imageBuilder) waitForJobToFinish(job *batchv1.Job) error {
	timeout := time.After(ib.imageConfig.BuildTimeoutDuration)
	ticker := time.NewTicker(time.Second * tickDurationInSeconds)
	jobClient := ib.kubeClient.BatchV1().Jobs(ib.imageConfig.BuildNamespace)

	for {
		select {
		case <-timeout:
			log.Errorf("timeout waiting for kaniko job completion %s", job.Name)
			return ErrTimeoutBuildingImage
		case <-ticker.C:
			j, err := jobClient.Get(job.Name, metav1.GetOptions{})
			if err != nil {
				log.Errorf("unable to get job status for job %s: %v", job.Name, err)
				return ErrUnableToBuildImage
			}

			if j.Status.Succeeded == 1 {
				// successfully created pod
				return nil
			} else if j.Status.Failed == 1 {
				log.Errorf("failed building OCI image %s: %v", job.Name, j.Status)
				return ErrUnableToBuildImage
			}
		}
	}
}

func (ib *imageBuilder) createKanikoJobSpec(
	kanikoPodName string,
	imageRef string,
	artifactURI string,
	buildLabels map[string]string,
) *batchv1.Job {
	splitURI := strings.Split(artifactURI, "/")
	folderName := splitURI[len(splitURI)-1]

	kanikoArgs := []string{
		fmt.Sprintf("--dockerfile=%s", ib.imageConfig.DockerfileFilePath),
		fmt.Sprintf("--context=%s", ib.imageConfig.BuildContextURI),
		fmt.Sprintf("--build-arg=MODEL_URL=%s", artifactURI),
		fmt.Sprintf("--build-arg=BASE_IMAGE=%s", ib.imageConfig.BaseImageRef),
		fmt.Sprintf("--build-arg=FOLDER_NAME=%s", folderName),
		fmt.Sprintf("--destination=%s", imageRef),
		"--single-snapshot",
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kanikoPodName,
			Namespace: ib.imageConfig.BuildNamespace,
			Labels:    buildLabels,
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
							Name:  imageBuilderContainerName,
							Image: fmt.Sprintf("%s:%s", ib.kanikoConfig.Image, ib.kanikoConfig.ImageVersion),
							Args:  kanikoArgs,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      kanikoSecretName,
									MountPath: kanikoSecretMountpath,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  googleApplicationEnvVarName,
									Value: fmt.Sprintf("%s/%s", kanikoSecretMountpath, kanikoSecretFileName),
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(ib.kanikoConfig.ResourceRequestsLimits.Requests.CPU),
									corev1.ResourceMemory: resource.MustParse(ib.kanikoConfig.ResourceRequestsLimits.Requests.Memory),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse(ib.kanikoConfig.ResourceRequestsLimits.Limits.CPU),
									corev1.ResourceMemory: resource.MustParse(ib.kanikoConfig.ResourceRequestsLimits.Limits.Memory),
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: kanikoSecretName,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: kanikoSecretName,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (ib *imageBuilder) checkIfImageExists(imageName string, imageTag string) (bool, error) {
	keychain := authn.DefaultKeychain
	if strings.Contains(ib.imageConfig.Registry, "gcr.io") {
		keychain = google.Keychain
	}

	repo, err := name.NewRepository(imageName)
	if err != nil {
		return false, errors.Wrapf(err, "unable to parse docker repository %s", imageName)
	}

	tags, err := remote.List(repo, remote.WithAuthFromKeychain(keychain))
	if err != nil {
		if terr, ok := err.(*transport.Error); ok {
			// If image not found, it does not exist yet
			if terr.StatusCode == http.StatusNotFound {
				return false, nil
			}
		} else {
			return false, errors.Wrapf(err, "error getting image tags for %s", repo)
		}
	}

	for _, tag := range tags {
		if tag == imageTag {
			return true, nil
		}
	}

	return false, nil
}

func checkParseResources(resourceRequestsLimits ResourceRequestsLimits) error {
	_, err := resource.ParseQuantity(resourceRequestsLimits.Requests.CPU)
	if err != nil {
		return err
	}

	_, err = resource.ParseQuantity(resourceRequestsLimits.Requests.Memory)
	if err != nil {
		return err
	}

	_, err = resource.ParseQuantity(resourceRequestsLimits.Limits.CPU)
	if err != nil {
		return err
	}

	_, err = resource.ParseQuantity(resourceRequestsLimits.Limits.Memory)
	if err != nil {
		return err
	}

	return nil
}
