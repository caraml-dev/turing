package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/caraml-dev/turing/api/turing/models"
)

type AlertsController struct {
	BaseController
}

var ErrAlertDisabled = errors.New("alert is disabled in turing-api")

func (c AlertsController) CreateAlert(r *http.Request, vars RequestVars, body interface{}) *Response {
	if c.AlertService == nil {
		return BadRequest(ErrAlertDisabled.Error(), "")
	}

	// Parse input
	var errResp *Response
	var router *models.Router
	var email string
	if _, errResp = c.getProjectFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if email, errResp = c.getEmailFromRequestHeader(r); errResp != nil {
		return errResp
	}

	// Create alert
	alert := body.(*models.Alert)
	alert.Service = c.getService(*router)

	environment, err := c.MLPService.GetEnvironment(router.EnvironmentName)
	if err != nil {
		return InternalServerError("unable to get MLP environment for the router", err.Error())
	}

	project, err := c.MLPService.GetProject(router.ProjectID)
	if err != nil {
		return InternalServerError("unable to get MLP project for the router", err.Error())
	}

	dashboardURL, err := c.AlertService.GetDashboardURL(alert, project, environment, router, nil)
	if err != nil {
		return InternalServerError("unable to generate dashboard URL for the alert", err.Error())
	}
	created, err := c.AlertService.Save(*alert, email, dashboardURL)
	if err != nil {
		return InternalServerError("unable to create alert", err.Error())
	}
	return Ok(created)
}

func (c AlertsController) ListAlerts(_ *http.Request, vars RequestVars, _ interface{}) *Response {
	if c.AlertService == nil {
		return BadRequest(ErrAlertDisabled.Error(), "")
	}

	// Parse input
	var errResp *Response
	var router *models.Router
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// List alerts
	alerts, err := c.AlertService.List(c.getService(*router))
	if err != nil {
		return InternalServerError("failed to list alerts", err.Error())
	}
	return Ok(alerts)
}

func (c AlertsController) GetAlert(_ *http.Request, vars RequestVars, _ interface{}) *Response {
	if c.AlertService == nil {
		return BadRequest(ErrAlertDisabled.Error(), "")
	}

	// Parse input
	var errResp *Response
	var alert *models.Alert
	if alert, errResp = c.getAlertFromRequestVars(vars); errResp != nil {
		return errResp
	}

	// Return alert
	return Ok(alert)
}

func (c AlertsController) UpdateAlert(r *http.Request, vars RequestVars, body interface{}) *Response {
	if c.AlertService == nil {
		return BadRequest(ErrAlertDisabled.Error(), "")
	}

	// Parse input
	var errResp *Response
	var router *models.Router
	var alert *models.Alert
	var email string
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if alert, errResp = c.getAlertFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if email, errResp = c.getEmailFromRequestHeader(r); errResp != nil {
		return errResp
	}

	// Update alert
	updateAlert := body.(*models.Alert)
	updateAlert.ID = alert.ID
	updateAlert.Service = c.getService(*router)

	environment, err := c.MLPService.GetEnvironment(router.EnvironmentName)
	if err != nil {
		return InternalServerError("unable to get MLP environment for the router", err.Error())
	}

	project, err := c.MLPService.GetProject(router.ProjectID)
	if err != nil {
		return InternalServerError("unable to get MLP project for the router", err.Error())
	}

	dashboardURL, err := c.AlertService.GetDashboardURL(alert, project, environment, router, nil)
	if err != nil {
		return InternalServerError("unable to generate dashboard URL for the alert", err.Error())
	}
	if err := c.AlertService.Update(*updateAlert, email, dashboardURL); err != nil {
		return InternalServerError("unable to update alert", err.Error())
	}
	return Ok(updateAlert)
}

func (c AlertsController) DeleteAlert(r *http.Request, vars RequestVars, _ interface{}) *Response {
	if c.AlertService == nil {
		return BadRequest(ErrAlertDisabled.Error(), "")
	}

	// Parse input
	var errResp *Response
	var router *models.Router
	var alert *models.Alert
	var email string
	if router, errResp = c.getRouterFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if alert, errResp = c.getAlertFromRequestVars(vars); errResp != nil {
		return errResp
	}
	if email, errResp = c.getEmailFromRequestHeader(r); errResp != nil {
		return errResp
	}

	environment, err := c.MLPService.GetEnvironment(router.EnvironmentName)
	if err != nil {
		return InternalServerError("unable to get MLP environment for the router", err.Error())
	}

	project, err := c.MLPService.GetProject(router.ProjectID)
	if err != nil {
		return InternalServerError("unable to get MLP project for the router", err.Error())
	}

	dashboardURL, err := c.AlertService.GetDashboardURL(alert, project, environment, router, nil)
	if err != nil {
		return InternalServerError("unable to generate dashboard URL for the alert", err.Error())
	}

	// Delete Alert
	if err := c.AlertService.Delete(*alert, email, dashboardURL); err != nil {
		return InternalServerError("unable to delete alert", err.Error())
	}
	return Ok(fmt.Sprintf("Alert with id '%d' deleted", alert.ID))
}

// getEmailFromRequestHeader ensures the request has a User-Email header (the account that sends the request)
func (c AlertsController) getEmailFromRequestHeader(r *http.Request) (string, *Response) {
	email := r.Header.Get("User-Email")
	if email == "" || !strings.Contains(email, "@") {
		return email, BadRequest("missing User-Email in header", "")
	}
	return email, nil
}

func (c AlertsController) getAlertFromRequestVars(vars RequestVars) (*models.Alert, *Response) {
	id, err := getIDFromVars(vars, "alert_id")
	if err != nil {
		return nil, BadRequest("invalid alert id", err.Error())
	}
	alert, err := c.AlertService.FindByID(id)
	if err != nil {
		return nil, NotFound("alert not found", err.Error())
	}
	return alert, nil
}

// getService returns service name from router name.
// The service name is assumed to be <router_name>-turing-router
func (c AlertsController) getService(r models.Router) string {
	return r.Name + "-turing-router"
}

func (c AlertsController) Routes() []Route {
	return []Route{
		{
			method:  http.MethodPost,
			path:    "/projects/{project_id}/routers/{router_id}/alerts",
			body:    models.Alert{},
			handler: c.CreateAlert,
		},
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/alerts",
			handler: c.ListAlerts,
		},
		{
			method:  http.MethodPut,
			path:    "/projects/{project_id}/routers/{router_id}/alerts/{alert_id}",
			body:    models.Alert{},
			handler: c.UpdateAlert,
		},
		{
			method:  http.MethodGet,
			path:    "/projects/{project_id}/routers/{router_id}/alerts/{alert_id}",
			body:    models.Alert{},
			handler: c.GetAlert,
		},
		{
			method:  http.MethodDelete,
			path:    "/projects/{project_id}/routers/{router_id}/alerts/{alert_id}",
			handler: c.DeleteAlert,
		},
	}
}
