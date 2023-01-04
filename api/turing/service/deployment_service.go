package service

// import (
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"strings"

// 	merlin "github.com/gojek/merlin/client"
// 	mlp "github.com/gojek/mlp/api/client"

// 	"github.com/caraml-dev/turing/api/turing/models"
// 	"github.com/caraml-dev/turing/engines/experiment/manager"
// )

// // RouterDeploymentService handles the deployment of routers
// type RouterDeploymentService interface {
// 	DeployOrRollbackRouter(project *mlp.Project, router *models.Router, routerVersion *models.RouterVersion) error
// 	UndeployRouter(project *mlp.Project, router *models.Router) error
// }

// type routerDeploymentService struct {
// 	services *Services
// }

// func NewRouterDeploymentService(services *Services) RouterDeploymentService {
// 	return &routerDeploymentService{
// 		services: services,
// 	}
// }
