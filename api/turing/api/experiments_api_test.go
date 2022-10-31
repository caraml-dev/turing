package api

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/api/turing/service/mocks"
	"github.com/caraml-dev/turing/engines/experiment/manager"
	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
)

func TestListExperimentEngines(t *testing.T) {
	// Create mock experiment service
	engines := []manager.Engine{
		{
			Name: "test-engine",
			StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
				ClientSelectionEnabled:     true,
				ExperimentSelectionEnabled: true,
			},
		},
	}
	svc := &mocks.ExperimentsService{}
	svc.On("ListEngines").Return(engines, nil)

	// Create controller
	ctrl := ExperimentsController{
		BaseController{
			AppContext: &AppContext{
				ExperimentsService: svc,
			},
		},
	}

	// Test
	expectedResponse := &Response{
		code: 200,
		data: engines,
	}
	assert.Equal(t, expectedResponse, ctrl.ListExperimentEngines(nil, nil, nil))
}

func TestListExperimentEngineClients(t *testing.T) {
	// Create mock experiment service
	clients := []manager.Client{
		{
			ID:       "1",
			Username: "client1",
		},
		{
			ID:       "2",
			Username: "client2",
		},
	}
	successSvc := &mocks.ExperimentsService{}
	successSvc.On("ListClients", "test-engine").Return(clients, nil)
	failureSvc := &mocks.ExperimentsService{}
	failureSvc.On("ListClients", "test-engine").Return([]manager.Client{}, errors.New("Test error"))

	// Define tests
	tests := map[string]struct {
		ctrl     ExperimentsController
		vars     RequestVars
		expected *Response
	}{
		"failure | bad input": {
			ctrl: ExperimentsController{
				BaseController{
					AppContext: &AppContext{},
				},
			},
			vars:     RequestVars{},
			expected: BadRequest("invalid experiment engine", "key engine not found in vars"),
		},
		"failure | bad response": {
			ctrl: ExperimentsController{
				BaseController{
					AppContext: &AppContext{
						ExperimentsService: failureSvc,
					},
				},
			},
			vars:     RequestVars{"engine": {"test-engine"}},
			expected: InternalServerError("error when querying test-engine clients", "Test error"),
		},
		"success": {
			ctrl: ExperimentsController{
				BaseController{
					AppContext: &AppContext{
						ExperimentsService: successSvc,
					},
				},
			},
			vars: RequestVars{"engine": {"test-engine"}},
			expected: &Response{
				code: 200,
				data: clients,
			},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			response := data.ctrl.ListExperimentEngineClients(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestListExperimentEngineExperiments(t *testing.T) {
	// Create mock experiment services
	experiments := []manager.Experiment{
		{
			ID:       "2",
			Name:     "name",
			ClientID: "1",
			Variants: []manager.Variant{
				{
					Name: "control",
				},
				{
					Name: "treament",
				},
			},
		},
	}
	successSvc := &mocks.ExperimentsService{}
	successSvc.On("ListExperiments", "test-engine", "1").Return(experiments, nil)
	failureSvc := &mocks.ExperimentsService{}
	failureSvc.On("ListExperiments", "test-engine", "2").Return([]manager.Experiment{}, errors.New("Test error"))

	// Define tests
	tests := map[string]struct {
		ctrl     ExperimentsController
		vars     RequestVars
		expected *Response
	}{
		"failure | bad input": {
			ctrl: ExperimentsController{
				BaseController{
					AppContext: &AppContext{},
				},
			},
			vars:     RequestVars{},
			expected: BadRequest("invalid experiment engine", "key engine not found in vars"),
		},
		"failure | bad response": {
			ctrl: ExperimentsController{
				BaseController{
					AppContext: &AppContext{
						ExperimentsService: failureSvc,
					},
				},
			},
			vars: RequestVars{
				"engine":    {"test-engine"},
				"client_id": {"2"},
			},
			expected: InternalServerError("error when querying test-engine experiments", "Test error"),
		},
		"success": {
			ctrl: ExperimentsController{
				BaseController{
					AppContext: &AppContext{
						ExperimentsService: successSvc,
					},
				},
			},
			vars: RequestVars{
				"engine":    {"test-engine"},
				"client_id": {"1"},
			},
			expected: &Response{
				code: 200,
				data: experiments,
			},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			response := data.ctrl.ListExperimentEngineExperiments(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}

func TestListExperimentEngineVariables(t *testing.T) {
	// Create mock experiment service
	variables := manager.Variables{
		ClientVariables: []manager.Variable{
			{
				Name:     "var-1",
				Required: true,
			},
		},
		ExperimentVariables: map[string][]manager.Variable{},
		Config: []manager.VariableConfig{
			{
				Name:        "var-1",
				Required:    true,
				FieldSource: request.HeaderFieldSource,
			},
		},
	}
	successSvc := &mocks.ExperimentsService{}
	successSvc.On("ListVariables", "test-engine", "1", []string{"1", "2"}).
		Return(variables, nil)
	failureSvc := &mocks.ExperimentsService{}
	failureSvc.On("ListVariables", "test-engine", "2", []string(nil)).
		Return(manager.Variables{}, errors.New("Test error"))

	// Define tests
	tests := map[string]struct {
		ctrl     ExperimentsController
		vars     RequestVars
		expected *Response
	}{
		"failure | bad input": {
			ctrl: ExperimentsController{
				BaseController{
					AppContext: &AppContext{},
				},
			},
			vars:     RequestVars{},
			expected: BadRequest("invalid experiment engine", "key engine not found in vars"),
		},
		"failure | bad response": {
			ctrl: ExperimentsController{
				BaseController{
					AppContext: &AppContext{
						ExperimentsService: failureSvc,
					},
				},
			},
			vars: RequestVars{
				"engine":    {"test-engine"},
				"client_id": {"2"},
			},
			expected: InternalServerError("error when querying test-engine variables", "Test error"),
		},
		"success": {
			ctrl: ExperimentsController{
				BaseController{
					AppContext: &AppContext{
						ExperimentsService: successSvc,
					},
				},
			},
			vars: RequestVars{
				"engine":        {"test-engine"},
				"client_id":     {"1"},
				"experiment_id": {"1,2"},
			},
			expected: &Response{
				code: 200,
				data: variables,
			},
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			response := data.ctrl.ListExperimentEngineVariables(nil, data.vars, nil)
			assert.Equal(t, data.expected, response)
		})
	}
}
