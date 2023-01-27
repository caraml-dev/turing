package service

import (
	"errors"
	"fmt"

	"github.com/caraml-dev/turing/api/turing/log"
	"github.com/caraml-dev/turing/api/turing/models"
	mlp "github.com/gojek/mlp/api/client"
)

// RouterVersionsService must implement the business logic for router version related operations.
type RouterVersionsService interface {
	// ListByRouterID list all RouterVersions associated with the given routerID
	ListByRouterID(routerID models.ID) ([]*models.RouterVersion, error)
	// ListByRouterIDAndStatus lists the RouterVersions for the given Router matching the given status.
	ListByRouterIDAndStatus(routerID models.ID, status models.RouterVersionStatus) ([]*models.RouterVersion, error)
	// FindByID finds the RouterVersion matching the given id.
	FindByID(routerVersionID models.ID) (*models.RouterVersion, error)
	// FindByRouterIDAndVersion finds the RouterVersion for the given Router matching the given version.
	FindByRouterIDAndVersion(routerID models.ID, version uint) (*models.RouterVersion, error)
	// FindLatestVersionByRouterID finds the latest RouterVersion for the given Router matching the given version.
	FindLatestVersionByRouterID(routerID models.ID) (*models.RouterVersion, error)
	// Create creates a new router version
	Create(routerVersion *models.RouterVersion) (*models.RouterVersion, error)
	// Update updates an existing router version
	Update(routerVersion *models.RouterVersion) (*models.RouterVersion, error)
	// Delete deletes the given RouterVersion from the db. This method deletes all child objects (enricher, ensembler).
	Delete(routerVersion *models.RouterVersion) error
	// Deploy deploys the given router version
	Deploy(project *mlp.Project, router *models.Router, routerVersion *models.RouterVersion) error
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

func (svc *routerVersionsService) ListByRouterID(routerID models.ID) ([]*models.RouterVersion, error) {
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

func (svc *routerVersionsService) ListByRouterIDAndStatus(
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

func (svc *routerVersionsService) Create(
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

func (svc *routerVersionsService) Update(
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

func (svc *routerVersionsService) Delete(
	routerVersion *models.RouterVersion,
) error {
	// Check router version's status
	if routerVersion.Status == models.RouterVersionStatusPending {
		return errors.New("unable to delete router version that is currently deploying")
	}

	// Check that the version is not linked to any router as the current version
	// (Usually, checking routerVersion.Router.CurrentRouterVersionID != routerVersion.ID should suffice;
	// but this check is more comprehensive.)
	activeRouters := svc.routersRepository.CountRoutersByCurrentVersionID(routerVersion.ID)
	if activeRouters > 0 {
		return errors.New("unable to delete router version - there exists a router that is currently using this version")
	}

	return svc.routerVersionsRepository.Delete(routerVersion)
}

func (svc *routerVersionsService) Deploy(
	project *mlp.Project,
	router *models.Router,
	routerVersion *models.RouterVersion,
) error {
	if router.Status == models.RouterStatusPending {
		return errors.New("router is currently deploying, cannot do another deployment")
	}
	if routerVersion.Status == models.RouterVersionStatusDeployed {
		return errors.New("router version is already deployed")
	}

	// Deploy the version asynchronously
	go func() {
		err := svc.services.RouterDeploymentService.DeployOrRollbackRouter(project, router, routerVersion)
		if err != nil {
			log.Errorf("Error deploying router version %s:%s:%d: %v",
				project.Name, router.Name, routerVersion.Version, err)
		}
	}()
	return nil
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
