package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	mlp "github.com/caraml-dev/mlp/api/client"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
)

type EnsemblersController struct {
	BaseController
}

func (c EnsemblersController) ListEnsemblers(
	_ *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	options := service.EnsemblersListOptions{}

	if err := c.ParseVars(&options, vars); err != nil {
		return BadRequest("unable to list ensemblers",
			fmt.Sprintf("failed to parse query string: %s", err))
	}

	results, err := c.EnsemblersService.List(options)
	if err != nil {
		return InternalServerError("unable to list ensemblers", err.Error())
	}

	return Ok(results)
}

func (c EnsemblersController) GetEnsembler(
	_ *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	options := EnsemblersPathOptions{}

	if err := c.ParseVars(&options, vars); err != nil {
		return BadRequest("failed to fetch ensembler",
			fmt.Sprintf("failed to parse query string: %s", err))
	}

	ensembler, err := c.EnsemblersService.FindByID(
		*options.EnsemblerID,
		service.EnsemblersFindByIDOptions{
			ProjectID: options.ProjectID,
		})
	if err != nil {
		return NotFound("ensembler not found", err.Error())
	}

	return Ok(ensembler)
}

func (c EnsemblersController) CreateEnsembler(
	_ *http.Request,
	vars RequestVars,
	body interface{},
) *Response {
	var errResp *Response
	var project *mlp.Project
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}
	var err error
	ensembler := body.(*request.CreateOrUpdateEnsemblerRequest).EnsemblerLike
	ensembler.SetProjectID(models.ID(project.ID))

	ensembler, err = c.EnsemblersService.Save(ensembler)
	if err != nil {
		return InternalServerError("unable to save the ensembler", err.Error())
	}

	return Created(ensembler)
}

func (c EnsemblersController) UpdateEnsembler(
	_ *http.Request,
	vars RequestVars,
	body interface{},
) *Response {
	options := EnsemblersPathOptions{}

	if err := c.ParseVars(&options, vars); err != nil {
		return BadRequest("failed to fetch ensembler",
			fmt.Sprintf("failed to parse query string: %s", err))
	}

	ensembler, err := c.EnsemblersService.FindByID(
		*options.EnsemblerID,
		service.EnsemblersFindByIDOptions{
			ProjectID: options.ProjectID,
		})
	if err != nil {
		return NotFound("ensembler not found", err.Error())
	}

	request := body.(*request.CreateOrUpdateEnsemblerRequest)

	if ensembler.GetType() != request.GetType() {
		return BadRequest("invalid ensembler configuration",
			"Ensembler type cannot be changed after creation")
	}

	pyFuncEnsembler, isPyfunc := ensembler.(*models.PyFuncEnsembler)

	var oldPyFuncEnsembler models.PyFuncEnsembler
	if isPyfunc {
		oldPyFuncEnsembler = models.PyFuncEnsembler{
			RunID:       pyFuncEnsembler.RunID,
			ArtifactURI: pyFuncEnsembler.ArtifactURI,
		}
	}

	if err = ensembler.Patch(request.EnsemblerLike); err != nil {
		return BadRequest("invalid ensembler configuration", err.Error())
	}
	ensembler, err = c.EnsemblersService.Save(ensembler)
	if err != nil {
		// Delete If Only RunID is changed
		// If only the ensembler name that changed, mlflow won't create a new run, so we don't need to delete old / new run
		// Since the sdk already create a new mlflow run, when the update failed, we need to clean up the new mlflow run
		if isPyfunc && oldPyFuncEnsembler.RunID != pyFuncEnsembler.RunID {
			pyFuncErr := c.MlflowService.DeleteRun(context.Background(),
				pyFuncEnsembler.RunID, pyFuncEnsembler.ArtifactURI, true)
			if pyFuncErr != nil {
				return InternalServerError("failed to update the ensembler", "cleanup process failed")
			}
		}
		return InternalServerError("failed to update the ensembler", err.Error())
	}

	// Delete If Only RunID is changed
	// If only the ensembler name that changed, mlflow won't create a new run, so we don't need to delete old / new run
	// The update is success, and now we need to cleanup old ensembler
	if isPyfunc && oldPyFuncEnsembler.RunID != pyFuncEnsembler.RunID {
		err = c.MlflowService.DeleteRun(context.Background(), oldPyFuncEnsembler.RunID, oldPyFuncEnsembler.ArtifactURI, true)
		if err != nil {
			fmt.Printf("failed to delete mlflow old run: %s", err.Error())
		}
	}

	return Ok(ensembler)
}

func (c EnsemblersController) DeleteEnsembler(
	_ *http.Request,
	vars RequestVars,
	_ interface{},
) *Response {
	options := EnsemblersPathOptions{}

	if err := c.ParseVars(&options, vars); err != nil {
		return BadRequest("failed to fetch ensembler",
			fmt.Sprintf("failed to parse query string: %s", err))
	}

	ensembler, err := c.EnsemblersService.FindByID(
		*options.EnsemblerID,
		service.EnsemblersFindByIDOptions{
			ProjectID: options.ProjectID,
		})

	if err != nil {
		return NotFound("ensembler not found", err.Error())
	}

	// First we need to check if there are active job / router using the ensembler
	// If such ensembler exist, the deletion process are restricted
	// Check if there are any deployed / pending router version using the ensembler
	httpStatus, err := c.checkActiveRouterVersion(options)
	if err != nil {
		return Error(httpStatus, "failed to delete the ensembler", err.Error())
	}

	// Check if there are any current router version using the ensembler
	httpStatus, err = c.checkCurrentRouterVersion(options)
	if err != nil {
		return Error(httpStatus, "failed to delete the ensembler", err.Error())
	}

	// Check if there are any building / pending / running ensembling jobs using the ensembler
	httpStatus, err = c.checkActiveEnsemblingJob(options)
	if err != nil {
		return Error(httpStatus, "failed to delete the ensembler", err.Error())
	}

	// If no active job / router is using the ensembler, we can proceed to delete it.
	// However, the inactive router versions and jobs using the ensembler will have to be deleted.
	// Deleting inactive router versions
	httpStatus, err = c.deleteInactiveRouterVersion(options)
	if err != nil {
		return Error(httpStatus, "unable to delete router version", err.Error())
	}

	// Deleting inactive ensembling jobs
	httpStatus, err = c.deleteInactiveEnsemblingJob(options)
	if err != nil {
		return Error(httpStatus, "unable to delete ensembling job", err.Error())
	}

	httpStatus, err = c.deleteEnsembler(ensembler)
	if err != nil {
		return Error(httpStatus, "failed to delete the ensembler", err.Error())
	}

	return Ok(ensembler.GetID())
}

func (c EnsemblersController) checkActiveRouterVersion(options EnsemblersPathOptions) (int, error) {
	routerVersionStatusActive := []models.RouterVersionStatus{
		models.RouterVersionStatusDeployed,
		models.RouterVersionStatusPending,
	}
	activeOption := service.RouterVersionListOptions{
		ProjectID:   options.ProjectID,
		EnsemblerID: options.EnsemblerID,
		Statuses:    routerVersionStatusActive,
	}
	activeRouter, err := c.RouterVersionsService.ListRouterVersionsWithFilter(activeOption)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if len(activeRouter) >= 1 {
		return http.StatusBadRequest, fmt.Errorf("there are active router version using this ensembler")
	}
	return http.StatusOK, nil
}

func (c EnsemblersController) checkCurrentRouterVersion(options EnsemblersPathOptions) (int, error) {
	activeOption := service.RouterVersionListOptions{
		ProjectID:   options.ProjectID,
		EnsemblerID: options.EnsemblerID,
		IsCurrent:   true,
	}
	activeRouter, err := c.RouterVersionsService.ListRouterVersionsWithFilter(activeOption)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if len(activeRouter) >= 1 {
		return http.StatusBadRequest, fmt.Errorf("" +
			"there are router version that is currently being used by a router using this ensembler",
		)
	}
	return http.StatusOK, nil
}

func (c EnsemblersController) checkActiveEnsemblingJob(options EnsemblersPathOptions) (int, error) {
	ensemblingJobActiveOption := service.EnsemblingJobListOptions{
		EnsemblerID: options.EnsemblerID,
		Statuses: []models.Status{
			models.JobPending,
			models.JobBuildingImage,
			models.JobRunning,
		},
	}
	activeEnsemblingJobs, err := c.EnsemblingJobService.List(ensemblingJobActiveOption)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if activeEnsemblingJobs.Paging.Total >= 1 {
		return http.StatusBadRequest, fmt.Errorf("there are active ensembling job using this ensembler")
	}
	return http.StatusOK, nil
}
func (c EnsemblersController) deleteInactiveRouterVersion(options EnsemblersPathOptions) (int, error) {
	routerVersionStatusInactive := []models.RouterVersionStatus{
		models.RouterVersionStatusFailed,
		models.RouterVersionStatusUndeployed,
	}
	inactiveOption := service.RouterVersionListOptions{
		ProjectID:   options.ProjectID,
		EnsemblerID: options.EnsemblerID,
		Statuses:    routerVersionStatusInactive,
	}
	inactiveRouter, err := c.RouterVersionsService.ListRouterVersionsWithFilter(inactiveOption)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	for _, routerVersion := range inactiveRouter {
		err = c.RouterVersionsService.Delete(routerVersion)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}
	return http.StatusOK, nil
}
func (c EnsemblersController) deleteInactiveEnsemblingJob(options EnsemblersPathOptions) (int, error) {
	ensemblingJobInactiveOption := service.EnsemblingJobListOptions{
		EnsemblerID: options.EnsemblerID,
		Statuses: []models.Status{
			models.JobFailed,
			models.JobCompleted,
			models.JobFailedBuildImage,
			models.JobFailedSubmission,
		},
	}
	inactiveEnsemblingJobs, err := c.EnsemblingJobService.List(ensemblingJobInactiveOption)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if inactiveEnsemblingJobs.Paging.Total > 0 {
		if results, ok := inactiveEnsemblingJobs.Results.([]*models.EnsemblingJob); ok {
			for _, ensemblingJob := range results {
				if ensemblingJob.Status == models.JobFailed {
					err = c.EnsemblingJobService.Delete(ensemblingJob)
					if err != nil {
						return http.StatusInternalServerError, err
					}
				} else {
					err = c.EnsemblingJobService.MarkEnsemblingJobForTermination(ensemblingJob)
					if err != nil {
						return http.StatusInternalServerError, err
					}
				}
			}
		}
	}
	return http.StatusOK, nil
}

func (c EnsemblersController) deleteEnsembler(ensembler models.EnsemblerLike) (int, error) {
	// check if the ensembler is a pyfunc ensembler
	// convert ensembler to a PyfuncEnsembler Model, if ok (the model is a pyfunc ensembler)
	// then proceed to delete pyfunc artifact from mlflow
	if pyFuncEnsembler, ok := ensembler.(*models.PyFuncEnsembler); ok {
		// if the ensembler is a pyfunc ensembler, also delete from mlflow
		// Convert id to string
		s := strconv.FormatUint(uint64(pyFuncEnsembler.ExperimentID), 10)
		err := c.MlflowService.DeleteExperiment(context.Background(), s, true)
		if err != nil {
			return http.StatusInternalServerError, err
		}

	}

	// delete ensembler from database
	err := c.EnsemblersService.Delete(ensembler)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil

}
func (c EnsemblersController) Routes() []Route {
	return []Route{
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/ensemblers",
			handler: c.ListEnsemblers,
		},
		{
			method:  http.MethodPost,
			path:    "/projects/{project_id}/ensemblers",
			body:    request.CreateOrUpdateEnsemblerRequest{},
			handler: c.CreateEnsembler,
		},
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/ensemblers/{ensembler_id}",
			handler: c.GetEnsembler,
		},
		{
			method:  http.MethodPut,
			path:    "/projects/{project_id}/ensemblers/{ensembler_id}",
			body:    request.CreateOrUpdateEnsemblerRequest{},
			handler: c.UpdateEnsembler,
		},
		{
			method:  http.MethodDelete,
			path:    "/projects/{project_id}/ensemblers/{ensembler_id}",
			handler: c.DeleteEnsembler,
		},
	}
}

type EnsemblersPathOptions struct {
	ProjectID   *models.ID `schema:"project_id" validate:"required"`
	EnsemblerID *models.ID `schema:"ensembler_id" validate:"required"`
}
