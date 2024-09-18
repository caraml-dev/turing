package imagebuilder

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/google"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/pkg/errors"
	apibatchv1 "k8s.io/api/batch/v1"
	apicorev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/caraml-dev/merlin/utils"
	mlp "github.com/caraml-dev/mlp/api/client"
	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/log"
	"github.com/caraml-dev/turing/api/turing/models"
)

var (
	jobCompletions            int32 = 1
	jobBackOffLimit           int32 = 3
	jobTTLSecondAfterComplete int32 = 3600 * 24
)

const (
	imageBuilderContainerName        = "kaniko-builder"
	googleApplicationEnvVarName      = "GOOGLE_APPLICATION_CREDENTIALS"
	kanikoSecretName                 = "kaniko-secret"
	kanikoSecretFileName             = "kaniko-secret.json"
	kanikoSecretMountpath            = "/secret"
	kanikoDockerCredentialConfigPath = "/kaniko/.docker"
)

// JobStatus is the current status of the image building job.
type JobStatus struct {
	State   JobState `json:"state"`
	Message string   `json:"message,omitempty"`
}

func (js JobStatus) IsActive() bool {
	return js.State == JobStateActive
}

type JobState string

const (
	// jobDeletionTimeoutInSeconds is the maximum time to wait for a job to be deleted from a cluster
	jobDeletionTimeoutInSeconds = 30
	// jobDeletionTickDurationInMilliseconds is the interval at which the API server checks if a job has been deleted
	jobDeletionTickDurationInMilliseconds = 100
	// jobCompletionTickDurationInSeconds is the interval at which the API server checks if a job has completed
	jobCompletionTickDurationInSeconds = 5
)

const (
	// JobStateActive is the status of the image building job is active
	JobStateActive JobState = "active"
	// JobStateSucceeded is when the image building job has succeeded
	JobStateSucceeded JobState = "succeeded"
	// JobStateFailed is when the image building job has failed
	JobStateFailed JobState = "failed"
	// JobStateUnknown is when the image building job status is unknown
	JobStateUnknown JobState = "unknown"
)

// BuildImageRequest contains the information needed to build the OCI image
type BuildImageRequest struct {
	ProjectName     string
	ResourceName    string
	ResourceID      models.ID
	VersionID       string
	ArtifactURI     string
	BuildLabels     map[string]string
	EnsemblerFolder string
	BaseImageRefTag string
}

type EnsemblerImage struct {
	ProjectID           models.ID                  `json:"project_id"`
	EnsemblerID         models.ID                  `json:"ensembler_id"`
	EnsemblerRunnerType models.EnsemblerRunnerType `json:"runner_type"`
	ImageRef            string                     `json:"image_ref"`
	Exists              bool                       `json:"exists"`
	JobStatus           JobStatus                  `json:"image_building_job_status"`
}

// ImageBuilder defines the operations on building and publishing OCI images.
type ImageBuilder interface {
	// Build OCI image based on a Dockerfile
	BuildImage(request BuildImageRequest) (string, error)
	GetEnsemblerImage(project *mlp.Project, ensembler *models.PyFuncEnsembler) (EnsemblerImage, error)
	GetImageBuildingJobStatus(
		projectName string,
		ensemblerName string,
		ensemblerID models.ID,
		versionID string,
	) JobStatus
	DeleteImageBuildingJob(
		projectName string,
		ensemblerName string,
		ensemblerID models.ID,
		versionID string,
	) error
}

type nameGenerator interface {
	// generateBuilderJobName generate kaniko job name that will be used to build a docker image
	generateBuilderName(projectName string, ensemblerName string, ensemblerID models.ID, versionID string) string
	// generateDockerImageName generate image name based on project and model
	generateDockerImageName(projectName string, ensemblerName string) string
}

type imageBuilder struct {
	clusterController   cluster.Controller
	imageBuildingConfig config.ImageBuildingConfig
	nameGenerator       nameGenerator
	runnerType          models.EnsemblerRunnerType
	artifactServiceType string
}

// NewImageBuilder creates a new ImageBuilder
func newImageBuilder(
	clusterController cluster.Controller,
	imageBuildingConfig config.ImageBuildingConfig,
	nameGenerator nameGenerator,
	runnerType models.EnsemblerRunnerType,
	artifactServiceType string,
) (ImageBuilder, error) {
	err := checkParseResources(imageBuildingConfig.KanikoConfig.ResourceRequestsLimits)
	if err != nil {
		return nil, ErrUnableToParseKanikoResource
	}

	return &imageBuilder{
		clusterController:   clusterController,
		imageBuildingConfig: imageBuildingConfig,
		nameGenerator:       nameGenerator,
		runnerType:          runnerType,
		artifactServiceType: artifactServiceType,
	}, nil
}

func (ib *imageBuilder) BuildImage(request BuildImageRequest) (string, error) {
	imageName := ib.nameGenerator.generateDockerImageName(request.ProjectName, request.ResourceName)
	imageExists, err := ib.checkIfImageExists(imageName, request.VersionID)
	imageRef := fmt.Sprintf("%s:%s", imageName, request.VersionID)
	if err != nil {
		log.Errorf("Unable to check existing image ref: %v", err)
		return "", ErrUnableToGetImageRef
	}

	if imageExists {
		log.Infof("Image %s already exists. Skipping build.", imageName)
		return imageRef, nil
	}

	// Check if there is an existing build job
	kanikoJobName := ib.nameGenerator.generateBuilderName(
		request.ProjectName,
		request.ResourceName,
		request.ResourceID,
		request.VersionID,
	)
	job, err := ib.clusterController.GetJob(
		context.Background(),
		ib.imageBuildingConfig.BuildNamespace,
		kanikoJobName,
	)

	if err != nil {
		if !kerrors.IsNotFound(err) {
			log.Errorf("error retrieving job status: %v", err)
			return "", ErrUnableToGetJobStatus
		}

		job, err = ib.createKanikoJob(kanikoJobName, imageRef, request.ArtifactURI, request.BuildLabels,
			request.EnsemblerFolder, request.BaseImageRefTag)
		if err != nil {
			log.Errorf("unable to build image %s, error: %v", imageRef, err)
			return "", ErrUnableToBuildImage
		}
	} else {
		// Only recreate when job has failed too many times, else no action required and just wait for it to finish
		if job.Status.Failed != 0 {
			// job already created before, so we have to delete it first if it failed
			err = ib.clusterController.DeleteJob(context.Background(), ib.imageBuildingConfig.BuildNamespace, job.Name)
			if err != nil {
				log.Errorf("error deleting job: %v", err)
				return "", ErrDeleteFailedJob
			}

			err = ib.waitForJobToBeDeleted(job)
			if err != nil {
				return "", ErrDeleteFailedJob
			}

			job, err = ib.createKanikoJob(kanikoJobName, imageRef, request.ArtifactURI, request.BuildLabels,
				request.EnsemblerFolder, request.BaseImageRefTag)
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
	timeout := time.After(ib.imageBuildingConfig.BuildTimeoutDuration)
	ticker := time.NewTicker(time.Second * jobCompletionTickDurationInSeconds)

	for {
		select {
		case <-timeout:
			return ErrTimeoutBuildingImage
		case <-ticker.C:
			j, err := ib.clusterController.GetJob(context.Background(), ib.imageBuildingConfig.BuildNamespace, job.Name)
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

func (ib *imageBuilder) waitForJobToBeDeleted(job *apibatchv1.Job) error {
	timeout := time.After(time.Second * jobDeletionTimeoutInSeconds)
	ticker := time.NewTicker(time.Millisecond * jobDeletionTickDurationInMilliseconds)

	for {
		select {
		case <-timeout:
			return ErrDeleteFailedJob
		case <-ticker.C:
			_, err := ib.clusterController.GetJob(context.Background(), ib.imageBuildingConfig.BuildNamespace, job.Name)
			if err != nil {
				if kerrors.IsNotFound(err) {
					return nil
				}
				log.Errorf("unable to get job status for job %s: %v", job.Name, err)
				return ErrDeleteFailedJob
			}
		}
	}
}

func (ib *imageBuilder) createKanikoJob(
	kanikoJobName string,
	imageRef string,
	artifactURI string,
	buildLabels map[string]string,
	ensemblerFolder string,
	baseImageRefTag string,
) (*apibatchv1.Job, error) {
	splitURI := strings.Split(artifactURI, "/")
	folderName := fmt.Sprintf("%s/%s", splitURI[len(splitURI)-1], ensemblerFolder)

	baseImage, ok := ib.imageBuildingConfig.BaseImageRef[baseImageRefTag]
	if !ok {
		return nil, fmt.Errorf("No matching base image for tag %s", baseImageRefTag)
	}

	kanikoSecretFilePath := fmt.Sprintf("%s/%s", kanikoSecretMountpath, kanikoSecretFileName)

	kanikoArgs := []string{
		fmt.Sprintf("--dockerfile=%s", ib.imageBuildingConfig.KanikoConfig.DockerfileFilePath),
		fmt.Sprintf("--context=%s", ib.imageBuildingConfig.KanikoConfig.BuildContextURI),
		fmt.Sprintf("--build-arg=MODEL_URL=%s", artifactURI),
		fmt.Sprintf("--build-arg=BASE_IMAGE=%s", baseImage),
		fmt.Sprintf("--build-arg=MLFLOW_ARTIFACT_STORAGE_TYPE=%s", ib.artifactServiceType),
		fmt.Sprintf("--build-arg=FOLDER_NAME=%s", folderName),
		fmt.Sprintf("--destination=%s", imageRef),
		"--cache=true",
		"--single-snapshot",
	}

	if ib.artifactServiceType == "s3" {
		kanikoArgs = append(
			kanikoArgs,
			fmt.Sprintf("--build-arg=%s=%s", "AWS_ACCESS_KEY_ID", os.Getenv("AWS_ACCESS_KEY_ID")),
			fmt.Sprintf("--build-arg=%s=%s", "AWS_SECRET_ACCESS_KEY", os.Getenv("AWS_SECRET_ACCESS_KEY")),
			fmt.Sprintf("--build-arg=%s=%s", "AWS_DEFAULT_REGION", os.Getenv("AWS_DEFAULT_REGION")),
			fmt.Sprintf("--build-arg=%s=%s", "AWS_ENDPOINT_URL", os.Getenv("AWS_ENDPOINT_URL")),
		)
	}

	annotations := make(map[string]string)
	if !ib.imageBuildingConfig.SafeToEvict {
		// The image-building jobs are timing out. We found that one of the root causes is the node pool got scaled down
		// resulting in the image building pods to be rescheduled.
		// Adding "cluster-autoscaler.kubernetes.io/safe-to-evict": "false" to avoid the pod get killed and rescheduled.
		// https://kubernetes.io/docs/reference/labels-annotations-taints/#cluster-autoscaler-kubernetes-io-safe-to-evict
		annotations["cluster-autoscaler.kubernetes.io/safe-to-evict"] = "false"
	}

	var volumes []cluster.SecretVolume
	var volumeMounts []cluster.VolumeMount
	var envVars []cluster.Env

	if ib.imageBuildingConfig.KanikoConfig.PushRegistryType == "gcr" {
		// If kaniko service account is not set, use kaniko secret
		if ib.imageBuildingConfig.KanikoConfig.ServiceAccount == "" {
			kanikoArgs = append(kanikoArgs,
				fmt.Sprintf("--build-arg=GOOGLE_APPLICATION_CREDENTIALS=%s", kanikoSecretFilePath))
			volumes = []cluster.SecretVolume{
				{
					Name:       kanikoSecretName,
					SecretName: kanikoSecretName,
				},
			}
			volumeMounts = []cluster.VolumeMount{
				{
					Name:      kanikoSecretName,
					MountPath: kanikoSecretMountpath,
				},
			}
			envVars = []cluster.Env{
				{
					Name:  googleApplicationEnvVarName,
					Value: kanikoSecretFilePath,
				},
			}
		}
	} else if ib.imageBuildingConfig.KanikoConfig.PushRegistryType == "docker" {
		volumes = []cluster.SecretVolume{
			{
				Name:       kanikoSecretName,
				SecretName: ib.imageBuildingConfig.KanikoConfig.DockerCredentialSecretName,
			},
		}
		volumeMounts = []cluster.VolumeMount{
			{
				Name:      kanikoSecretName,
				MountPath: kanikoDockerCredentialConfigPath,
			},
		}
	}

	job := cluster.Job{
		Name:                    kanikoJobName,
		Namespace:               ib.imageBuildingConfig.BuildNamespace,
		Labels:                  buildLabels,
		Annotations:             annotations,
		Completions:             &jobCompletions,
		BackOffLimit:            &jobBackOffLimit,
		TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
		RestartPolicy:           apicorev1.RestartPolicyNever,
		Containers: []cluster.Container{
			{
				Name: imageBuilderContainerName,
				Image: fmt.Sprintf(
					"%s:%s",
					ib.imageBuildingConfig.KanikoConfig.Image,
					ib.imageBuildingConfig.KanikoConfig.ImageVersion,
				),
				Args:         kanikoArgs,
				VolumeMounts: volumeMounts,
				Envs:         envVars,
				Resources: cluster.RequestLimitResources{
					Request: cluster.Resource{
						CPU: resource.MustParse(
							ib.imageBuildingConfig.KanikoConfig.ResourceRequestsLimits.Requests.CPU,
						),
						Memory: resource.MustParse(
							ib.imageBuildingConfig.KanikoConfig.ResourceRequestsLimits.Requests.Memory,
						),
					},
					Limit: cluster.Resource{
						CPU: resource.MustParse(
							ib.imageBuildingConfig.KanikoConfig.ResourceRequestsLimits.Limits.CPU,
						),
						Memory: resource.MustParse(
							ib.imageBuildingConfig.KanikoConfig.ResourceRequestsLimits.Limits.Memory,
						),
					},
				},
			},
		},
		SecretVolumes:  volumes,
		TolerationName: ib.imageBuildingConfig.TolerationName,
		NodeSelector:   ib.imageBuildingConfig.NodeSelector,
		ServiceAccount: ib.imageBuildingConfig.KanikoConfig.ServiceAccount,
	}

	return ib.clusterController.CreateJob(
		context.Background(),
		ib.imageBuildingConfig.BuildNamespace,
		job,
	)
}

// getGCPSubDomains returns the list of GCP container registry and artifact registry subdomains.
//
// GCP container registry and artifact registry domains are used to determine which keychain
// to use when interacting with container registry.
// This is needed because GCP registries use different authentication method than other container registry.
func getGCPSubDomains() []string {
	return []string{"gcr.io", "pkg.dev"}
}

func (ib *imageBuilder) checkIfImageExists(imageName string, imageTag string) (bool, error) {
	var keychain authn.Keychain
	keychain = authn.DefaultKeychain

	for _, domain := range getGCPSubDomains() {
		if strings.Contains(ib.imageBuildingConfig.DestinationRegistry, domain) {
			keychain = google.Keychain
		}
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
	ensemblerName string,
	ensemblerID models.ID,
	versionID string,
) (status JobStatus) {
	status.State = JobStateUnknown

	jobConditionTable := ""
	podContainerTable := ""
	podLastTerminationMessage := ""
	podLastTerminationReason := ""

	defer func() {
		if jobConditionTable != "" {
			status.Message = fmt.Sprintf("%s\n\nJob conditions:\n%s", status.Message, jobConditionTable)
		}

		if podContainerTable != "" {
			status.Message = fmt.Sprintf("%s\n\nPod container status:\n%s", status.Message, podContainerTable)
		}

		if podLastTerminationMessage != "" {
			status.Message = fmt.Sprintf("%s\n\nPod last termination message:\n%s", status.Message, podLastTerminationMessage)
		}
	}()

	kanikoJobName := ib.nameGenerator.generateBuilderName(
		projectName,
		ensemblerName,
		ensemblerID,
		versionID,
	)
	log.Infof("Checking status of image building job %s", kanikoJobName)
	job, err := ib.clusterController.GetJob(
		context.Background(),
		ib.imageBuildingConfig.BuildNamespace,
		kanikoJobName,
	)
	if err != nil && !kerrors.IsNotFound(err) {
		status.Message = err.Error()
		return
	}

	if job.Status.Active != 0 {
		status.State = JobStateActive
		return
	}

	if job.Status.Succeeded != 0 {
		status.State = JobStateSucceeded
		return
	}

	if job.Status.Failed != 0 {
		status.State = JobStateFailed
	}

	if len(job.Status.Conditions) > 0 {
		jobConditionTable, err = parseJobConditions(job.Status.Conditions)
		status.Message = err.Error()
	}

	pods, err := ib.clusterController.ListPods(
		context.Background(),
		ib.imageBuildingConfig.BuildNamespace,
		fmt.Sprintf("job-name=%s", kanikoJobName),
	)
	if err != nil && !kerrors.IsNotFound(err) {
		status.Message = err.Error()
		return
	}

	for _, pod := range pods.Items {
		if len(pod.Status.ContainerStatuses) > 0 {
			podContainerTable, podLastTerminationMessage,
				podLastTerminationReason = utils.ParsePodContainerStatuses(pod.Status.ContainerStatuses)
			status.Message = podLastTerminationReason
			break
		}
	}

	return
}

func (ib *imageBuilder) DeleteImageBuildingJob(
	projectName string,
	ensemblerName string,
	ensemblerID models.ID,
	versionID string,
) error {
	kanikoJobName := ib.nameGenerator.generateBuilderName(
		projectName,
		ensemblerName,
		ensemblerID,
		versionID,
	)
	job, err := ib.clusterController.GetJob(
		context.Background(),
		ib.imageBuildingConfig.BuildNamespace,
		kanikoJobName,
	)
	if err != nil {
		// Not found.
		return nil
	}
	// Delete job
	err = ib.clusterController.DeleteJob(context.Background(), ib.imageBuildingConfig.BuildNamespace, job.Name)
	return err
}

func (ib *imageBuilder) GetEnsemblerImage(
	project *mlp.Project,
	ensembler *models.PyFuncEnsembler,
) (EnsemblerImage, error) {
	imageName := ib.nameGenerator.generateDockerImageName(project.Name, ensembler.Name)
	imageExists, err := ib.checkIfImageExists(imageName, ensembler.RunID)
	if err != nil {
		return EnsemblerImage{}, err
	}

	imageRef := fmt.Sprintf("%s:%s", imageName, ensembler.RunID)

	image := EnsemblerImage{
		ProjectID:           models.ID(project.ID),
		EnsemblerID:         models.ID(ensembler.GetID()),
		EnsemblerRunnerType: ib.runnerType,
		ImageRef:            imageRef,
		Exists:              imageExists,
	}
	return image, nil
}

func parseJobConditions(jobConditions []apibatchv1.JobCondition) (string, error) {
	var err error

	jobConditionHeaders := []string{"TIMESTAMP", "TYPE", "REASON", "MESSAGE"}
	jobConditionRows := [][]string{}

	sort.Slice(jobConditions, func(i, j int) bool {
		return jobConditions[i].LastProbeTime.Before(&jobConditions[j].LastProbeTime)
	})

	for _, condition := range jobConditions {
		jobConditionRows = append(jobConditionRows, []string{
			condition.LastProbeTime.Format(time.RFC1123),
			string(condition.Type),
			condition.Reason,
			condition.Message,
		})

		err = errors.New(condition.Reason)
	}

	jobTable := utils.LogTable(jobConditionHeaders, jobConditionRows)
	return jobTable, err
}
