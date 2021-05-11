package service

import (
	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
)

// EnsemblingJobFindByIDOptions contains the options allowed when finding ensembling jobs.
type EnsemblingJobFindByIDOptions struct {
	ProjectID *models.ID
}

// EnsemblingJobService is the data access object for the EnsemblingJob from the db.
type EnsemblingJobService interface {
	Save(ensembler *models.EnsemblingJob) error
	FindByID(id models.ID, options EnsemblingJobFindByIDOptions) (*models.EnsemblingJob, error)
	FindPendingJobs(limit int) ([]*models.EnsemblingJob, error)
	UpdateJobStatus(id models.ID, status models.State, errString string) error
}

// NewEnsemblingJobService creates a new ensembling job service
func NewEnsemblingJobService(db *gorm.DB) EnsemblingJobService {
	return &ensemblingJobService{db: db}
}

type ensemblingJobService struct {
	db *gorm.DB
}

// Save the given router to the db. Updates the existing record if already exists
func (s *ensemblingJobService) Save(ensemblingJob *models.EnsemblingJob) error {
	return s.db.Save(ensemblingJob).Error
}

func (s *ensemblingJobService) FindByID(id models.ID, options EnsemblingJobFindByIDOptions) (*models.EnsemblingJob, error) {
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

func (s *ensemblingJobService) FindPendingJobs(limit int) ([]*models.EnsemblingJob, error) {
	query := s.db.Order("id").Limit(limit).Where("status = ?", models.JobPending)
	var ensemblingJobs []*models.EnsemblingJob
	err := query.Find(&ensemblingJobs).Error
	return ensemblingJobs, err
}

func (s *ensemblingJobService) UpdateJobStatus(id models.ID, status models.State, errString string) error {
	updateMap := map[string]interface{}{
		"status": status,
	}
	if errString != "" {
		updateMap["error"] = errString
	}
	return s.db.Model(&models.EnsemblingJob{}).Where("id = ?", id).Updates(updateMap).Error
}
