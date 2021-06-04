package batchrunner

import (
	"time"

	mlp "github.com/gojek/mlp/client"
	batchcontroller "github.com/gojek/turing/api/turing/batch/controller"
	"github.com/gojek/turing/api/turing/imagebuilder"
	"github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/gojek/turing/api/turing/log"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
)

type ensemblingJobRunner struct {
	ensemblingController      batchcontroller.EnsemblingController
	ensemblingJobService      service.EnsemblingJobService
	mlpService                service.MLPService
	imageBuilder              imagebuilder.ImageBuilder
	injectGojekLabels         bool
	environment               string
	batchSize                 int
	maxRetryCount             int
	imageBuildTimeoutDuration time.Duration
}

var (
	labelTeamName         = "gojek.com/team"
	labelStreamName       = "gojek.com/stream"
	labelAppName          = "gojek.com/app"
	labelEnvironment      = "gojek.com/environment"
	labelOrchestratorName = "gojek.com/orchestrator"
	valueOrchestratorName = "turing"
)

// NewBatchEnsemblingJobRunner creates a new batch ensembling job
// This service controls the orchestration of batch ensembling jobs.
func NewBatchEnsemblingJobRunner(
	ensemblingController batchcontroller.EnsemblingController,
	ensemblingJobService service.EnsemblingJobService,
	mlpService service.MLPService,
	imageBuilder imagebuilder.ImageBuilder,
	injectGojekLabels bool,
	environment string,
	batchSize int,
	maxRetryCount int,
	imageBuildTimeoutDuration time.Duration,
) BatchJobRunner {
	return &ensemblingJobRunner{
		ensemblingController:      ensemblingController,
		ensemblingJobService:      ensemblingJobService,
		mlpService:                mlpService,
		imageBuilder:              imageBuilder,
		injectGojekLabels:         injectGojekLabels,
		environment:               environment,
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
	// i.e. JobBuildingImage but with updated_at timeout
	//		JobRunning
	optionsJobRunning := service.EnsemblingJobListOptions{
		PaginationOptions: service.PaginationOptions{
			Page:     testutils.NullableInt(1),
			PageSize: &r.batchSize,
		},
		Statuses: []models.Status{
			models.JobRunning,
		},
	}
	r.processEnsemblingJobs(optionsJobRunning)

	// To ensure that we only pickup jobs that weren't picked up, we check the last updated at.
	updatedAtAfter := time.Now().Add(r.imageBuildTimeoutDuration * -1)
	optionsJobBuildingImage := service.EnsemblingJobListOptions{
		PaginationOptions: service.PaginationOptions{
			Page:     testutils.NullableInt(1),
			PageSize: &r.batchSize,
		},
		Statuses: []models.Status{
			models.JobBuildingImage,
		},
		UpdatedAtAfter:     &updatedAtAfter,
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
	// So we only check locked state, it's ok for it to be processed multiple times
	// between multiple instances because they will have the same outcome.
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
	// For all of these cases, we have to unlock the record.

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

	if state == batchcontroller.SparkApplicationStateUnknown {
		// Do nothing, just wait to see if something happens
		log.Warnf("Spark application state is unknown: %s", ensemblingJob.Name)
		return
	}

	if state == batchcontroller.SparkApplicationStateCompleted {
		ensemblingJob.Status = models.JobCompleted
	} else if state == batchcontroller.SparkApplicationStateFailed {
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
	ensemblingJob.IsLocked = false
	ensemblingJob.RetryCount++
	saveErr := r.ensemblingJobService.Save(ensemblingJob)
	if saveErr != nil {
		log.Errorf("Unable to save ensemblingJob %d: %v", ensemblingJob.ID, err)
	}
}

func (r *ensemblingJobRunner) processJobs() {
	isLocked := false
	options := service.EnsemblingJobListOptions{
		PaginationOptions: service.PaginationOptions{
			Page:     testutils.NullableInt(1),
			PageSize: &r.batchSize,
		},
		Statuses: []models.Status{
			models.JobPending,
			models.JobFailedSubmission,
			models.JobFailedBuildImage,
		},
		IsLocked:           &isLocked,
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

func (r *ensemblingJobRunner) unlockJob(ensemblingJob *models.EnsemblingJob) {
	ensemblingJob.IsLocked = false
	err := r.ensemblingJobService.Save(ensemblingJob)
	if err != nil {
		log.Errorf("unable to unlock ensembling job %v", err)
	}
}

func (r *ensemblingJobRunner) processOneEnsemblingJob(ensemblingJob *models.EnsemblingJob) {
	defer r.unlockJob(ensemblingJob)

	// lock ensembling job so others parallel processes don't take it
	ensemblingJob.IsLocked = true
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
		&batchcontroller.CreateEnsemblingJobRequest{
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
	buildLabels := make(map[string]string)
	if r.injectGojekLabels {
		buildLabels[labelTeamName] = mlpProject.Team
		buildLabels[labelStreamName] = mlpProject.Stream
		buildLabels[labelAppName] = ensemblingJob.InfraConfig.EnsemblerName
		buildLabels[labelEnvironment] = r.environment
		buildLabels[labelOrchestratorName] = valueOrchestratorName
	}

	return buildLabels
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
		ProjectName: mlpProject.Name,
		ModelName:   ensemblingJob.InfraConfig.EnsemblerName,
		VersionID:   ensemblingJob.EnsemblerID,
		ArtifactURI: ensemblingJob.InfraConfig.ArtifactURI,
		BuildLabels: buildLabels,
	}
	return r.imageBuilder.BuildImage(request)
}
