package service

import (
	"errors"

	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
)

// RoutersService is the data access object for the Routers from the db.
type RoutersService interface {
	//ListRouters List routers within the given project and environment.
	ListRouters(projectID models.ID, environmentName string) ([]*models.Router, error)
	// Save the given router to the db. Updates the existing record if already exists.
	Save(router *models.Router) (*models.Router, error)
	//FindByID Find a router matching the given router id.
	FindByID(routerID models.ID) (*models.Router, error)
	//FindByProjectAndName Find a router within the given project that matches the given name.
	FindByProjectAndName(projectID models.ID, routerName string) (*models.Router, error)
	//FindByProjectAndEnvironmentAndName Find a router within the given project and environment
	// that matches the given name.
	FindByProjectAndEnvironmentAndName(
		projectID models.ID,
		environmentName string,
		routerName string,
	) (*models.Router, error)
	// Delete a router. This deletes all child objects of the router (router versions, ensemblers and enrichers)
	// (Transactional).
	Delete(router *models.Router) error
}

// NewRoutersService creates a new router service
func NewRoutersService(db *gorm.DB) RoutersService {
	return &routersService{db: db}
}

type routersService struct {
	db *gorm.DB
}

func (service *routersService) query() *gorm.DB {
	return service.db.
		Preload("CurrRouterVersion").
		Preload("CurrRouterVersion.Enricher").
		Preload("CurrRouterVersion.Ensembler")
}

func (service *routersService) ListRouters(projectID models.ID, environmentName string) ([]*models.Router, error) {
	var routers []*models.Router
	query := service.query()
	if projectID > 0 {
		query = query.Where("routers.project_id = ?", projectID)
	}
	if environmentName != "" {
		query = query.Where("routers.environment_name = ?", environmentName)
	}
	err := query.Find(&routers).Error
	return routers, err
}

func (service *routersService) Save(router *models.Router) (*models.Router, error) {
	var err error
	if service.db.NewRecord(router) {
		err = service.db.Create(router).Error
	} else {
		err = service.db.Save(router).Error
	}
	if err != nil {
		return nil, err
	}
	return service.FindByID(router.ID)
}

func (service *routersService) FindByID(routerID models.ID) (*models.Router, error) {
	var router models.Router
	query := service.query().
		Where("routers.id = ?", routerID).
		First(&router)
	if err := query.Error; err != nil {
		return nil, err
	}
	return &router, nil
}

func (service *routersService) FindByProjectAndEnvironmentAndName(
	projectID models.ID,
	environmentName string,
	name string,
) (*models.Router, error) {
	var router models.Router
	query := service.query().
		Where("routers.project_id = ?", projectID).
		Where("routers.environment_name = ?", environmentName).
		Where("routers.name = ?", name).
		First(&router)
	if err := query.Error; err != nil {
		return nil, err
	}
	return &router, nil
}

func (service *routersService) FindByProjectAndName(projectID models.ID, name string) (*models.Router, error) {
	var router models.Router
	query := service.query().
		Where("routers.project_id = ?", projectID).
		Where("routers.name = ?", name).
		First(&router)
	if err := query.Error; err != nil {
		return nil, err
	}
	return &router, nil
}

func (service *routersService) Delete(router *models.Router) error {
	if router.ID == 0 {
		return errors.New("router must have valid primary key to be deleted")
	}
	tx := service.db.Begin()

	// remove associations
	router.ClearCurrRouterVersion()
	tx.Save(router)

	var routerVersions []*models.RouterVersion
	if err := service.db.
		Where("router_id = ?", router.ID).
		Find(&routerVersions).Error; err != nil {
		return err
	}

	for _, routerVersion := range routerVersions {
		tx.Delete(routerVersion)
		if routerVersion.EnricherID.Valid {
			tx.Delete(models.Enricher{}, "id = ?", routerVersion.EnricherID.Int32)
		}
		if routerVersion.EnsemblerID.Valid {
			tx.Delete(models.Ensembler{}, "id = ?", routerVersion.EnsemblerID.Int32)
		}
	}
	tx.Delete(router)

	return tx.Commit().Error
}
