package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
)

const (
	sparkHomeFolder string = "/home/spark"
)

// EnsemblingJobFindByIDOptions contains the options allowed when finding ensembling jobs.
type EnsemblingJobFindByIDOptions struct {
	ProjectID *models.ID
}

// EnsemblingJobListOptions holds query parameters for EnsemblersService.List method.
type EnsemblingJobListOptions struct {
	PaginationOptions
	ProjectID *models.ID
	Status    models.Status
}

// EnsemblingJobService is the data access object for the EnsemblingJob from the db.
type EnsemblingJobService interface {
	Save(ensembler *models.EnsemblingJob) error
	FindByID(id models.ID, options EnsemblingJobFindByIDOptions) (*models.EnsemblingJob, error)
	List(options EnsemblingJobListOptions) (*PaginatedResults, error)
	GetDefaultEnvironment() string
	GenerateDefaultJobName(ensemblerName string) string
	GetEnsemblerDirectory(ensembler models.EnsemblerLike) (string, error)
	GetArtifactURI(ensembler models.EnsemblerLike) (string, error)
}

// NewEnsemblingJobService creates a new ensembling job service
func NewEnsemblingJobService(db *gorm.DB, defaultEnvironment string) EnsemblingJobService {
	return &ensemblingJobService{
		db:                 db,
		defaultEnvironment: defaultEnvironment,
	}
}

type ensemblingJobService struct {
	db                 *gorm.DB
	defaultEnvironment string
}

// Save the given router to the db. Updates the existing record if already exists
func (s *ensemblingJobService) Save(ensemblingJob *models.EnsemblingJob) error {
	return s.db.Save(ensemblingJob).Error
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
		return &ensemblingJob, err
	}

	return &ensemblingJob, nil
}

func (s *ensemblingJobService) List(options EnsemblingJobListOptions) (*PaginatedResults, error) {
	var results []*models.EnsemblingJob
	var count int
	done := make(chan bool, 1)

	query := s.db
	if options.ProjectID != nil {
		query = query.Where("project_id = ?", options.ProjectID)
	}

	if options.Status != "" {
		query = query.Where("status = ?", options.Status)
	}

	go func() {
		query.Model(&results).Count(&count)
		done <- true
	}()

	result := query.
		Scopes(PaginationScope(options.PaginationOptions)).
		Find(&results)
	<-done

	if err := result.Error; err != nil {
		return nil, err
	}

	paginatedResults := CreatePaginatedResults(options.PaginationOptions, count, results)
	return paginatedResults, nil
}

func (s *ensemblingJobService) GetDefaultEnvironment() string {
	return s.defaultEnvironment
}

func (ensemblingJobService) GenerateDefaultJobName(ensemblerName string) string {
	return fmt.Sprintf("%s: %s", ensemblerName, time.Now().Format(time.RFC3339))
}

// GetEnsemblerDirectory gets the ensembler directory that local to the container's local directory
func (ensemblingJobService) GetEnsemblerDirectory(ensembler models.EnsemblerLike) (string, error) {
	// Ensembler URI will be a local directory
	// Dockerfile will build copy the artifact into the local directory.
	// See engines/batch-ensembler/app.Dockerfile
	switch v := ensembler.(type) {
	case *models.PyFuncEnsembler:
		pyFuncEnsembler := ensembler.(*models.PyFuncEnsembler)
		splitURI := strings.Split(pyFuncEnsembler.ArtifactURI, "/")
		return fmt.Sprintf(
			"%s/%s",
			sparkHomeFolder,
			splitURI[len(splitURI)-1],
		), nil
	default:
		return "", fmt.Errorf("only pyfunc ensemblers are supported for now, given %T", v)
	}
}

// GetEnsemblerDirectory gets the ensembler directory that local to the container's local directory
func (ensemblingJobService) GetArtifactURI(ensembler models.EnsemblerLike) (string, error) {
	// Ensembler URI will be a local directory
	// Dockerfile will build copy the artifact into the local directory.
	// See engines/batch-ensembler/app.Dockerfile
	switch v := ensembler.(type) {
	case *models.PyFuncEnsembler:
		pyFuncEnsembler := ensembler.(*models.PyFuncEnsembler)
		return pyFuncEnsembler.ArtifactURI, nil
	default:
		return "", fmt.Errorf("only pyfunc ensemblers are supported for now, given %T", v)
	}
}
