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
	apibatchv1 "k8s.io/api/batch/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/log"
	"github.com/gojek/turing/api/turing/models"
)

// JobStatus is the current state of the image building job.
type JobStatus int

const (
	tickDurationInSeconds = 5
	// JobStatusActive is the status of the image building job is active
	JobStatusActive = JobStatus(iota)
	// JobStatusFailed is when the image building job has failed
	JobStatusFailed
	// JobStatusSucceeded is when the image building job has succeeded
	JobStatusSucceeded
	// JobStatusUnknown is when the image building job status is unknown
	JobStatusUnknown
)

// BuildImageRequest contains the information needed to build the OCI image
type BuildImageRequest struct {
	ProjectName string
	ModelName   string
	VersionID   models.ID
	ArtifactURI string
	BuildLabels map[string]string
}

// ImageBuilder defines the operations on building and publishing OCI images.
type ImageBuilder interface {
	// Build OCI image based on a Dockerfile
	BuildImage(request BuildImageRequest) (string, error)
	GetImageBuildingJobStatus(
		projectName string,
		modelName string,
		versionID models.ID,
	) (JobStatus, error)
}

type nameGenerator interface {
	// generateBuilderJobName generate kaniko job name that will be used to build a docker image
	generateBuilderJobName(projectName string, modelName string, versionID models.ID) string
	// generateDockerImageName generate image name based on project and model
	generateDockerImageName(projectName string, modelName string) string
}

type imageBuilder struct {
	clusterController cluster.Controller
	imageConfig       config.ImageBuilderConfig
	kanikoConfig      config.KanikoConfig
	nameGenerator     nameGenerator
}

// NewImageBuilder creates a new ImageBuilder
func newImageBuilder(
	clusterController cluster.Controller,
	imageConfig config.ImageBuilderConfig,
	kanikoConfig config.KanikoConfig,
	nameGenerator nameGenerator,
) (ImageBuilder, error) {
	err := checkParseResources(kanikoConfig.ResourceRequestsLimits)
	if err != nil {
		return nil, ErrUnableToParseKanikoResource
	}

	return &imageBuilder{
		clusterController: clusterController,
		imageConfig:       imageConfig,
		kanikoConfig:      kanikoConfig,
		nameGenerator:     nameGenerator,
	}, nil
}

func (ib *imageBuilder) BuildImage(request BuildImageRequest) (string, error) {
	imageName := ib.nameGenerator.generateDockerImageName(request.ProjectName, request.ModelName)
	imageExists, err := ib.checkIfImageExists(imageName, strconv.Itoa(int(request.VersionID)))
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
	kanikoJobName := ib.nameGenerator.generateBuilderJobName(
		request.ProjectName,
		request.ModelName,
		request.VersionID,
	)
	job, err := ib.clusterController.GetJob(
		ib.imageConfig.BuildNamespace,
		kanikoJobName,
	)

	if err != nil {
		if !kerrors.IsNotFound(err) {
			log.Errorf("error retrieving job status: %v", err)
			return "", ErrUnableToGetJobStatus
		}

		job, err = ib.createKanikoJob(kanikoJobName, imageRef, request.ArtifactURI, request.BuildLabels)
		if err != nil {
			log.Errorf("unable to build image %s, error: %v", imageRef, err)
			return "", ErrUnableToBuildImage
		}
	} else {
		// Only recreate when job has failed too many times, else no action required and just wait for it to finish
		if job.Status.Failed != 0 {
			// job already created before so we have to delete it first if it failed
			err = ib.clusterController.DeleteJob(ib.imageConfig.BuildNamespace, job.Name)
			if err != nil {
				log.Errorf("error deleting job: %v", err)
				return "", ErrDeleteFailedJob
			}

			job, err = ib.createKanikoJob(kanikoJobName, imageRef, request.ArtifactURI, request.BuildLabels)
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

func (ib *imageBuilder) waitForJobToFinish(job *apibatchv1.Job) error {
	timeout := time.After(ib.imageConfig.BuildTimeoutDuration)
	ticker := time.NewTicker(time.Second * tickDurationInSeconds)

	for {
		select {
		case <-timeout:
			log.Errorf("timeout waiting for kaniko job completion %s", job.Name)
			return ErrTimeoutBuildingImage
		case <-ticker.C:
			j, err := ib.clusterController.GetJob(ib.imageConfig.BuildNamespace, job.Name)
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

func (ib *imageBuilder) createKanikoJob(
	kanikoJobName string,
	imageRef string,
	artifactURI string,
	buildLabels map[string]string,
) (*apibatchv1.Job, error) {
	splitURI := strings.Split(artifactURI, "/")
	folderName := splitURI[len(splitURI)-1]

	kanikoArgs := []string{
		fmt.Sprintf("--dockerfile=%s", ib.imageConfig.DockerfileFilePath),
		fmt.Sprintf("--context=%s", ib.imageConfig.BuildContextURI),
		fmt.Sprintf("--build-arg=MODEL_URL=%s", artifactURI),
		fmt.Sprintf("--build-arg=BASE_IMAGE=%s", ib.imageConfig.BaseImageRef),
		fmt.Sprintf("--build-arg=FOLDER_NAME=%s", folderName),
		fmt.Sprintf("--destination=%s", imageRef),
		"--cache=true",
		"--single-snapshot",
	}

	spec := cluster.KanikoJobSpec{
		JobName:       kanikoJobName,
		Namespace:     ib.imageConfig.BuildNamespace,
		Labels:        buildLabels,
		Args:          kanikoArgs,
		Image:         ib.kanikoConfig.Image,
		Version:       ib.kanikoConfig.ImageVersion,
		CPURequest:    resource.MustParse(ib.kanikoConfig.ResourceRequestsLimits.Requests.CPU),
		MemoryRequest: resource.MustParse(ib.kanikoConfig.ResourceRequestsLimits.Requests.Memory),
		CPULimit:      resource.MustParse(ib.kanikoConfig.ResourceRequestsLimits.Limits.CPU),
		MemoryLimit:   resource.MustParse(ib.kanikoConfig.ResourceRequestsLimits.Limits.Memory),
	}

	return ib.clusterController.CreateKanikoJob(
		ib.imageConfig.BuildNamespace,
		spec,
	)
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

func checkParseResources(resourceRequestsLimits config.ResourceRequestsLimits) error {
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

func (ib *imageBuilder) GetImageBuildingJobStatus(
	projectName string,
	modelName string,
	versionID models.ID,
) (JobStatus, error) {
	kanikoJobName := ib.nameGenerator.generateBuilderJobName(
		projectName,
		modelName,
		versionID,
	)
	job, err := ib.clusterController.GetJob(
		ib.imageConfig.BuildNamespace,
		kanikoJobName,
	)
	if err != nil {
		return JobStatusUnknown, err
	}

	if job.Status.Active != 0 {
		return JobStatusActive, nil
	}

	if job.Status.Succeeded != 0 {
		return JobStatusSucceeded, nil
	}

	if job.Status.Failed != 0 {
		return JobStatusFailed, nil
	}

	return JobStatusUnknown, nil
}
