package service

import (
	"bytes"
	"errors"
	"fmt"
	"text/template"

	logger "github.com/gojek/turing/api/turing/log"
	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
)

const (
	grafanaAllVariable = "$__all"
)

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
	// GenerateMonitoringURL generates the monitoring url based on the router version.
	GenerateMonitoringURL(
		projectName string,
		environmentName string,
		routerName string,
		routerVersion *uint,
	) (string, error)
}

func NewRouterVersionsService(
	db *gorm.DB,
	mlpService MLPService,
	monitoringURLFormat *string,
) RouterVersionsService {
	var monitoringURLTemplate *template.Template
	var err error
	if monitoringURLFormat != nil {
		monitoringURLTemplate, err = template.New("monitoringURLTemplate").Parse(*monitoringURLFormat)
		if err != nil {
			logger.Warnf("error parsing monitoring url template: %s", err)
		}
	}

	return &routerVersionsService{
		db:                    db,
		mlpService:            mlpService,
		monitoringURLTemplate: monitoringURLTemplate,
	}
}

type routerVersionsService struct {
	db                    *gorm.DB
	mlpService            MLPService
	monitoringURLTemplate *template.Template
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
	return routerVersions, query.Error
}

func (service *routerVersionsService) Save(routerVersion *models.RouterVersion) (*models.RouterVersion, error) {
	var err error
	tx := service.db.Begin()
	if service.db.NewRecord(routerVersion) {
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
		// We don't allow ensembler and enricher updates. Changes to those elements
		// will always result in a new version being spawned.
		err = tx.Save(routerVersion).Error
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
	return &routerVersion, nil
}

func (service *routerVersionsService) Delete(routerVersion *models.RouterVersion) error {
	if routerVersion.ID == 0 {
		return errors.New("router version must have valid primary key to be deleted")
	}

	var upstreamRefs int
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

type monitoringURLValues struct {
	ClusterName string
	ProjectName string
	RouterName  string
	Version     string
}

func (service *routerVersionsService) GenerateMonitoringURL(
	projectName string,
	environmentName string,
	routerName string,
	routerVersion *uint,
) (string, error) {
	if service.monitoringURLTemplate == nil {
		return "", nil
	}

	env, err := service.mlpService.GetEnvironment(environmentName)
	if err != nil {
		return "", err
	}

	var versionString string
	if routerVersion == nil {
		versionString = grafanaAllVariable
	} else {
		versionString = fmt.Sprintf("%d", *routerVersion)
	}

	values := monitoringURLValues{
		ClusterName: env.Cluster,
		ProjectName: projectName,
		RouterName:  routerName,
		Version:     versionString,
	}

	var b bytes.Buffer
	err = service.monitoringURLTemplate.Execute(&b, values)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
