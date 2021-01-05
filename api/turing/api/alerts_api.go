package api

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/gojek/turing/api/turing/service"

	"github.com/gojek/turing/api/turing/models"
)

type AlertsController struct {
	*baseController
}

var ErrAlertDisabled = errors.New("alert is disabled in turing-api")

func (c *AlertsController) CreateAlert(r *http.Request, vars map[string]string, body interface{}) *Response {
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
	dashboardURL, err := c.getDashboardURL(c.AlertService.GetDashboardURLTemplate(), alert, router, nil)
	if err != nil {
		return InternalServerError("unable to generate dashboard URL for the alert", err.Error())
	}
	created, err := c.AlertService.Save(*alert, email, dashboardURL)
	if err != nil {
		return InternalServerError("unable to create alert", err.Error())
	}
	return Ok(created)
}

func (c *AlertsController) ListAlerts(r *http.Request, vars map[string]string, _ interface{}) *Response {
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

func (c *AlertsController) GetAlert(r *http.Request, vars map[string]string, _ interface{}) *Response {
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

func (c *AlertsController) UpdateAlert(r *http.Request, vars map[string]string, body interface{}) *Response {
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
	dashboardURL, err := c.getDashboardURL(c.AlertService.GetDashboardURLTemplate(), alert, router, nil)
	if err != nil {
		return InternalServerError("unable to generate dashboard URL for the alert", err.Error())
	}
	if err := c.AlertService.Update(*updateAlert, email, dashboardURL); err != nil {
		return InternalServerError("unable to update alert", err.Error())
	}
	return Ok(updateAlert)
}

func (c *AlertsController) DeleteAlert(r *http.Request, vars map[string]string, _ interface{}) *Response {
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

	dashboardURL, err := c.getDashboardURL(c.AlertService.GetDashboardURLTemplate(), alert, router, nil)
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
func (c *AlertsController) getEmailFromRequestHeader(r *http.Request) (string, *Response) {
	email := r.Header.Get("User-Email")
	if email == "" || !strings.Contains(email, "@") {
		return email, BadRequest("missing User-Email in header", "")
	}
	return email, nil
}

func (c *AlertsController) getAlertFromRequestVars(vars map[string]string) (*models.Alert, *Response) {
	id, err := getIntFromVars(vars, "alert_id")
	if err != nil {
		return nil, BadRequest("invalid alert id", err.Error())
	}
	alert, err := c.AlertService.FindByID(uint(id))
	if err != nil {
		return nil, NotFound("alert not found", err.Error())
	}
	return alert, nil
}

// getService returns service name from router name.
// The service name is assumed to be <router_name>-turing-router
func (c *AlertsController) getService(r models.Router) string {
	return r.Name + "-turing-router"
}

// getDashboardURL returns the dashboard URL for the router alert given a dashboardURL template
// from the alertService. The template will be executed with DashboardURLValue.
//
// If routerVersion is nil, the dashboard URL should return the dashboard showing metrics
// for the router across all revisions. Else, the dashboard should show metrics for a specific
// router version.
//
// If the MLPService fails to resolve the MLP environment and project for the router,
// an error will be returned.
func (c *AlertsController) getDashboardURL(
	template template.Template,
	alert *models.Alert,
	router *models.Router,
	routerVersion *models.RouterVersion) (string, error) {
	if alert == nil || router == nil {
		return "", nil
	}

	environment, err := c.MLPService.GetEnvironment(router.EnvironmentName)
	if err != nil {
		return "", err
	}

	project, err := c.MLPService.GetProject(router.ProjectID)
	if err != nil {
		return "", err
	}

	var revision string
	if routerVersion != nil {
		revision = fmt.Sprintf("%d", routerVersion.Version)
	} else {
		revision = "$__all"
	}

	value := service.DashboardURLValue{
		Environment: alert.Environment,
		Cluster:     environment.Cluster,
		Project:     project.Name,
		Router:      router.Name,
		Revision:    revision,
	}

	var buf bytes.Buffer
	err = template.Execute(&buf, value)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
