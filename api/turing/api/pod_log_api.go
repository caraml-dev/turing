package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gojek/turing/api/turing/batch"
	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/service"

	mlp "github.com/gojek/mlp/api/client"
	"github.com/gojek/turing/api/turing/cluster/servicebuilder"
	"github.com/gojek/turing/api/turing/models"
)

type PodLogController struct {
	BaseController
}

var (
	batchEnsemblingAllowedComponents = map[string]struct{}{
		batch.ImageBuilderPodType: {},
		batch.DriverPodType:       {},
		batch.ExecutorPodType:     {},
	}
	routerAllowedComponents = map[string]struct{}{
		servicebuilder.ComponentTypes.Router:    {},
		servicebuilder.ComponentTypes.Enricher:  {},
		servicebuilder.ComponentTypes.Ensembler: {},
	}
)

// ListRouterPodLogs handles the HTTP request for getting Router Pod Logs
// It supports 3 types of pods:
//   1. Router
//   2. Enricher
//   3. Ensembler
func (c PodLogController) ListRouterPodLogs(_ *http.Request, vars RequestVars, _ interface{}) *Response {
	// Parse input
	var errResp *Response
	var project *mlp.Project
	var router *models.Router
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}

	var routerVersion *models.RouterVersion
	var err error
	if _, ok := vars.get("version"); ok {
		// Specific router version value is requested
		routerVersion, errResp = c.getRouterVersionFromRequestVars(vars)
		if errResp != nil {
			return errResp
		}
	} else {
		// Unspecified router version, so default to current router version (assumed to be not null i.e valid value)
		if !router.CurrRouterVersionID.Valid {
			return BadRequest("Current router version id is invalid", "Make sure current router is deployed")
		}
		routerVersion, err = c.RouterVersionsService.FindByID(models.ID(router.CurrRouterVersionID.Int32))
		if err != nil {
			return InternalServerError("Failed to find current router version", err.Error())
		}
	}

	componentType, err := getComponentType(servicebuilder.ComponentTypes.Router, vars, routerAllowedComponents)
	if err != nil {
		return BadRequest(err.Error(), "must be one of router, enricher or ensembler")
	}

	request := service.PodLogRequest{
		Namespace:        servicebuilder.GetNamespace(project),
		DefaultContainer: cluster.KnativeUserContainerName,
		Environment:      router.EnvironmentName,
		LabelSelectors: []service.LabelSelector{
			{
				Key:   cluster.KnativeServiceLabelKey,
				Value: servicebuilder.GetComponentName(routerVersion, componentType),
			},
		},
	}
	varLogError := c.parsePodLogRequest(&request, vars)

	if varLogError != nil {
		return BadRequest(varLogError.Description, varLogError.ErrorMessage)
	}

	logs, err := c.PodLogService.ListPodLogs(request)
	if err != nil {
		return InternalServerError("Failed to list logs", err.Error())
	}

	return Ok(logs)
}

// ListEnsemblingJobPodLogs handles the HTTP request for getting Ensembling Pod Logs
// It supports 3 types of pods:
//   1. Image Builder
//   2. Spark Driver
//   3. Spark Executor
func (c PodLogController) ListEnsemblingJobPodLogs(_ *http.Request, vars RequestVars, _ interface{}) *Response {
	var errResp *Response
	var project *mlp.Project
	if project, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}

	componentType, err := getComponentType(batch.ImageBuilderPodType, vars, batchEnsemblingAllowedComponents)
	if err != nil {
		return BadRequest(err.Error(), fmt.Sprintf(
			"must be one of the following: %s, %s or %s",
			batch.ImageBuilderPodType,
			batch.DriverPodType,
			batch.ExecutorPodType,
		))
	}

	options := &ListEnsemblingPodLogsOptions{}
	if err := c.ParseVars(options, vars); err != nil {
		return BadRequest(
			"failed to fetch ensembling job",
			fmt.Sprintf("failed to parse query string: %s", err),
		)
	}

	ensemblingJob, err := c.EnsemblingJobService.FindByID(
		*options.ID,
		service.EnsemblingJobFindByIDOptions{
			ProjectID: options.ProjectID,
		},
	)

	if err != nil {
		return NotFound("ensembling job not found", err.Error())
	}

	namespace := c.EnsemblingJobService.GetNamespaceByComponent(componentType, project)
	environment := c.EnsemblingJobService.GetDefaultEnvironment()
	ensemblerName := ensemblingJob.InfraConfig.EnsemblerName
	request := service.PodLogRequest{
		Namespace:   namespace,
		Environment: environment,
		LabelSelectors: c.EnsemblingJobService.CreatePodLabelSelector(
			ensemblerName,
			componentType,
		),
	}

	varLogError := c.parsePodLogRequest(&request, vars)
	if varLogError != nil {
		return BadRequest(varLogError.Description, varLogError.ErrorMessage)
	}

	legacyFormatLogs, err := c.PodLogService.ListPodLogs(request)
	if err != nil {
		return InternalServerError("Failed to list logs", err.Error())
	}

	loggingURL, err := c.EnsemblingJobService.FormatLoggingURL(ensemblerName, namespace, componentType)
	if err != nil {
		return InternalServerError("Failed to format monitoring URL", err.Error())
	}

	logs := service.ConvertPodLogsToV2(
		namespace,
		environment,
		loggingURL,
		legacyFormatLogs,
	)

	return Ok(logs)
}

func getComponentType(defaultComponentType string, vars RequestVars, allowedTypes map[string]struct{}) (string, error) {
	componentType, _ := vars.get("component_type")
	if componentType == "" {
		return defaultComponentType, nil
	}

	if _, ok := allowedTypes[componentType]; ok {
		return componentType, nil
	}

	return "", fmt.Errorf("Invalid component type '%s'", componentType)
}

type ListEnsemblingPodLogsOptions struct {
	ProjectID *models.ID `schema:"project_id" validate:"required"`
	ID        *models.ID `schema:"job_id" validate:"required"`
}

type logVarParseError struct {
	Description  string
	ErrorMessage string
}

func (c PodLogController) parsePodLogRequest(
	podLogRequest *service.PodLogRequest,
	vars RequestVars,
) *logVarParseError {
	podLogRequest.Container, _ = vars.get("container")

	if previous, ok := vars.get("previous"); ok {
		previous, err := strconv.ParseBool(previous)
		if err != nil {
			return &logVarParseError{
				Description:  "Query string 'previous' must be a truthy value",
				ErrorMessage: err.Error(),
			}
		}
		podLogRequest.Previous = previous
	}

	if sinceTime, ok := vars.get("since_time"); ok {
		t, err := time.Parse(time.RFC3339, sinceTime)
		if err != nil {
			return &logVarParseError{
				Description:  "Query string 'since_time' must be in RFC3339 format",
				ErrorMessage: err.Error(),
			}
		}
		podLogRequest.SinceTime = &t
	}

	// sometimes the client passes tail_lines= and this causes ok to be true with empty string.
	if tailLines, _ := vars.get("tail_lines"); tailLines != "" {
		i, err := strconv.ParseInt(tailLines, 10, 64)
		if err != nil {
			return &logVarParseError{
				Description:  "Query string 'tail_lines' must be a positive number",
				ErrorMessage: err.Error(),
			}
		}
		if i <= 0 {
			return &logVarParseError{
				Description: "Query string 'tail_lines' must be a positive number",
			}
		}
		podLogRequest.TailLines = &i
	}

	// sometimes the client passes head_lines= and this causes ok to be true with empty string.
	if headLines, _ := vars.get("head_lines"); headLines != "" {
		i, err := strconv.ParseInt(headLines, 10, 64)
		if err != nil {
			return &logVarParseError{
				Description:  "Query string 'head_lines' must be a positive number",
				ErrorMessage: err.Error(),
			}
		}
		if i <= 0 {
			return &logVarParseError{
				Description: "Query string 'head_lines' must be a positive number",
			}
		}
		podLogRequest.HeadLines = &i
	}
	return nil
}

func (c PodLogController) Routes() []Route {
	return []Route{
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/logs",
			handler: c.ListRouterPodLogs,
		},
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/versions/{version}/logs",
			handler: c.ListRouterPodLogs,
		},
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/jobs/{job_id}/logs",
			handler: c.ListEnsemblingJobPodLogs,
		},
	}
}
