package service

import "github.com/caraml-dev/turing/api/turing/models"

type RouterVersionsRepository interface {
	// List lists all RouterVersions associated with the given routerID
	List(routerID models.ID) ([]*models.RouterVersion, error)
	// ListByStatus lists the RouterVersions for the given Router matching the given status.
	ListByStatus(routerID models.ID, status models.RouterVersionStatus) ([]*models.RouterVersion, error)
	// Save saves the given RouterVersion to the db. Updates the existing record if already exists.
	Save(routerVersion *models.RouterVersion) (*models.RouterVersion, error)
	// FindByID finds the RouterVersion matching the given id.
	FindByID(routerVersionID models.ID) (*models.RouterVersion, error)
	// FindByRouterIDAndVersion finds the RouterVersion for the given Router matching the given version.
	FindByRouterIDAndVersion(routerID models.ID, version uint) (*models.RouterVersion, error)
	// FindLatestVersion finds the latest RouterVersion for the given Router.
	FindLatestVersion(routerID models.ID) (*models.RouterVersion, error)
	// Delete deletes the given RouterVersion from the db. This method deletes all child objects (enricher, ensembler).
	Delete(routerVersion *models.RouterVersion) error
}
