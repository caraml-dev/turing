package api

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"strings"

	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service/mocks"
	testifyAssert "github.com/stretchr/testify/assert"
	"gotest.tools/assert"
)

func TestAlertsContollerWhenAlertIsDisabled(t *testing.T) {
	controller := &AlertsController{
		&baseController{
			&AppContext{
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

	checkErr(t, controller.CreateAlert(&http.Request{}, map[string]string{}, nil))
	checkErr(t, controller.ListAlerts(&http.Request{}, map[string]string{}, nil))
	checkErr(t, controller.GetAlert(&http.Request{}, map[string]string{}, nil))
	checkErr(t, controller.UpdateAlert(&http.Request{}, map[string]string{}, nil))
	checkErr(t, controller.DeleteAlert(&http.Request{}, map[string]string{}, nil))
}

func TestAlertsControllerCreateAlert(t *testing.T) {
	// Set up mock services
	mockMLPService := &mocks.MLPService{}
	mockMLPService.On("GetProject", 1).Return(&mlp.Project{}, nil)

	mockRouterService := &mocks.RoutersService{}
	mockRouterService.On("FindByID", uint(1)).Return(nil, errors.New("Test router error"))
	mockRouterService.On("FindByID", uint(2)).Return(&models.Router{Name: "router"}, nil)

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
	mockAlertService.On("Save", *alert, "user@gojek.com").Return(nil, errors.New("Test alert error"))
	mockAlertService.On("Save", *alert, "user2@gojek.com").Return(alert, nil)

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
		vars map[string]string
		body interface{}
		want *Response
	}{
		"failure | missing project_id": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{},
			body: &models.Alert{},
			want: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		"failure | router not found": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{"project_id": "1", "router_id": "1"},
			body: inputAlert,
			want: NotFound("router not found", "Test router error"),
		},
		"failure | email not found": {
			req: &http.Request{
				Header: map[string][]string{},
			},
			vars: map[string]string{"project_id": "1", "router_id": "2"},
			body: inputAlert,
			want: BadRequest("missing User-Email in header", ""),
		},
		"failure | save alert": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{"project_id": "1", "router_id": "2"},
			body: inputAlert,
			want: InternalServerError("unable to create alert", "Test alert error"),
		},
		"success": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user2@gojek.com"}},
			},
			vars: map[string]string{"project_id": "1", "router_id": "2"},
			body: inputAlert,
			want: Ok(alert),
		},
	}

	// Create test controller
	controller := &AlertsController{
		&baseController{
			&AppContext{
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
	mockRouterService.On("FindByID", uint(1)).Return(nil, errors.New("Test router error"))
	mockRouterService.On("FindByID", uint(2)).Return(&models.Router{Name: "router2"}, nil)
	mockRouterService.On("FindByID", uint(3)).Return(&models.Router{Name: "router3"}, nil)

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
	mockAlertService.On("List", "router2-turing-router").Return(nil, errors.New("Test alert error"))
	mockAlertService.On("List", "router3-turing-router").Return(alerts, nil)

	// Define tests
	tests := map[string]struct {
		req  *http.Request
		vars map[string]string
		want *Response
	}{
		"failure | router not found": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{"router_id": "1"},
			want: NotFound("router not found", "Test router error"),
		},
		"failure | save alert": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{"router_id": "2"},
			want: InternalServerError("failed to list alerts", "Test alert error"),
		},
		"success": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user2@gojek.com"}},
			},
			vars: map[string]string{"router_id": "3"},
			want: Ok(alerts),
		},
	}

	// Create test controller
	controller := &AlertsController{
		&baseController{
			&AppContext{
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
	mockMLPService.On("GetProject", 1).Return(&mlp.Project{}, nil)

	mockRouterService := &mocks.RoutersService{}
	mockRouterService.On("FindByID", uint(1)).Return(&models.Router{Name: "router"}, nil)

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
	mockAlertService.On("FindByID", uint(1)).Return(nil, errors.New("Test alert error"))
	mockAlertService.On("FindByID", uint(2)).Return(alert, nil)

	// Define tests
	tests := map[string]struct {
		req  *http.Request
		vars map[string]string
		want *Response
	}{
		"failure | alert not found": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{"alert_id": "1"},
			want: NotFound("alert not found", "Test alert error"),
		},
		"success": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{"alert_id": "2"},
			want: Ok(alert),
		},
	}

	// Create test controller
	controller := &AlertsController{
		&baseController{
			&AppContext{
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
	mockMLPService := &mocks.MLPService{}
	mockMLPService.On("GetProject", 1).Return(&mlp.Project{}, nil)

	mockRouterService := &mocks.RoutersService{}
	mockRouterService.On("FindByID", uint(1)).Return(&models.Router{Name: "test"}, nil)
	mockRouterService.On("FindByID", uint(2)).Return(&models.Router{Name: "test"}, nil)

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
	mockAlertService.On("FindByID", uint(1)).Return(oldAlert1, nil)
	mockAlertService.On("FindByID", uint(2)).Return(oldAlert2, nil)
	mockAlertService.On("FindByID", uint(10)).Return(nil, errors.New("Test alert find error"))
	mockAlertService.On("Update", *alert1, "user@gojek.com").Return(errors.New("Test alert error"))
	mockAlertService.On("Update", *alert2, "user@gojek.com").Return(nil)

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
		vars map[string]string
		body interface{}
		want *Response
	}{
		"failure | missing router_id": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{},
			body: body,
			want: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		"failure | alert not found": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{"router_id": "1", "alert_id": "10"},
			body: body,
			want: NotFound("alert not found", "Test alert find error"),
		},
		"failure | missing email": {
			req: &http.Request{
				Header: map[string][]string{},
			},
			vars: map[string]string{"router_id": "1", "alert_id": "1"},
			want: BadRequest("missing User-Email in header", ""),
		},
		"failure | update alert": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{"router_id": "2", "alert_id": "1"},
			body: body,
			want: InternalServerError("unable to update alert", "Test alert error"),
		},
		"success": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{"router_id": "2", "alert_id": "2"},
			body: body,
			want: Ok(alert2),
		},
	}

	// Create test controller
	controller := &AlertsController{
		&baseController{
			&AppContext{
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
	mockMLPService := &mocks.MLPService{}
	mockMLPService.On("GetProject", 1).Return(&mlp.Project{}, nil)

	mockRouterService := &mocks.RoutersService{}
	mockRouterService.On("FindByID", uint(1)).Return(&models.Router{Name: "test"}, nil)
	mockRouterService.On("FindByID", uint(2)).Return(&models.Router{Name: "test"}, nil)

	alert1 := &models.Alert{Model: models.Model{ID: 1}}
	alert2 := &models.Alert{Model: models.Model{ID: 2}}
	mockAlertService := &mocks.AlertService{}
	mockAlertService.On("FindByID", uint(1)).Return(alert1, nil)
	mockAlertService.On("FindByID", uint(2)).Return(alert2, nil)
	mockAlertService.On("FindByID", uint(10)).Return(nil, errors.New("Test alert find error"))
	mockAlertService.On("Delete", *alert1, "user@gojek.com").Return(errors.New("Test alert error"))
	mockAlertService.On("Delete", *alert2, "user@gojek.com").Return(nil)

	// Delete tests
	tests := map[string]struct {
		req  *http.Request
		vars map[string]string
		want *Response
	}{
		"failure | alert not found": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{"alert_id": "10"},
			want: NotFound("alert not found", "Test alert find error"),
		},
		"failure | missing email": {
			req: &http.Request{
				Header: map[string][]string{},
			},
			vars: map[string]string{"alert_id": "1"},
			want: BadRequest("missing User-Email in header", ""),
		},
		"failure | delete alert": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{"alert_id": "1"},
			want: InternalServerError("unable to delete alert", "Test alert error"),
		},
		"success": {
			req: &http.Request{
				Header: map[string][]string{"User-Email": {"user@gojek.com"}},
			},
			vars: map[string]string{"alert_id": "2"},
			want: Ok("Alert with id '2' deleted"),
		},
	}

	// Create test controller
	controller := &AlertsController{
		&baseController{
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
		Header: map[string][]string{"User-Email": {"test@abc.com"}},
	}
	testRequestFailure := &http.Request{
		Header: map[string][]string{"User-Email": {"test-email"}},
	}
	testRequestFailureEmpty := &http.Request{
		Header: map[string][]string{},
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
	alertSvc.On("FindByID", uint(1)).Return(nil, errors.New("Test alert error"))
	alertSvc.On("FindByID", uint(2)).Return(&models.Alert{}, nil)

	// Define test cases
	tests := map[string]struct {
		vars     map[string]string
		expected *Response
	}{
		"failure | invalid alert id": {
			vars:     map[string]string{},
			expected: BadRequest("invalid alert id", "key alert_id not found in vars"),
		},
		"failure | alert not found": {
			vars:     map[string]string{"alert_id": "1"},
			expected: NotFound("alert not found", "Test alert error"),
		},
		"success": {
			vars:     map[string]string{"alert_id": "2"},
			expected: nil,
		},
	}

	// Validate
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := &AlertsController{
				&baseController{
					&AppContext{
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
