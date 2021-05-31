package batchrunner

import (
	mlp "github.com/gojek/mlp/client"
	batchcontroller "github.com/gojek/turing/api/turing/batch/controller"
	"github.com/gojek/turing/api/turing/imagebuilder"
	"github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/gojek/turing/api/turing/log"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
)

type ensemblingJobRunner struct {
	ensemblingController batchcontroller.EnsemblingController
	ensemblingJobService service.EnsemblingJobService
	mlpService           service.MLPService
	imageBuilder         imagebuilder.ImageBuilder
	injectGojekLabels    bool
	environment          string
	batchSize            int
	maxRetryCount        int
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
) BatchJobRunner {
	return &ensemblingJobRunner{
		ensemblingController: ensemblingController,
		ensemblingJobService: ensemblingJobService,
		mlpService:           mlpService,
		imageBuilder:         imageBuilder,
		injectGojekLabels:    injectGojekLabels,
		environment:          environment,
		batchSize:            batchSize,
		maxRetryCount:        maxRetryCount,
	}
}

func (r *ensemblingJobRunner) Run() {
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
			models.JobFailed,
		},
		IsLocked:           &isLocked,
		RetryCountLessThan: &r.maxRetryCount,
	}
	ensemblingJobsPaginated, err := r.ensemblingJobService.List(options)
	if err != nil {
		log.Errorf("unable to query ensembling jobs", err)
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
		log.Errorf("unable to unlock ensembling job", err)
	}
}

func (r *ensemblingJobRunner) processOneEnsemblingJob(ensemblingJob *models.EnsemblingJob) {
	defer r.unlockJob(ensemblingJob)

	// lock ensembling job so others parallel processes don't take it
	ensemblingJob.IsLocked = true
	ensemblingJob.Status = models.JobBuildingImage
	err := r.ensemblingJobService.Save(ensemblingJob)
	if err != nil {
		r.saveStatus(ensemblingJob, models.JobPending, err.Error())
		return
	}

	mlpProject, queryErr := r.mlpService.GetProject(ensemblingJob.ProjectID)
	if queryErr != nil {
		r.saveStatus(ensemblingJob, models.JobFailedBuildImage, queryErr.Error())
		return
	}

	// Build Image
	labels := r.buildLabels(ensemblingJob, mlpProject)
	imageRef, imageBuildErr := r.buildImage(ensemblingJob, mlpProject, labels)
	if imageBuildErr != nil {
		r.saveStatus(ensemblingJob, models.JobFailedBuildImage, imageBuildErr.Error())
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
		r.saveStatus(ensemblingJob, models.JobFailedSubmission, controllerError.Error())
		return
	}

	r.saveStatus(ensemblingJob, models.JobRunning, "")
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
) {
	if !status.IsSuccessful() {
		ensemblingJob.RetryCount++
	}
	ensemblingJob.Status = status
	ensemblingJob.Error = errorMessage
	err := r.ensemblingJobService.Save(ensemblingJob)
	if err != nil {
		log.Errorf("unable to save ensembling job", err)
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
		VersionID:   int(ensemblingJob.EnsemblerID),
		ArtifactURI: ensemblingJob.InfraConfig.ArtifactURI,
		BuildLabels: buildLabels,
	}
	return r.imageBuilder.BuildImage(request)
}
