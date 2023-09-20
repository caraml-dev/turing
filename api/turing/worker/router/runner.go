package router

import (
	"time"

	"github.com/caraml-dev/turing/api/turing/api"
	"github.com/caraml-dev/turing/api/turing/worker"
)

var pageOne = 1

type routerJobRunner struct {
	deploymentController           api.RouterDeploymentController
	recordsToProcessInOneIteration int
	maxRetryCount                  int
	timeInterval                   time.Duration
}

// NewRouterJobRunner creates a new router job runner
// This service controls the orchestration of router version.

func NewRouterJobRunner(deploymentController api.RouterDeploymentController) worker.JobRunner {
	return &routerJobRunner{deploymentController: deploymentController}
}

func (r *routerJobRunner) GetInterval() time.Duration {
	return r.timeInterval
}

func (r *routerJobRunner) Run() {
	r.processJobs()
	r.updateStatus()
}

func (r *routerJobRunner) updateStatus() {
}

func (r *routerJobRunner) processJobs() {
}
