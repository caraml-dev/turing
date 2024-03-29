package service

import (
	"errors"
	"fmt"
	"text/template"

	"gorm.io/gorm"

	logger "github.com/caraml-dev/turing/api/turing/log"
	"github.com/caraml-dev/turing/api/turing/models"
)

const (
	grafanaAllVariable = "$__all"
)

type RouterVersionListOptions struct {
	ProjectID   *models.ID                   `schema:"project_id" validate:"required"`
	EnsemblerID *models.ID                   `schema:"ensembler_id"`
	Statuses    []models.RouterVersionStatus `schema:"status"`
	IsCurrent   bool                         `schema:"is_current"`
}

// RouterVersionsService is the data access object for RouterVersions from the db.
type RouterVersionsService interface {
	// ListRouterVersions List all RouterVersions associated with the given routerID
	ListRouterVersions(routerID models.ID) ([]*models.RouterVersion, error)
	// ListRouterVersionsWithStatus Lists the RouterVersions for the given Router matching the given status.
	ListRouterVersionsWithStatus(routerID models.ID, status models.RouterVersionStatus) ([]*models.RouterVersion, error)
	// Save the given RouterVersion to the db. Updates the existing record if already exists.
	Save(routerVersion *models.RouterVersion) (*models.RouterVersion, error)
	// FindByID Finds the RouterVersion matching the given id.
	FindByID(routerVersionID models.ID) (*models.RouterVersion, error)
	// FindByRouterIDAndVersion Finds the RouterVersion for the given Router matching the given version.
	FindByRouterIDAndVersion(routerID models.ID, version uint) (*models.RouterVersion, error)
	// FindLatestVersionByRouterID Finds the latest RouterVersion for the given Router matching the given version.
	FindLatestVersionByRouterID(routerID models.ID) (*models.RouterVersion, error)
	// Delete Deletes the given RouterVersion from the db. This method deletes all child objects (enricher, ensembler).
	Delete(routerVersion *models.RouterVersion) error
	// ListRouterVersionsWithFilter Lists router version with parameter option
	// the filter contain status, ensembler_id and current router version
	ListRouterVersionsWithFilter(options RouterVersionListOptions) ([]*models.RouterVersion, error)
}

func NewRouterVersionsService(
	db *gorm.DB,
	mlpService MLPService,
	monitoringURLFormat *string,
) RouterVersionsService {
	var monitoringURLTemplate *template.Template
	if monitoringURLFormat != nil {
		var err error
		monitoringURLTemplate, err = template.New("monitoringURLTemplate").Parse(*monitoringURLFormat)
		if err != nil {
			logger.Warnf("error parsing monitoring url template: %s", err)
		}
	}

	return &routerVersionsService{
		db:                      db,
		RouterMonitoringService: NewRouterMonitoringService(mlpService, monitoringURLTemplate),
	}
}

type routerVersionsService struct {
	RouterMonitoringService
	db *gorm.DB
}

func (service *routerVersionsService) query() *gorm.DB {
	return service.db.
		Preload("Router").
		Preload("Enricher").
		Preload("Ensembler")
}

func (service *routerVersionsService) ListRouterVersions(routerID models.ID) ([]*models.RouterVersion, error) {
	var routerVersions []*models.RouterVersion
	query := service.query().Where("router_id = ?", routerID).Find(&routerVersions)

	for _, routerVersion := range routerVersions {
		var err error
		routerVersion.MonitoringURL, err = service.GenerateMonitoringURL(
			routerVersion.Router.ProjectID,
			routerVersion.Router.EnvironmentName,
			routerVersion.Router.Name,
			&routerVersion.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to generate MonitoringURL for router version: %s", err.Error())
		}
	}

	return routerVersions, query.Error
}

func (service *routerVersionsService) ListRouterVersionsWithStatus(
	routerID models.ID,
	status models.RouterVersionStatus,
) ([]*models.RouterVersion, error) {
	var routerVersions []*models.RouterVersion
	query := service.query().
		Where("router_id = ?", routerID).
		Where("status = ?", status).
		Find(&routerVersions)

	for _, routerVersion := range routerVersions {
		var err error
		routerVersion.MonitoringURL, err = service.GenerateMonitoringURL(
			routerVersion.Router.ProjectID,
			routerVersion.Router.EnvironmentName,
			routerVersion.Router.Name,
			&routerVersion.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to generate MonitoringURL for router version: %s", err.Error())
		}
	}

	return routerVersions, query.Error
}

func (service *routerVersionsService) Save(routerVersion *models.RouterVersion) (*models.RouterVersion, error) {
	var err error
	tx := service.db.Begin()

	if routerVersion.Model.IsNew() {
		if routerVersion.Ensembler != nil {
			if err := tx.Create(routerVersion.Ensembler).Error; err != nil {
				return nil, err
			}
			routerVersion.SetEnsemblerID(routerVersion.Ensembler.ID)
		}
		if routerVersion.Enricher != nil {
			if err := tx.Create(routerVersion.Enricher).Error; err != nil {
				return nil, err
			}
			routerVersion.SetEnricherID(routerVersion.Enricher.ID)
		}
		err = tx.Create(routerVersion).Error
	} else {
		// We don't allow ensembler and enricher updates without an update to the entire router version.
		// Changes to those elements will always result in a new version being spawned.
		err = tx.Save(routerVersion).Error
		if err != nil {
			return nil, err
		}
		// These steps below are meant to ensure the ensembler and enricher config tables are up-to-date with the
		// router version
		if routerVersion.Enricher != nil {
			err = tx.Save(routerVersion.Enricher).Error
			if err != nil {
				return nil, err
			}
		}
		if routerVersion.Ensembler != nil {
			err = tx.Save(routerVersion.Ensembler).Error
			if err != nil {
				return nil, err
			}
		}
	}

	if err != nil {
		return nil, err
	}
	tx.Commit()
	return service.FindByID(routerVersion.ID)
}

func (service *routerVersionsService) FindByID(
	routerVersionID models.ID,
) (*models.RouterVersion, error) {
	var routerVersion models.RouterVersion
	query := service.query().
		Where("router_versions.id = ?", routerVersionID).
		First(&routerVersion)
	if err := query.Error; err != nil {
		return nil, err
	}

	var err error
	routerVersion.MonitoringURL, err = service.GenerateMonitoringURL(
		routerVersion.Router.ProjectID,
		routerVersion.Router.EnvironmentName,
		routerVersion.Router.Name,
		&routerVersion.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to generate MonitoringURL for router version: %s", err.Error())
	}

	return &routerVersion, nil
}

func (service *routerVersionsService) FindByRouterIDAndVersion(
	routerID models.ID,
	version uint,
) (*models.RouterVersion, error) {
	var routerVersion models.RouterVersion
	query := service.query().
		Where("router_id = ?", routerID).
		Where("version = ?", version).
		First(&routerVersion)
	if err := query.Error; err != nil {
		return nil, err
	}

	var err error
	routerVersion.MonitoringURL, err = service.GenerateMonitoringURL(
		routerVersion.Router.ProjectID,
		routerVersion.Router.EnvironmentName,
		routerVersion.Router.Name,
		&routerVersion.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to generate MonitoringURL for router version: %s", err.Error())
	}

	return &routerVersion, nil
}

func (service *routerVersionsService) FindLatestVersionByRouterID(
	routerID models.ID,
) (*models.RouterVersion, error) {
	var routerVersion models.RouterVersion
	query := service.query().
		Where("router_id = ?", routerID).
		Order("version desc").
		First(&routerVersion)
	if err := query.Error; err != nil {
		return nil, err
	}

	var err error
	routerVersion.MonitoringURL, err = service.GenerateMonitoringURL(
		routerVersion.Router.ProjectID,
		routerVersion.Router.EnvironmentName,
		routerVersion.Router.Name,
		&routerVersion.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to generate MonitoringURL for router version: %s", err.Error())
	}

	return &routerVersion, nil
}

func (service *routerVersionsService) Delete(routerVersion *models.RouterVersion) error {
	if routerVersion.ID == 0 {
		return errors.New("router version must have valid primary key to be deleted")
	}

	var upstreamRefs int64
	service.db.Table("routers").
		Where("routers.curr_router_version_id = ?", routerVersion.ID).
		Count(&upstreamRefs)
	if upstreamRefs != 0 {
		return errors.New("unable to delete router version - there exists a router that is currently using this version")
	}

	tx := service.db.Begin()
	tx.Delete(routerVersion)
	if routerVersion.Enricher != nil {
		tx.Delete(routerVersion.Enricher)
	}
	if routerVersion.Ensembler != nil {
		tx.Delete(routerVersion.Ensembler)
	}
	return tx.Commit().Error
}

func (service *routerVersionsService) ListRouterVersionsWithFilter(
	option RouterVersionListOptions,
) ([]*models.RouterVersion, error) {
	var routerVersions []*models.RouterVersion
	query := service.query()

	if option.EnsemblerID != nil {
		query = query.Where("ensembler_id IN (?)",
			service.db.Table("ensembler_configs").Select("id").
				Where("CAST(pyfunc_config->>'ensembler_id' AS INTEGER) = ?", option.EnsemblerID))
	}
	if option.Statuses != nil {
		query = query.Where("status IN (?)", option.Statuses)
	}

	if option.IsCurrent {
		query = query.Where("id IN (?)", service.db.Table("routers").Select("curr_router_version_id"))
	}

	query = query.Find(&routerVersions)

	if err := query.Error; err != nil {
		return nil, err
	}

	return routerVersions, nil
}
