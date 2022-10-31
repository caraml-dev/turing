package service

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	mlp "github.com/gojek/mlp/api/client"
	"gorm.io/gorm"

	"github.com/caraml-dev/turing/api/turing/batch"
	"github.com/caraml-dev/turing/api/turing/cluster/labeller"
	"github.com/caraml-dev/turing/api/turing/cluster/servicebuilder"
	"github.com/caraml-dev/turing/api/turing/config"
	openapi "github.com/caraml-dev/turing/api/turing/generated"
	logger "github.com/caraml-dev/turing/api/turing/log"
	"github.com/caraml-dev/turing/api/turing/models"
)

const (
	// SparkHomeFolder is the home folder of the spark user in the Docker container
	// used in engines/batch-ensembler/Dockerfile
	SparkHomeFolder = "/home/spark"

	kubernetesSparkRoleLabel         = "spark-role"
	kubernetesSparkRoleDriverValue   = "driver"
	kubernetesSparkRoleExecutorValue = "executor"
)

var (
	// EnsemblerFolder is the folder created by the Turing SDK that contains
	// the ensembler dependencies and pickled Python files.
	EnsemblerFolder = "ensembler"

	loggingPodPostfixesInSearch = map[string]string{
		batch.DriverPodType:       ".*-driver",
		batch.ExecutorPodType:     ".*-exec-.*",
		batch.ImageBuilderPodType: "",
	}
)

// EnsemblingJobFindByIDOptions contains the options allowed when finding ensembling jobs.
type EnsemblingJobFindByIDOptions struct {
	ProjectID *models.ID
}

// EnsemblingJobListOptions holds query parameters for EnsemblersService.List method.
type EnsemblingJobListOptions struct {
	PaginationOptions
	ProjectID          *models.ID      `schema:"project_id" validate:"required"`
	EnsemblerID        *models.ID      `schema:"ensembler_id"`
	Statuses           []models.Status `schema:"status"`
	Search             *string         `schema:"search"`
	RetryCountLessThan *int            `schema:"-"`
	UpdatedAtBefore    *time.Time      `schema:"-"`
}

// EnsemblingJobService is the data access object for the EnsemblingJob from the db.
type EnsemblingJobService interface {
	Save(ensemblingJob *models.EnsemblingJob) error
	Delete(ensemblingJob *models.EnsemblingJob) error
	FindByID(
		id models.ID,
		options EnsemblingJobFindByIDOptions,
	) (*models.EnsemblingJob, error)
	List(options EnsemblingJobListOptions) (*PaginatedResults, error)
	CreateEnsemblingJob(
		job *models.EnsemblingJob,
		projectID models.ID,
		ensembler *models.PyFuncEnsembler,
	) (*models.EnsemblingJob, error)
	MarkEnsemblingJobForTermination(ensemblingJob *models.EnsemblingJob) error
	GetNamespaceByComponent(componentType string, project *mlp.Project) string
	GetDefaultEnvironment() string
	CreatePodLabelSelector(ensemblerName, componentType string) []LabelSelector
	FormatLoggingURL(ensemblerName string, namespace string, componentType string) (string, error)
}

// NewEnsemblingJobService creates a new ensembling job service
func NewEnsemblingJobService(
	db *gorm.DB,
	defaultEnvironment string,
	imageBuilderNamespace string,
	loggingURLFormat *string,
	dashboardURLFormat *string,
	defaultConfig config.DefaultEnsemblingJobConfigurations,
	mlpService MLPService,
) EnsemblingJobService {
	var loggingURLTemplate *template.Template
	var err error
	if loggingURLFormat != nil {
		loggingURLTemplate, err = template.New("dashboardTemplate").Parse(*loggingURLFormat)
		if err != nil {
			logger.Warnf("error parsing ensembling logging url template: %s", err)
		}
	}

	var dashboardTemplate *template.Template
	if dashboardURLFormat != nil {
		dashboardTemplate, err = template.New("dashboardTemplate").Parse(*dashboardURLFormat)
		if err != nil {
			logger.Warnf("error parsing ensembling dashboard template, values will be nil: %s", err.Error())
		}
	}

	return &ensemblingJobService{
		db:                    db,
		defaultEnvironment:    defaultEnvironment,
		imageBuilderNamespace: imageBuilderNamespace,
		dashboardURLTemplate:  dashboardTemplate,
		loggingURLTemplate:    loggingURLTemplate,
		defaultConfig:         defaultConfig,
		mlpService:            mlpService,
	}
}

type ensemblingJobService struct {
	db                    *gorm.DB
	defaultEnvironment    string
	imageBuilderNamespace string
	dashboardURLTemplate  *template.Template
	loggingURLTemplate    *template.Template
	defaultConfig         config.DefaultEnsemblingJobConfigurations
	mlpService            MLPService
}

// Save the given router to the db. Updates the existing record if already exists
func (s *ensemblingJobService) Save(ensemblingJob *models.EnsemblingJob) error {
	return s.db.Save(ensemblingJob).Error
}

func (s *ensemblingJobService) Delete(ensemblingJob *models.EnsemblingJob) error {
	return s.db.Delete(ensemblingJob).Error
}

func (s *ensemblingJobService) FindByID(
	id models.ID,
	options EnsemblingJobFindByIDOptions,
) (*models.EnsemblingJob, error) {
	query := s.db.Where("id = ?", id)

	if options.ProjectID != nil {
		query = query.Where("project_id = ?", options.ProjectID)
	}

	var ensemblingJob models.EnsemblingJob
	result := query.First(&ensemblingJob)

	if err := result.Error; err != nil {
		return nil, err
	}

	project, err := s.mlpService.GetProject(ensemblingJob.ProjectID)
	if err != nil {
		return nil, err
	}

	url, err := s.generateMonitoringURL(&ensemblingJob, project)
	if err != nil {
		return nil, err
	}
	ensemblingJob.MonitoringURL = url

	return &ensemblingJob, nil
}

func (s *ensemblingJobService) List(options EnsemblingJobListOptions) (*PaginatedResults, error) {
	var results []*models.EnsemblingJob
	var count int64

	query := s.db
	if options.ProjectID != nil {
		query = query.Where("project_id = ?", options.ProjectID)
	}

	if options.EnsemblerID != nil {
		query = query.Where("ensembler_id = ?", options.EnsemblerID)
	}

	if options.Search != nil && len(*options.Search) > 0 {
		query = query.Where("name ilike ?", fmt.Sprintf("%%%s%%", *options.Search))
	}

	if options.Statuses != nil {
		query = query.Where("status IN (?)", options.Statuses)
	}

	if options.RetryCountLessThan != nil {
		query = query.Where("retry_count < ?", options.RetryCountLessThan)
	}

	if options.UpdatedAtBefore != nil {
		query = query.Where("updated_at < ?", options.UpdatedAtBefore)
	}

	query.Model(&results).Count(&count)
	result := query.
		Scopes(PaginationScope(options.PaginationOptions)).
		Find(&results)

	if err := result.Error; err != nil {
		return nil, err
	}

	for _, r := range results {
		project, err := s.mlpService.GetProject(r.ProjectID)
		if err != nil {
			return nil, err
		}

		url, err := s.generateMonitoringURL(r, project)
		if err != nil {
			return nil, err
		}
		r.MonitoringURL = url
	}

	paginatedResults := createPaginatedResults(options.PaginationOptions, int(count), results)
	return paginatedResults, nil
}

func getEnsemblerDirectory(ensembler *models.PyFuncEnsembler) string {
	// Ensembler URI will be a local directory
	// Dockerfile will build copy the artifact into the local directory.
	// See engines/batch-ensembler/app.Dockerfile
	splitURI := strings.Split(ensembler.ArtifactURI, "/")
	return fmt.Sprintf(
		"%s/%s/ensembler",
		SparkHomeFolder,
		splitURI[len(splitURI)-1],
	)
}

// EnsemblingMonitoringVariables the values supplied to BatchEnsemblingConfig.MonitoringURLTemplate
type EnsemblingMonitoringVariables struct {
	// Project is the MLP Project associated with the batch ensembler
	Project string
	// Job is the name of the ensembling job.
	Job string
}

func (s *ensemblingJobService) generateMonitoringURL(job *models.EnsemblingJob, project *mlp.Project) (string, error) {
	if s.dashboardURLTemplate == nil {
		return "", nil
	}

	values := EnsemblingMonitoringVariables{
		Project: project.Name,
		Job:     job.Name,
	}

	var b bytes.Buffer
	err := s.dashboardURLTemplate.Execute(&b, values)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

// CreateEnsemblingJob creates an ensembling job.
func (s *ensemblingJobService) CreateEnsemblingJob(
	job *models.EnsemblingJob,
	projectID models.ID,
	ensembler *models.PyFuncEnsembler,
) (*models.EnsemblingJob, error) {
	job.ProjectID = projectID
	job.EnvironmentName = s.defaultEnvironment

	job.JobConfig.Spec.Ensembler.Uri = getEnsemblerDirectory(ensembler)
	job.InfraConfig.RunId = &ensembler.RunID
	job.InfraConfig.ArtifactUri = &ensembler.ArtifactURI
	job.InfraConfig.EnsemblerName = &ensembler.Name

	project, err := s.mlpService.GetProject(projectID)
	if err != nil {
		return nil, err
	}

	s.mergeDefaultConfigurations(job)

	// Save ensembling job
	if err := s.Save(job); err != nil {
		return nil, err
	}

	url, err := s.generateMonitoringURL(job, project)
	if err != nil {
		return nil, err
	}
	job.MonitoringURL = url

	return job, nil
}

func (s *ensemblingJobService) MarkEnsemblingJobForTermination(job *models.EnsemblingJob) error {
	job.Status = models.JobTerminating
	if err := s.Save(job); err != nil {
		return err
	}

	project, err := s.mlpService.GetProject(job.ProjectID)
	if err != nil {
		return err
	}

	url, err := s.generateMonitoringURL(job, project)
	if err != nil {
		return err
	}
	job.MonitoringURL = url

	return nil
}

func (s *ensemblingJobService) GetNamespaceByComponent(componentType string, project *mlp.Project) string {
	if componentType == batch.ImageBuilderPodType {
		return s.imageBuilderNamespace
	}
	return servicebuilder.GetNamespace(project)
}

func (s *ensemblingJobService) GetDefaultEnvironment() string {
	return s.defaultEnvironment
}

func (s *ensemblingJobService) CreatePodLabelSelector(ensemblerName, componentType string) []LabelSelector {
	labelSelector := []LabelSelector{
		{
			Key:   labeller.GetLabelName(labeller.AppLabel),
			Value: ensemblerName,
		},
	}

	if componentType == batch.DriverPodType {
		labelSelector = append(labelSelector, LabelSelector{
			Key:   kubernetesSparkRoleLabel,
			Value: kubernetesSparkRoleDriverValue,
		})
	} else if componentType == batch.ExecutorPodType {
		labelSelector = append(labelSelector, LabelSelector{
			Key:   kubernetesSparkRoleLabel,
			Value: kubernetesSparkRoleExecutorValue,
		})
	}

	return labelSelector
}

type ensemblingLogURLValues struct {
	PodName   string
	Namespace string
}

func (s *ensemblingJobService) FormatLoggingURL(ensemblerName, namespace, componentType string) (string, error) {
	if s.loggingURLTemplate == nil {
		// Not configured properly or not configured
		return "", nil
	}

	podName := fmt.Sprintf("%s%s", ensemblerName, loggingPodPostfixesInSearch[componentType])
	v := ensemblingLogURLValues{
		PodName:   podName,
		Namespace: namespace,
	}

	var b bytes.Buffer
	err := s.loggingURLTemplate.Execute(&b, v)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func (s *ensemblingJobService) mergeDefaultConfigurations(job *models.EnsemblingJob) {
	if job.JobConfig.Metadata == nil {
		job.JobConfig.Metadata = &openapi.EnsemblingJobMeta{
			Annotations: make(map[string]string),
		}
	}

	// Only apply default if key does not exist, we should respect the users annotation override.
	for key, value := range s.defaultConfig.SparkConfigAnnotations {
		if _, ok := job.JobConfig.Metadata.Annotations[key]; !ok {
			job.JobConfig.Metadata.Annotations[key] = value
		}
	}

	if !job.InfraConfig.Resources.IsSet() {
		configCopy := s.defaultConfig.BatchEnsemblingJobResources
		job.InfraConfig.Resources.Set(&configCopy)
		return
	}

	resources := job.InfraConfig.GetResources()

	if resources.DriverCpuRequest == nil {
		resources.DriverCpuRequest = s.defaultConfig.BatchEnsemblingJobResources.DriverCpuRequest
	}

	if resources.DriverMemoryRequest == nil {
		resources.DriverMemoryRequest = s.defaultConfig.BatchEnsemblingJobResources.DriverMemoryRequest
	}

	if resources.ExecutorReplica == nil {
		resources.ExecutorReplica = s.defaultConfig.BatchEnsemblingJobResources.ExecutorReplica
	}

	if resources.ExecutorCpuRequest == nil {
		resources.ExecutorCpuRequest = s.defaultConfig.BatchEnsemblingJobResources.ExecutorCpuRequest
	}

	if resources.ExecutorMemoryRequest == nil {
		resources.ExecutorMemoryRequest = s.defaultConfig.BatchEnsemblingJobResources.ExecutorMemoryRequest
	}

	// Required as it returns a copy of resources and not the pointer address
	job.InfraConfig.SetResources(resources)
}
