package service

import "github.com/caraml-dev/turing/api/turing/models"

type RoutersRepository interface {
	// CountRoutersByCurrentVersionID returns the number of routers with the current version set to the given version.
	CountRoutersByCurrentVersionID(routerVersionID models.ID) int64
}
