package service

import (
	"errors"
	"fmt"

	"github.com/caraml-dev/turing/api/turing/models"
)

// RouterVersionsService must implement the business logic for router version related operations.
type RouterVersionsService interface {
	// ListRouterVersions list all RouterVersions associated with the given routerID
	ListRouterVersions(routerID models.ID) ([]*models.RouterVersion, error)
	// ListRouterVersionsWithStatus lists the RouterVersions for the given Router matching the given status.
	ListRouterVersionsWithStatus(routerID models.ID, status models.RouterVersionStatus) ([]*models.RouterVersion, error)
	// FindByID finds the RouterVersion matching the given id.
	FindByID(routerVersionID models.ID) (*models.RouterVersion, error)
	// FindByRouterIDAndVersion finds the RouterVersion for the given Router matching the given version.
	FindByRouterIDAndVersion(routerID models.ID, version uint) (*models.RouterVersion, error)
	// FindLatestVersionByRouterID finds the latest RouterVersion for the given Router matching the given version.
	FindLatestVersionByRouterID(routerID models.ID) (*models.RouterVersion, error)
	// CreateRouterVersion creates a new router version
	CreateRouterVersion(routerVersion *models.RouterVersion) (*models.RouterVersion, error)
	// UpdateRouterVersion updates an existing router version
	UpdateRouterVersion(routerVersion *models.RouterVersion) (*models.RouterVersion, error)
	// Delete deletes the given RouterVersion from the db. This method deletes all child objects (enricher, ensembler).
	Delete(routerVersion *models.RouterVersion) error
}

func NewRouterVersionsService(
	r RoutersRepository,
	rv RouterVersionsRepository,
	services *Services,
) RouterVersionsService {
	return &routerVersionsService{
		routersRepository:        r,
		routerVersionsRepository: rv,
		services:                 services,
	}
}

type routerVersionsService struct {
	routersRepository        RoutersRepository
	routerVersionsRepository RouterVersionsRepository
	services                 *Services
}

func (svc *routerVersionsService) ListRouterVersions(routerID models.ID) ([]*models.RouterVersion, error) {
	routerVersions, err := svc.routerVersionsRepository.List(routerID)
	if err != nil {
		return nil, err
	}

	// Generate monitoring URL
	for _, routerVersion := range routerVersions {
		err := svc.setMonitoringURL(routerVersion)
		if err != nil {
			return nil, err
		}
	}

	return routerVersions, nil
}

func (svc *routerVersionsService) ListRouterVersionsWithStatus(
	routerID models.ID,
	status models.RouterVersionStatus,
) ([]*models.RouterVersion, error) {
	routerVersions, err := svc.routerVersionsRepository.ListByStatus(routerID, status)
	if err != nil {
		return nil, err
	}

	// Generate monitoring URL
	for _, routerVersion := range routerVersions {
		err := svc.setMonitoringURL(routerVersion)
		if err != nil {
			return nil, err
		}
	}

	return routerVersions, nil
}

func (svc *routerVersionsService) FindByID(
	routerVersionID models.ID,
) (*models.RouterVersion, error) {
	routerVersion, err := svc.routerVersionsRepository.FindByID(routerVersionID)
	if err != nil {
		return nil, err
	}

	// Generate monitoring URL
	err = svc.setMonitoringURL(routerVersion)
	if err != nil {
		return nil, err
	}

	return routerVersion, nil
}

func (svc *routerVersionsService) FindByRouterIDAndVersion(
	routerID models.ID,
	version uint,
) (*models.RouterVersion, error) {
	routerVersion, err := svc.routerVersionsRepository.FindByRouterIDAndVersion(routerID, version)
	if err != nil {
		return nil, err
	}

	// Generate monitoring URL
	err = svc.setMonitoringURL(routerVersion)
	if err != nil {
		return nil, err
	}

	return routerVersion, nil
}

func (svc *routerVersionsService) FindLatestVersionByRouterID(
	routerID models.ID,
) (*models.RouterVersion, error) {
	routerVersion, err := svc.routerVersionsRepository.FindLatestVersion(routerID)
	if err != nil {
		return nil, err
	}

	// Generate monitoring URL
	err = svc.setMonitoringURL(routerVersion)
	if err != nil {
		return nil, err
	}

	return routerVersion, nil
}

func (svc *routerVersionsService) CreateRouterVersion(
	routerVersion *models.RouterVersion,
) (*models.RouterVersion, error) {
	routerVersion, err := svc.routerVersionsRepository.Save(routerVersion)
	if err != nil {
		return nil, err
	}

	// Generate monitoring URL
	err = svc.setMonitoringURL(routerVersion)
	if err != nil {
		return nil, err
	}

	return routerVersion, nil
}

func (svc *routerVersionsService) UpdateRouterVersion(
	routerVersion *models.RouterVersion,
) (*models.RouterVersion, error) {
	routerVersion, err := svc.routerVersionsRepository.Save(routerVersion)
	if err != nil {
		return nil, err
	}

	// Generate monitoring URL
	err = svc.setMonitoringURL(routerVersion)
	if err != nil {
		return nil, err
	}

	return routerVersion, nil
}

func (svc *routerVersionsService) Delete(routerVersion *models.RouterVersion) error {
	if routerVersion.ID == 0 {
		return errors.New("router version must have valid primary key to be deleted")
	}

	activeRouters := svc.routersRepository.CountRoutersByCurrentVersionID(routerVersion.ID)
	if activeRouters > 0 {
		return errors.New("unable to delete router version - there exists a router that is currently using this version")
	}

	return svc.routerVersionsRepository.Delete(routerVersion)
}

func (svc *routerVersionsService) setMonitoringURL(routerVersion *models.RouterVersion) error {
	var err error
	routerVersion.MonitoringURL, err = svc.services.RouterMonitoringService.GenerateMonitoringURL(
		routerVersion.Router.ProjectID,
		routerVersion.Router.EnvironmentName,
		routerVersion.Router.Name,
		&routerVersion.Version,
	)
	if err != nil {
		return fmt.Errorf("unable to generate MonitoringURL for router version: %s", err.Error())
	}
	return nil
}
