package batchensembling

import (
	"time"

	mlp "github.com/gojek/mlp/api/client"

	batchrunner "github.com/caraml-dev/turing/api/turing/batch/runner"
	"github.com/caraml-dev/turing/api/turing/cluster/labeller"
	"github.com/caraml-dev/turing/api/turing/imagebuilder"
	"github.com/caraml-dev/turing/api/turing/log"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
)

var pageOne = 1

type ensemblingJobRunner struct {
	ensemblingController           EnsemblingController
	ensemblingJobService           service.EnsemblingJobService
	ensemblersService              service.EnsemblersService
	mlpService                     service.MLPService
	imageBuilder                   imagebuilder.ImageBuilder
	recordsToProcessInOneIteration int
	maxRetryCount                  int
	imageBuildTimeoutDuration      time.Duration
	timeInterval                   time.Duration
}

// NewBatchEnsemblingJobRunner creates a new batch ensembling job runner
// This service controls the orchestration of batch ensembling jobs.
func NewBatchEnsemblingJobRunner(
	ensemblingController EnsemblingController,
	ensemblingJobService service.EnsemblingJobService,
	ensemblersService service.EnsemblersService,
	mlpService service.MLPService,
	imageBuilder imagebuilder.ImageBuilder,
	recordsToProcessInOneIteration int,
	maxRetryCount int,
	imageBuildTimeoutDuration time.Duration,
	timeInterval time.Duration,
) batchrunner.BatchJobRunner {
	return &ensemblingJobRunner{
		ensemblingController:           ensemblingController,
		ensemblingJobService:           ensemblingJobService,
		ensemblersService:              ensemblersService,
		mlpService:                     mlpService,
		imageBuilder:                   imageBuilder,
		recordsToProcessInOneIteration: recordsToProcessInOneIteration,
		maxRetryCount:                  maxRetryCount,
		imageBuildTimeoutDuration:      imageBuildTimeoutDuration,
		timeInterval:                   timeInterval,
	}
}

func (r *ensemblingJobRunner) GetInterval() time.Duration {
	return r.timeInterval
}

func (r *ensemblingJobRunner) Run() {
	r.processJobs()
	r.updateStatus()
}

func (r *ensemblingJobRunner) updateStatus() {
	// Here we want to check all non terminal cases.
	// 1. JobRunning
	// 2. JobBuildingImage but with updated_at timeout.
	// 3. JobTerminating

	// JobRunning
	optionsJobRunning := service.EnsemblingJobListOptions{
		PaginationOptions: service.PaginationOptions{
			Page:     &pageOne,
			PageSize: &r.recordsToProcessInOneIteration,
		},
		Statuses: []models.Status{
			models.JobRunning,
		},
	}
	r.processEnsemblingJobs(optionsJobRunning)

	// JobBuildingImage
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
			PageSize: &r.recordsToProcessInOneIteration,
		},
		Statuses: []models.Status{
			models.JobBuildingImage,
		},
		UpdatedAtBefore:    &updatedAtBefore,
		RetryCountLessThan: &r.maxRetryCount,
	}
	r.processEnsemblingJobs(optionsJobBuildingImage)

	// JobTerminating
	optionsJobTerminatingImage := service.EnsemblingJobListOptions{
		PaginationOptions: service.PaginationOptions{
			Page:     &pageOne,
			PageSize: &r.recordsToProcessInOneIteration,
		},
		Statuses: []models.Status{
			models.JobTerminating,
		},
	}
	r.processEnsemblingJobs(optionsJobTerminatingImage)
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
		r.saveStatusOrTerminate(ensemblingJob, mlpProject, models.JobFailedBuildImage, queryErr.Error(), true)
		return
	}
	switch ensemblingJob.Status {
	case models.JobBuildingImage:
		r.processBuildingImage(ensemblingJob, mlpProject)
	case models.JobRunning:
		r.processJobRunning(ensemblingJob, mlpProject)
	case models.JobTerminating:
		r.terminateJob(ensemblingJob, mlpProject)
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

	// Check if up for termination
	if shouldTerminate := r.terminateIfRequired(ensemblingJob.ID, mlpProject); shouldTerminate {
		return
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
		*ensemblingJob.InfraConfig.EnsemblerName,
		ensemblingJob.EnsemblerID,
		*ensemblingJob.InfraConfig.RunId,
	)

	if status == imagebuilder.JobStatusActive {
		// Do nothing
		return
	}

	// Check if up for termination
	if shouldTerminate := r.terminateIfRequired(ensemblingJob.ID, mlpProject); shouldTerminate {
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
			PageSize: &r.recordsToProcessInOneIteration,
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
		return
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
	mlpProject, queryErr := r.mlpService.GetProject(ensemblingJob.ProjectID)
	if queryErr != nil {
		r.saveStatusOrTerminate(
			ensemblingJob,
			mlpProject,
			models.JobFailedBuildImage,
			queryErr.Error(),
			true,
		)
		return
	}

	// Check if termination required.
	if shouldTerminate := r.terminateIfRequired(ensemblingJob.ID, mlpProject); shouldTerminate {
		return
	}

	ensemblingJob.Status = models.JobBuildingImage
	err := r.ensemblingJobService.Save(ensemblingJob)
	if err != nil {
		r.saveStatusOrTerminate(ensemblingJob, mlpProject, models.JobPending, err.Error(), true)
		return
	}

	// Get ensembler
	ensembler, err := r.ensemblersService.FindByID(
		ensemblingJob.EnsemblerID,
		service.EnsemblersFindByIDOptions{ProjectID: &ensemblingJob.ProjectID},
	)
	if err != nil {
		r.saveStatusOrTerminate(ensemblingJob, mlpProject, models.JobPending, err.Error(), true)
		return
	}

	// Get base image tag
	var baseImageTag string
	if pyfuncEnsembler, ok := ensembler.(*models.PyFuncEnsembler); ok {
		baseImageTag = pyfuncEnsembler.PythonVersion
	} else {
		r.saveStatusOrTerminate(ensemblingJob, mlpProject, models.JobPending, err.Error(), true)
		return
	}

	// Build Image
	labels := r.buildLabels(ensemblingJob, mlpProject)
	imageRef, imageBuildErr := r.buildImage(ensemblingJob, mlpProject, labels, baseImageTag)

	if imageBuildErr != nil {
		// Here unfortunately we have to wait till the image building process has
		// finished/errored before we can delete the job
		// It's still worth cleaning up even after the job is done.
		r.saveStatusOrTerminate(
			ensemblingJob,
			mlpProject,
			models.JobFailedBuildImage,
			imageBuildErr.Error(),
			true,
		)
		return
	}

	// Before submitting to Kubernetes Spark, check if job should terminate
	if shouldTerminate := r.terminateIfRequired(ensemblingJob.ID, mlpProject); shouldTerminate {
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
		r.saveStatusOrTerminate(
			ensemblingJob,
			mlpProject,
			models.JobFailedSubmission,
			controllerError.Error(),
			true,
		)
		return
	}

	r.saveStatusOrTerminate(ensemblingJob, mlpProject, models.JobRunning, "", false)
}

func (r *ensemblingJobRunner) terminateJob(ensemblingJob *models.EnsemblingJob, mlpProject *mlp.Project) {
	// Delete building image job
	jobErr := r.imageBuilder.DeleteImageBuildingJob(
		mlpProject.Name,
		*ensemblingJob.InfraConfig.EnsemblerName,
		ensemblingJob.EnsemblerID,
		*ensemblingJob.InfraConfig.RunId,
	)
	// Delete spark resource
	sparkErr := r.ensemblingController.Delete(mlpProject.Name, ensemblingJob)

	if jobErr != nil || sparkErr != nil {
		return
	}

	// Delete record
	err := r.ensemblingJobService.Delete(ensemblingJob)
	if err != nil {
		log.Errorf("unable to delete ensembling job %v", err)
	}
}

// terminateIfRequired returns true if the process should drop what it is doing.
func (r *ensemblingJobRunner) terminateIfRequired(ensemblingJobID models.ID, mlpProject *mlp.Project) bool {
	ensemblingJob, err := r.ensemblingJobService.FindByID(ensemblingJobID, service.EnsemblingJobFindByIDOptions{})

	if err != nil {
		// Job already deleted, must not allow job to be revived.
		// Because of the async activities, the job could have been deleted.
		// There might be a chance where the job has been already deleted but
		// a stale record was processing.
		// We return true here because this will happen.
		return true
	}

	if ensemblingJob.Status == models.JobTerminating {
		r.terminateJob(ensemblingJob, mlpProject)
		return true
	}
	return false
}

func (r *ensemblingJobRunner) buildLabels(
	ensemblingJob *models.EnsemblingJob,
	mlpProject *mlp.Project,
) map[string]string {
	rq := labeller.KubernetesLabelsRequest{
		Stream: mlpProject.Stream,
		Team:   mlpProject.Team,
		App:    *ensemblingJob.InfraConfig.EnsemblerName,
	}

	return labeller.BuildLabels(rq)
}

func (r *ensemblingJobRunner) saveStatusOrTerminate(
	ensemblingJob *models.EnsemblingJob,
	mlpProject *mlp.Project,
	status models.Status,
	errorMessage string,
	incrementRetry bool,
) bool {
	if shouldTerminate := r.terminateIfRequired(ensemblingJob.ID, mlpProject); shouldTerminate {
		return true
	}
	if incrementRetry {
		ensemblingJob.RetryCount++
	}
	ensemblingJob.Status = status
	ensemblingJob.Error = errorMessage
	err := r.ensemblingJobService.Save(ensemblingJob)
	if err != nil {
		log.Errorf("unable to save ensembling job %v", err)
	}

	return false
}

func (r *ensemblingJobRunner) buildImage(
	ensemblingJob *models.EnsemblingJob,
	mlpProject *mlp.Project,
	buildLabels map[string]string,
	baseImageTag string,
) (string, error) {
	request := imagebuilder.BuildImageRequest{
		ProjectName:     mlpProject.Name,
		ResourceName:    *ensemblingJob.InfraConfig.EnsemblerName,
		ResourceID:      ensemblingJob.EnsemblerID,
		VersionID:       *ensemblingJob.InfraConfig.RunId,
		ArtifactURI:     *ensemblingJob.InfraConfig.ArtifactUri,
		BuildLabels:     buildLabels,
		EnsemblerFolder: service.EnsemblerFolder,
		BaseImageRefTag: baseImageTag,
	}
	return r.imageBuilder.BuildImage(request)
}
