package service

import (
	"math"

	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
)

// EnsemblersService is the data access object for the Ensemblers from the db.
type EnsemblersService interface {
	// FindByID Find an ensembler matching the given id
	FindByID(id models.ID) (models.EnsemblerLike, error)
	// List tbu
	List(projectID models.ID, query ListEnsemblersQuery) (*PaginatedResults, error)
	// Save the given router to the db. Updates the existing record if already exists
	Save(ensembler models.EnsemblerLike) (models.EnsemblerLike, error)
}

type ListEnsemblersQuery struct {
	paginationQuery
}

// NewEnsemblersService creates a new ensemblers service
func NewEnsemblersService(db *gorm.DB) EnsemblersService {
	return &ensemblersService{db: db.Debug()}
}

type ensemblersService struct {
	db *gorm.DB
}

func (service *ensemblersService) FindByID(id models.ID) (models.EnsemblerLike, error) {
	var ensembler models.GenericEnsembler
	query := service.db.Where("id = ?", id)

	result := query.First(&ensembler)
	if err := result.Error; err != nil {
		return nil, err
	}
	instance := ensembler.Instance()
	result = query.Scopes(models.EnsemblerTable(&ensembler)).First(instance)
	if err := result.Error; err != nil {
		return nil, err
	}
	return instance, nil
}

func (service *ensemblersService) List(
	projectID models.ID,
	query ListEnsemblersQuery,
) (*PaginatedResults, error) {
	var results []*models.GenericEnsembler
	var count int
	done := make(chan bool, 1)

	db := service.db.Where("project_id = ?", projectID)
	go func() {
		db.Model(&results).Count(&count)
		done <- true
	}()

	result := db.Scopes(Paginate(query)).Find(&results)
	<-done

	if err := result.Error; err != nil {
		return nil, err
	}
	return &PaginatedResults{
		Results: results,
		Paging: Paging{
			Total: count,
			Page:  query.Page(),
			Pages: int(math.Ceil(float64(count) / float64(query.PageSize()))),
		},
	}, nil
}

func (service *ensemblersService) Save(ensembler models.EnsemblerLike) (models.EnsemblerLike, error) {
	result := service.db.
		Scopes(models.EnsemblerTable(ensembler)).
		Save(ensembler)

	if err := result.Error; err != nil {
		return nil, err
	}
	return ensembler, nil
}
