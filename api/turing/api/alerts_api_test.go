package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"
	testifyAssert "github.com/stretchr/testify/assert"
	"gotest.tools/assert"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service/mocks"
)

func TestAlertsControllerWhenAlertIsDisabled(t *testing.T) {
	controller := &AlertsController{
		BaseController{
			AppContext: &AppContext{
				MLPService:     &mocks.MLPService{},
				RoutersService: &mocks.RoutersService{},
				AlertService:   nil,
			},
		},
	}
	checkErr := func(t *testing.T, resp *Response) {
		assert.Equal(t, resp.code, http.StatusBadRequest)
		assert.Check(t, strings.Contains(fmt.Sprintf("%v", resp.data), ErrAlertDisabled.Error()))
	}

	checkErr(t, controller.CreateAlert(&http.Request{}, RequestVars{}, nil))
	checkErr(t, controller.ListAlerts(&http.Request{}, RequestVars{}, nil))
	checkErr(t, controller.GetAlert(&http.Request{}, RequestVars{}, nil))
	checkErr(t, controller.UpdateAlert(&http.Request{}, RequestVars{}, nil))
	checkErr(t, controller.DeleteAlert(&http.Request{}, RequestVars{}, nil))
}

func TestAlertsControllerCreateAlert(t *testing.T) {
	// Set up mock services
	project := &mlp.Project{Name: "testproject"}
	environment := &merlin.Environment{Cluster: "testcluster"}
	router := &models.Router{Name: "router"}
	var routerVersion *models.RouterVersion

	mockMLPService := &mocks.MLPService{}
	mockMLPService.
		On("GetProject", mock.AnythingOfType("models.ID")).
		Return(project, nil)
	mockMLPService.
		On("GetEnvironment", mock.AnythingOfType("string")).
		Return(environment, nil)

	mockRouterService := &mocks.RoutersService{}
	mockRouterService.On("FindByID", models.ID(1)).
		Return(nil, errors.New("test router error"))
	mockRouterService.On("FindByID", models.ID(2)).Return(router, nil)

	alert := &models.Alert{
		Environment:       "env",
		Team:              "team",
		Service:           "router-turing-router",
		Metric:            "throughput",
		WarningThreshold:  20,
		CriticalThreshold: 10,
		Duration:          "5m",
	}

	mockAlertService := &mocks.AlertService{}
	mockAlertService.
		On(
			"GetDashboardURL",
			alert,
			project,
			environment,
			router,
			routerVersion,
		).
		Return("https://grafana.example.com/dashboard?var-cluster=testcluster"+
			"&var-project=testproject&var-experiment=router&var-revision=$__all", nil)
	mockAlertService.
		On(
			"Save",
			*alert,
			"user@gojek.com",
			"https://grafana.example.com/dashboard?var-cluster=testcluster&var-project=testproject"+
				"&var-experiment=router&var-revision=$__all").
		Return(nil, errors.New("test alert error"))
	mockAlertService.
		On(
			"Save",
			*alert,
			"user2@gojek.com",
			"https://grafana.example.com/dashboard?var-cluster=testcluster&var-project=testproject"+
				"&var-experiment=router&var-revision=$__all").
		Return(alert, nil)

	// Define tests
	inputAlert := &models.Alert{
		Environment:       "env",
		Team:              "team",
		Metric:            "throughput",
		WarningThreshold:  20,
		CriticalThreshold: 10,
		Duration:          "5m",
	}
	tests := map[string]struct {
		req  *http.Request
		vars RequestVars
		body interface{}
		want *Response
	}{
		"failure | missing project_id": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{},
			body: &models.Alert{},
			want: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		"failure | router not found": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{"project_id": {"1"}, "router_id": {"1"}},
			body: inputAlert,
			want: NotFound("router not found", "test router error"),
		},
		"failure | email not found": {
			req: &http.Request{
				Header: http.Header{},
			},
			vars: RequestVars{"project_id": {"1"}, "router_id": {"2"}},
			body: inputAlert,
			want: BadRequest("missing User-Email in header", ""),
		},
		"failure | save alert": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{"project_id": {"1"}, "router_id": {"2"}},
			body: inputAlert,
			want: InternalServerError("unable to create alert", "test alert error"),
		},
		"success": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user2@gojek.com"}},
			},
			vars: RequestVars{"project_id": {"1"}, "router_id": {"2"}},
			body: inputAlert,
			want: Ok(alert),
		},
	}

	// Create test controller
	controller := &AlertsController{
		BaseController{
			AppContext: &AppContext{
				MLPService:     mockMLPService,
				RoutersService: mockRouterService,
				AlertService:   mockAlertService,
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			got := controller.CreateAlert(data.req, data.vars, data.body)
			assert.Equal(t, data.want.code, got.code)
			assert.DeepEqual(t, data.want.data, got.data)
		})
	}
}

func TestAlertsControllerListAlerts(t *testing.T) {
	// Set up mock services
	mockMLPService := &mocks.MLPService{}
	mockMLPService.On("GetProject", 1).Return(&mlp.Project{}, nil)

	mockRouterService := &mocks.RoutersService{}
	mockRouterService.On("FindByID", models.ID(1)).
		Return(nil, errors.New("test router error"))
	mockRouterService.On("FindByID", models.ID(2)).Return(&models.Router{Name: "router2"}, nil)
	mockRouterService.On("FindByID", models.ID(3)).Return(&models.Router{Name: "router3"}, nil)

	alerts := []*models.Alert{
		{
			Environment:       "env",
			Team:              "team",
			Service:           "router-turing-router",
			Metric:            "throughput",
			WarningThreshold:  20,
			CriticalThreshold: 10,
			Duration:          "5m",
		},
	}
	mockAlertService := &mocks.AlertService{}
	mockAlertService.On("List", "router2-turing-router").
		Return(nil, errors.New("test alert error"))
	mockAlertService.On("List", "router3-turing-router").Return(alerts, nil)

	// Define tests
	tests := map[string]struct {
		req  *http.Request
		vars RequestVars
		want *Response
	}{
		"failure | router not found": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{"router_id": {"1"}},
			want: NotFound("router not found", "test router error"),
		},
		"failure | save alert": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{"router_id": {"2"}},
			want: InternalServerError("failed to list alerts", "test alert error"),
		},
		"success": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user2@gojek.com"}},
			},
			vars: RequestVars{"router_id": {"3"}},
			want: Ok(alerts),
		},
	}

	// Create test controller
	controller := &AlertsController{
		BaseController{
			AppContext: &AppContext{
				MLPService:     mockMLPService,
				RoutersService: mockRouterService,
				AlertService:   mockAlertService,
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			got := controller.ListAlerts(data.req, data.vars, nil)
			assert.Equal(t, data.want.code, got.code)
			assert.DeepEqual(t, data.want.data, got.data)
		})
	}
}

func TestAlertsControllerGetAlert(t *testing.T) {
	// Set up mock services
	mockMLPService := &mocks.MLPService{}
	mockMLPService.On("GetProject", models.ID(1)).Return(&mlp.Project{}, nil)

	mockRouterService := &mocks.RoutersService{}
	mockRouterService.On("FindByID", models.ID(1)).Return(&models.Router{Name: "router"}, nil)

	alert := &models.Alert{
		Model:             models.Model{ID: 2},
		Environment:       "env",
		Team:              "team",
		Service:           "router-turing-router",
		Metric:            "throughput",
		WarningThreshold:  20,
		CriticalThreshold: 10,
		Duration:          "5m",
	}
	mockAlertService := &mocks.AlertService{}
	mockAlertService.On("FindByID", models.ID(1)).
		Return(nil, errors.New("test alert error"))
	mockAlertService.On("FindByID", models.ID(2)).Return(alert, nil)

	// Define tests
	tests := map[string]struct {
		req  *http.Request
		vars RequestVars
		want *Response
	}{
		"failure | alert not found": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{"alert_id": {"1"}},
			want: NotFound("alert not found", "test alert error"),
		},
		"success": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{"alert_id": {"2"}},
			want: Ok(alert),
		},
	}

	// Create test controller
	controller := &AlertsController{
		BaseController{
			AppContext: &AppContext{
				MLPService:     mockMLPService,
				RoutersService: mockRouterService,
				AlertService:   mockAlertService,
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			got := controller.GetAlert(data.req, data.vars, nil)
			assert.Equal(t, data.want.code, got.code)
			assert.DeepEqual(t, data.want.data, got.data)
		})
	}
}

func TestAlertsControllerUpdateAlert(t *testing.T) {
	// Set up mock services
	project := &mlp.Project{Name: "testproject"}
	environment := &merlin.Environment{Cluster: "testcluster"}
	router := &models.Router{Name: "test"}
	var routerVersion *models.RouterVersion

	mockMLPService := &mocks.MLPService{}
	mockMLPService.On("GetProject", mock.AnythingOfType("models.ID")).Return(project, nil)
	mockMLPService.
		On("GetEnvironment", mock.AnythingOfType("string")).
		Return(environment, nil)

	mockRouterService := &mocks.RoutersService{}
	mockRouterService.On("FindByID", models.ID(1)).Return(router, nil)
	mockRouterService.On("FindByID", models.ID(2)).Return(router, nil)

	oldAlert1 := &models.Alert{
		Model:             models.Model{ID: 1},
		Environment:       "env",
		Team:              "team",
		Metric:            "throughput",
		WarningThreshold:  2,
		CriticalThreshold: 1,
		Duration:          "5m",
	}
	oldAlert2 := &models.Alert{
		Model:             models.Model{ID: 2},
		Environment:       "env",
		Team:              "team",
		Metric:            "throughput",
		WarningThreshold:  2,
		CriticalThreshold: 1,
		Duration:          "5m",
	}
	alert1 := &models.Alert{
		Model:             models.Model{ID: 1},
		Environment:       "env",
		Team:              "team",
		Metric:            "throughput",
		WarningThreshold:  20,
		CriticalThreshold: 10,
		Duration:          "5m",
		Service:           "test-turing-router",
	}
	alert2 := &models.Alert{
		Model:             models.Model{ID: 2},
		Environment:       "env",
		Team:              "team",
		Metric:            "throughput",
		WarningThreshold:  20,
		CriticalThreshold: 10,
		Duration:          "5m",
		Service:           "test-turing-router",
	}
	mockAlertService := &mocks.AlertService{}
	mockAlertService.On("FindByID", models.ID(1)).Return(oldAlert1, nil)
	mockAlertService.On("FindByID", models.ID(2)).Return(oldAlert2, nil)
	mockAlertService.On("FindByID", models.ID(10)).
		Return(nil, errors.New("test alert find error"))
	mockAlertService.
		On(
			"GetDashboardURL",
			mock.AnythingOfType("*models.Alert"),
			project,
			environment,
			router,
			routerVersion,
		).
		Return("https://grafana.example.com/dashboard?var-cluster=testcluster"+
			"&var-project=testproject&var-experiment=test", nil)
	mockAlertService.
		On("Update", *alert1, "user@gojek.com", "https://grafana.example.com/dashboard?var-cluster=testcluster"+
			"&var-project=testproject&var-experiment=test").
		Return(errors.New("test alert error"))
	mockAlertService.
		On("Update", *alert2, "user@gojek.com", "https://grafana.example.com/dashboard?var-cluster=testcluster"+
			"&var-project=testproject&var-experiment=test").
		Return(nil)

	// Define tests
	body := &models.Alert{
		Environment:       "env",
		Team:              "team",
		Metric:            "throughput",
		WarningThreshold:  20,
		CriticalThreshold: 10,
		Duration:          "5m",
	}
	tests := map[string]struct {
		req  *http.Request
		vars RequestVars
		body interface{}
		want *Response
	}{
		"failure | missing router_id": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{},
			body: body,
			want: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | alert not found": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{"router_id": {"1"}, "alert_id": {"10"}},
			body: body,
			want: NotFound("alert not found", "test alert find error"),
		},
		"failure | missing email": {
			req: &http.Request{
				Header: http.Header{},
			},
			vars: RequestVars{"router_id": {"1"}, "alert_id": {"1"}},
			want: BadRequest("missing User-Email in header", ""),
		},
		"failure | update alert": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{"router_id": {"2"}, "alert_id": {"1"}},
			body: body,
			want: InternalServerError("unable to update alert", "test alert error"),
		},
		"success": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{"router_id": {"2"}, "alert_id": {"2"}},
			body: body,
			want: Ok(alert2),
		},
	}

	// Create test controller
	controller := &AlertsController{
		BaseController{
			AppContext: &AppContext{
				MLPService:     mockMLPService,
				RoutersService: mockRouterService,
				AlertService:   mockAlertService,
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			got := controller.UpdateAlert(data.req, data.vars, data.body)
			assert.Equal(t, data.want.code, got.code)
			assert.DeepEqual(t, data.want.data, got.data)
		})
	}
}

func TestAlertsControllerDeleteAlert(t *testing.T) {
	// Set up mock services
	project := &mlp.Project{Name: "testproject"}
	environment := &merlin.Environment{Cluster: "testcluster"}
	router := &models.Router{Name: "test"}
	var routerVersion *models.RouterVersion

	mockMLPService := &mocks.MLPService{}
	mockMLPService.On("GetProject", mock.AnythingOfType("models.ID")).Return(project, nil)
	mockMLPService.
		On("GetEnvironment", mock.AnythingOfType("string")).
		Return(environment, nil)

	mockRouterService := &mocks.RoutersService{}
	mockRouterService.On("FindByID", models.ID(1)).Return(router, nil)
	mockRouterService.On("FindByID", models.ID(2)).Return(router, nil)

	alert1 := &models.Alert{Model: models.Model{ID: 1}}
	alert2 := &models.Alert{Model: models.Model{ID: 2}}
	mockAlertService := &mocks.AlertService{}
	mockAlertService.On("FindByID", models.ID(1)).Return(alert1, nil)
	mockAlertService.On("FindByID", models.ID(2)).Return(alert2, nil)
	mockAlertService.On("FindByID", models.ID(10)).
		Return(nil, errors.New("test alert find error"))
	mockAlertService.
		On(
			"GetDashboardURL",
			mock.AnythingOfType("*models.Alert"),
			project,
			environment,
			router,
			routerVersion,
		).
		Return("https://grafana.example.com/dashboard?var-cluster=testcluster"+
			"&var-project=testproject&var-experiment=test", nil)
	mockAlertService.On(
		"Delete",
		*alert1,
		"user@gojek.com",
		"https://grafana.example.com/dashboard?var-cluster=testcluster&var-project=testproject&var-experiment=test").
		Return(errors.New("test alert error"))
	mockAlertService.On(
		"Delete",
		*alert2,
		"user@gojek.com",
		"https://grafana.example.com/dashboard?var-cluster=testcluster&var-project=testproject&var-experiment=test").
		Return(nil)

	// Delete tests
	tests := map[string]struct {
		req  *http.Request
		vars RequestVars
		want *Response
	}{
		"failure | alert not found": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{"alert_id": {"10"}, "router_id": {"1"}},
			want: NotFound("alert not found", "test alert find error"),
		},
		"failure | missing email": {
			req: &http.Request{
				Header: http.Header{},
			},
			vars: RequestVars{"alert_id": {"1"}, "router_id": {"1"}},
			want: BadRequest("missing User-Email in header", ""),
		},
		"failure | delete alert": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{"alert_id": {"1"}, "router_id": {"1"}},
			want: InternalServerError("unable to delete alert", "test alert error"),
		},
		"success": {
			req: &http.Request{
				Header: http.Header{"User-Email": {"user@gojek.com"}},
			},
			vars: RequestVars{"alert_id": {"2"}, "router_id": {"1"}},
			want: Ok("Alert with id '2' deleted"),
		},
	}

	// Create test controller
	controller := &AlertsController{
		BaseController{
			AppContext: &AppContext{
				MLPService:     mockMLPService,
				RoutersService: mockRouterService,
				AlertService:   mockAlertService,
			},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			got := controller.DeleteAlert(data.req, data.vars, nil)
			assert.Equal(t, data.want.code, got.code)
			assert.DeepEqual(t, data.want.data, got.data)
		})
	}
}

func TestAlertsControllerGetEmailFromRequestHeader(t *testing.T) {
	// Create test requests
	testRequestSuccess := &http.Request{
		Header: http.Header{"User-Email": {"test@abc.com"}},
	}
	testRequestFailure := &http.Request{
		Header: http.Header{"User-Email": {"test-email"}},
	}
	testRequestFailureEmpty := &http.Request{
		Header: http.Header{},
	}

	// Define test cases
	tests := map[string]struct {
		request  *http.Request
		expected *Response
	}{
		"failure | bad email": {
			request:  testRequestFailure,
			expected: BadRequest("missing User-Email in header", ""),
		},
		"failure | empty email": {
			request:  testRequestFailureEmpty,
			expected: BadRequest("missing User-Email in header", ""),
		},
		"success": {
			request:  testRequestSuccess,
			expected: nil,
		},
	}

	// Validate
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &AlertsController{}
			// Run test method and validate
			_, response := ctrl.getEmailFromRequestHeader(data.request)
			testifyAssert.Equal(t, data.expected, response)
		})
	}
}

func TestAlertsControllerGetAlertFromRequestVars(t *testing.T) {
	// Create mock services
	alertSvc := &mocks.AlertService{}
	alertSvc.On("FindByID", models.ID(1)).
		Return(nil, errors.New("test alert error"))
	alertSvc.On("FindByID", models.ID(2)).Return(&models.Alert{}, nil)

	// Define test cases
	tests := map[string]struct {
		vars     RequestVars
		expected *Response
	}{
		"failure | invalid alert id": {
			vars:     RequestVars{},
			expected: BadRequest("invalid alert id", "key alert_id not found in vars"),
		},
		"failure | alert not found": {
			vars:     RequestVars{"alert_id": {"1"}},
			expected: NotFound("alert not found", "test alert error"),
		},
		"success": {
			vars:     RequestVars{"alert_id": {"2"}},
			expected: nil,
		},
	}

	// Validate
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &AlertsController{
				BaseController{
					AppContext: &AppContext{
						AlertService: alertSvc,
					},
				},
			}
			// Run test method and validate
			_, response := ctrl.getAlertFromRequestVars(data.vars)
			testifyAssert.Equal(t, data.expected, response)
		})
	}
}
