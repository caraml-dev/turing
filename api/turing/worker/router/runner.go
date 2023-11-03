package router

import (
	"time"

	mlp "github.com/caraml-dev/mlp/api/client"
	"github.com/caraml-dev/turing/api/turing/log"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/api/turing/worker"
)

const deploymentTimeoutBuffer = time.Second * 5

type routerJobRunner struct {
	deploymentController DeploymentController
	maxRetryCount        int
	timeInterval         time.Duration
}

// NewRouterJobRunner creates a new router job runner
// This service controls the orchestration of router version.
func NewRouterJobRunner(deploymentController DeploymentController) worker.JobRunner {
	return &routerJobRunner{deploymentController: deploymentController}
}

func (r *routerJobRunner) GetInterval() time.Duration {
	return r.timeInterval
}

func (r *routerJobRunner) Run() {
	r.processRouterVersions()
}

func (r *routerJobRunner) processRouterVersions() {
	options := service.RouterVersionListOptions{
		Statuses: []models.RouterVersionStatus{models.RouterVersionStatusPending},
	}
	routerVersions, err := r.deploymentController.RouterVersionsService.ListRouterVersionsWithFilter(options)
	if err != nil {
		log.Errorf("unable to get router versions with pending status: %v", err)
		return
	}

	if len(routerVersions) == 0 {
		return
	}

	deploymentTimeout := r.deploymentController.DeploymentService.GetDeploymentTimeout() + deploymentTimeoutBuffer
	for _, routerVersion := range routerVersions {
		// TODO: Acquire row lock and re-get the router version before comparison.
		if time.Since(routerVersion.DeploymentStartTime) > deploymentTimeout {
			router, err := r.deploymentController.RoutersService.FindByID(routerVersion.ID)
			if err != nil {
				log.Errorf("unable to get router id for router with pending status: %v", err)
				continue
			}
			mlpProject, err := r.deploymentController.MLPService.GetProject(router.ProjectID)
			if err != nil {
				log.Errorf("unable to get mlp project for router with pending status: %v", err)
				continue
			}
			go func(project *mlp.Project, router *models.Router, routerVersion *models.RouterVersion,
			) {
				err := r.deploymentController.DeployOrRollbackRouter(project, router, routerVersion)
				if err != nil {
					log.Errorf("Error deploying router version %s:%s:%d: %v",
						project.Name, router.Name, routerVersion.Version, err)
				}
			}(mlpProject, router, routerVersion)
		}
	}
}
