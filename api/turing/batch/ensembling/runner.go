package batchensembling

import (
	"time"

	mlp "github.com/gojek/mlp/client"
	batchrunner "github.com/gojek/turing/api/turing/batch/runner"
	"github.com/gojek/turing/api/turing/cluster/labeller"
	"github.com/gojek/turing/api/turing/imagebuilder"
	"github.com/gojek/turing/api/turing/log"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
)

var pageOne = 1

type ensemblingJobRunner struct {
	ensemblingController      EnsemblingController
	ensemblingJobService      service.EnsemblingJobService
	mlpService                service.MLPService
	imageBuilder              imagebuilder.ImageBuilder
	batchSize                 int
	maxRetryCount             int
	imageBuildTimeoutDuration time.Duration
}

// NewBatchEnsemblingJobRunner creates a new batch ensembling job runner
// This service controls the orchestration of batch ensembling jobs.
func NewBatchEnsemblingJobRunner(
	ensemblingController EnsemblingController,
	ensemblingJobService service.EnsemblingJobService,
	mlpService service.MLPService,
	imageBuilder imagebuilder.ImageBuilder,
	batchSize int,
	maxRetryCount int,
	imageBuildTimeoutDuration time.Duration,
) batchrunner.BatchJobRunner {
	return &ensemblingJobRunner{
		ensemblingController:      ensemblingController,
		ensemblingJobService:      ensemblingJobService,
		mlpService:                mlpService,
		imageBuilder:              imageBuilder,
		batchSize:                 batchSize,
		maxRetryCount:             maxRetryCount,
		imageBuildTimeoutDuration: imageBuildTimeoutDuration,
	}
}

func (r *ensemblingJobRunner) Run() {
	r.processJobs()
	r.updateStatus()
}

func (r *ensemblingJobRunner) updateStatus() {
	// Here we want to check all non terminal cases.
	// 1. JobRunning
	// 2. JobBuildingImage but with updated_at timeout.
	optionsJobRunning := service.EnsemblingJobListOptions{
		PaginationOptions: service.PaginationOptions{
			Page:     &pageOne,
			PageSize: &r.batchSize,
		},
		Statuses: []models.Status{
			models.JobRunning,
		},
	}
	r.processEnsemblingJobs(optionsJobRunning)

	// To ensure that we only pickup jobs that weren't picked up, we check the last updated at.
	// Two conditions that can happen:
	//
	// Case 1: Job is actually being watched by a live goroutine.
	//              <---imageBuildTimeoutDuration--->
	// ------------|-----------------|---------------|---------> time
	//          job start          now       timeout cut off
	//          (updated at)
	// We don't want to pick up this situation since the image building might take a long time.
	//
	// Case 2: job hanged or Turing API unexpectedly terminated.
	//              <---imageBuildTimeoutDuration--->
	// ------------|---------------------------------|------> time
	//           job start                          now
	//           (updated at)
	// Here Turing API might have crashed and we probably want to retry this situation.

	updatedAtBefore := time.Now().Add(r.imageBuildTimeoutDuration * -1)
	optionsJobBuildingImage := service.EnsemblingJobListOptions{
		PaginationOptions: service.PaginationOptions{
			Page:     &pageOne,
			PageSize: &r.batchSize,
		},
		Statuses: []models.Status{
			models.JobBuildingImage,
		},
		UpdatedAtBefore:    &updatedAtBefore,
		RetryCountLessThan: &r.maxRetryCount,
	}
	r.processEnsemblingJobs(optionsJobBuildingImage)
}

func (r *ensemblingJobRunner) processEnsemblingJobs(queryOptions service.EnsemblingJobListOptions) {
	ensemblingJobsPaginated, err := r.ensemblingJobService.List(queryOptions)
	if err != nil {
		log.Errorf("unable to query ensembling jobs", err)
		return
	}

	ensemblingJobs := ensemblingJobsPaginated.Results.([]*models.EnsemblingJob)
	for _, ensemblingJob := range ensemblingJobs {
		r.updateOneStatus(ensemblingJob)
	}
}

func (r *ensemblingJobRunner) updateOneStatus(ensemblingJob *models.EnsemblingJob) {
	// Consider that the application may terminate when processing halfway.
	// It's ok for it to be processed multiple times between multiple processes/goroutines
	// because they will have the same outcome.
	// If that happens, we should mark them for retry
	// i.e. in the general case, we set back to JobPending but bump retry count
	// If JobBuildingImage, we want to check if the image building has already been done
	//   -> if image building is active, we do not do anything
	//   -> else we just set to JobPending because it has hanged for some reason.
	// If JobRunning, check if the spark driver is running
	//   -> If running, do nothing
	//   -> If error, set it to JobFailed; we don't want to retry because spark
	//      has a retry mechanism
	//   -> If completed, mark as JobCompleted

	mlpProject, queryErr := r.mlpService.GetProject(ensemblingJob.ProjectID)
	if queryErr != nil {
		r.saveStatus(ensemblingJob, models.JobFailedBuildImage, queryErr.Error(), true)
		return
	}
	switch ensemblingJob.Status {
	case models.JobBuildingImage:
		r.processBuildingImage(ensemblingJob, mlpProject)
	case models.JobRunning:
		r.processJobRunning(ensemblingJob, mlpProject)
	}
}

func (r *ensemblingJobRunner) processJobRunning(
	ensemblingJob *models.EnsemblingJob,
	mlpProject *mlp.Project,
) {
	state, err := r.ensemblingController.GetStatus(mlpProject.Name, ensemblingJob)
	if err != nil {
		log.Errorf("Unable to get status of spark application: %v", err)
		return
	}

	if state == SparkApplicationStateUnknown {
		// Do nothing, just wait to see if something happens
		log.Warnf("Spark application state is unknown: %s", ensemblingJob.Name)
		return
	}

	if state == SparkApplicationStateCompleted {
		ensemblingJob.Status = models.JobCompleted
	} else if state == SparkApplicationStateFailed {
		ensemblingJob.Status = models.JobFailed
	}

	err = r.ensemblingJobService.Save(ensemblingJob)
	if err != nil {
		log.Errorf("Unable to save ensemblingJob %d: %v", ensemblingJob.ID, err)
		return
	}
}

func (r *ensemblingJobRunner) processBuildingImage(
	ensemblingJob *models.EnsemblingJob,
	mlpProject *mlp.Project,
) {
	status, err := r.imageBuilder.GetImageBuildingJobStatus(
		mlpProject.Name,
		ensemblingJob.InfraConfig.EnsemblerName,
		ensemblingJob.EnsemblerID,
	)

	if status == imagebuilder.JobStatusActive {
		// Do nothing
		return
	}

	// We retry on all other possible outcomes
	ensemblingJob.Status = models.JobPending
	ensemblingJob.RetryCount++
	saveErr := r.ensemblingJobService.Save(ensemblingJob)
	if saveErr != nil {
		log.Errorf("Unable to save ensemblingJob %d: %v", ensemblingJob.ID, err)
	}
}

func (r *ensemblingJobRunner) processJobs() {
	options := service.EnsemblingJobListOptions{
		PaginationOptions: service.PaginationOptions{
			Page:     &pageOne,
			PageSize: &r.batchSize,
		},
		Statuses: []models.Status{
			models.JobPending,
			models.JobFailedSubmission,
			models.JobFailedBuildImage,
		},
		RetryCountLessThan: &r.maxRetryCount,
	}
	ensemblingJobsPaginated, err := r.ensemblingJobService.List(options)
	if err != nil {
		log.Errorf("unable to query ensembling jobs: %v", err)
	}

	if ensemblingJobsPaginated.Results == nil {
		return
	}
	ensemblingJobs := ensemblingJobsPaginated.Results.([]*models.EnsemblingJob)
	for _, ensemblingJob := range ensemblingJobs {
		// Don't bother waiting, let the person calling the
		// BatchJobRunner interface decide how many to submit at once.
		go r.processOneEnsemblingJob(ensemblingJob)
	}
}

func (r *ensemblingJobRunner) processOneEnsemblingJob(ensemblingJob *models.EnsemblingJob) {
	ensemblingJob.Status = models.JobBuildingImage
	err := r.ensemblingJobService.Save(ensemblingJob)
	if err != nil {
		r.saveStatus(ensemblingJob, models.JobPending, err.Error(), true)
		return
	}

	mlpProject, queryErr := r.mlpService.GetProject(ensemblingJob.ProjectID)
	if queryErr != nil {
		r.saveStatus(ensemblingJob, models.JobFailedBuildImage, queryErr.Error(), true)
		return
	}

	// Build Image
	labels := r.buildLabels(ensemblingJob, mlpProject)
	imageRef, imageBuildErr := r.buildImage(ensemblingJob, mlpProject, labels)
	if imageBuildErr != nil {
		r.saveStatus(ensemblingJob, models.JobFailedBuildImage, imageBuildErr.Error(), true)
		return
	}

	// Submit to Kubernetes
	controllerError := r.ensemblingController.Create(
		&CreateEnsemblingJobRequest{
			EnsemblingJob: ensemblingJob,
			Labels:        labels,
			ImageRef:      imageRef,
			Namespace:     mlpProject.Name,
		},
	)
	if controllerError != nil {
		r.saveStatus(ensemblingJob, models.JobFailedSubmission, controllerError.Error(), true)
		return
	}

	r.saveStatus(ensemblingJob, models.JobRunning, "", false)
}

func (r *ensemblingJobRunner) buildLabels(
	ensemblingJob *models.EnsemblingJob,
	mlpProject *mlp.Project,
) map[string]string {
	rq := labeller.KubernetesLabelsRequest{
		Stream: mlpProject.Stream,
		Team:   mlpProject.Team,
		App:    ensemblingJob.InfraConfig.EnsemblerName,
	}

	return labeller.BuildLabels(rq)
}

func (r *ensemblingJobRunner) saveStatus(
	ensemblingJob *models.EnsemblingJob,
	status models.Status,
	errorMessage string,
	incrementRetry bool,
) {
	if incrementRetry {
		ensemblingJob.RetryCount++
	}
	ensemblingJob.Status = status
	ensemblingJob.Error = errorMessage
	err := r.ensemblingJobService.Save(ensemblingJob)
	if err != nil {
		log.Errorf("unable to save ensembling job %v", err)
	}
}

func (r *ensemblingJobRunner) buildImage(
	ensemblingJob *models.EnsemblingJob,
	mlpProject *mlp.Project,
	buildLabels map[string]string,
) (string, error) {
	request := imagebuilder.BuildImageRequest{
		ProjectName:  mlpProject.Name,
		ResourceName: ensemblingJob.InfraConfig.EnsemblerName,
		VersionID:    ensemblingJob.EnsemblerID,
		ArtifactURI:  ensemblingJob.InfraConfig.ArtifactURI,
		BuildLabels:  buildLabels,
	}
	return r.imageBuilder.BuildImage(request)
}
