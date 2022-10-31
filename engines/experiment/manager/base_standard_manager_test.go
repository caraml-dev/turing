package manager_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/experiment/manager"
	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
)

func TestBaseStandardExperimentManagerMethods(t *testing.T) {
	em := &manager.BaseStandardExperimentManager{}
	actual, err := em.IsCacheEnabled()
	assert.NoError(t, err)
	assert.True(t, actual)

	// Get clients
	clients, err := em.ListClients()
	assert.Equal(t, []manager.Client{}, clients)
	assert.NoError(t, err)

	// Get experiments
	experiments, err := em.ListExperiments()
	assert.Equal(t, []manager.Experiment{}, experiments)
	assert.NoError(t, err)
	experiments, err = em.ListExperimentsForClient(manager.Client{})
	assert.Equal(t, []manager.Experiment{}, experiments)
	assert.NoError(t, err)

	// Get variables
	variables, err := em.ListVariablesForClient(manager.Client{})
	assert.Equal(t, []manager.Variable{}, variables)
	assert.NoError(t, err)
	expVars, err := em.ListVariablesForExperiments([]manager.Experiment{})
	assert.Equal(t, make(map[string][]manager.Variable), expVars)
	assert.NoError(t, err)
}

func TestValidateExperimentConfig(t *testing.T) {
	tests := map[string]struct {
		engine manager.Engine
		cfg    manager.TuringExperimentConfig
		err    string
	}{
		"failure | no engine info": {
			err: "Missing Standard Engine configuration",
		},
		"failure | missing client info": {
			engine: manager.Engine{
				StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
					ClientSelectionEnabled:     true,
					ExperimentSelectionEnabled: false,
				},
			},
			cfg: manager.TuringExperimentConfig{},
			err: strings.Join([]string{
				"Key: 'TuringExperimentConfig.Client.ID' Error:",
				"Field validation for 'ID' failed on the 'required' tag\n",
				"Key: 'TuringExperimentConfig.Client.Username' Error:",
				"Field validation for 'Username' failed on the 'required' tag",
			}, ""),
		},
		"failure | no experiment": {
			engine: manager.Engine{
				StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
					ClientSelectionEnabled:     false,
					ExperimentSelectionEnabled: true,
				},
			},
			cfg: manager.TuringExperimentConfig{},
			err: "Expected at least 1 experiment in the configuration",
		},
		"failure | client ID mismatch": {
			engine: manager.Engine{
				StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
					ClientSelectionEnabled:     true,
					ExperimentSelectionEnabled: true,
				},
			},
			cfg: manager.TuringExperimentConfig{
				Client: manager.Client{
					ID:       "1",
					Username: "client-a",
				},
				Experiments: []manager.Experiment{
					{
						ID:       "1",
						Name:     "test-exp",
						ClientID: "2",
					},
				},
			},
			err: "Client information does not match with the experiment",
		},
		"failure | missing experiment info": {
			engine: manager.Engine{
				StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
					ClientSelectionEnabled:     false,
					ExperimentSelectionEnabled: true,
				},
			},
			cfg: manager.TuringExperimentConfig{
				Experiments: []manager.Experiment{
					{
						ID: "1",
					},
				},
			},
			err: strings.Join([]string{
				"Key: 'TuringExperimentConfig.Experiments[0].Name' Error:",
				"Field validation for 'Name' failed on the 'required' tag",
			}, ""),
		},
		"failure | required variable not configured": {
			engine: manager.Engine{
				StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
					ClientSelectionEnabled:     false,
					ExperimentSelectionEnabled: false,
				},
			},
			cfg: manager.TuringExperimentConfig{
				Variables: manager.Variables{
					Config: []manager.VariableConfig{
						{
							Name:        "a",
							FieldSource: request.HeaderFieldSource,
						},
						{
							Name:        "b",
							Required:    true,
							FieldSource: request.HeaderFieldSource,
						},
					},
				},
			},
			err: strings.Join([]string{
				"Key: 'TuringExperimentConfig.Variables.Config[1].Field' Error:",
				"Field validation for 'Field' failed on the 'required_with' tag",
			}, ""),
		},
		"failure | bad field source": {
			engine: manager.Engine{
				StandardExperimentManagerConfig: &manager.StandardExperimentManagerConfig{
					ClientSelectionEnabled:     false,
					ExperimentSelectionEnabled: false,
				},
			},
			cfg: manager.TuringExperimentConfig{
				Variables: manager.Variables{
					Config: []manager.VariableConfig{
						{
							Name:        "a",
							FieldSource: request.FieldSource("unknown"),
						},
					},
				},
			},
			err: strings.Join([]string{
				"Key: 'TuringExperimentConfig.Variables.Config[0].FieldSource' Error:",
				"Field validation for 'FieldSource' failed on the 'field-src' tag",
			}, ""),
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			em := manager.NewBaseStandardExperimentManager(data.engine)

			cfg, _ := json.Marshal(data.cfg)
			err := em.ValidateExperimentConfig(cfg)
			if data.err != "" {
				// Expect error
				assert.Error(t, err)
				if err != nil {
					assert.Equal(t, data.err, err.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
