package manager

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	common "github.com/gojek/turing/engines/experiment/common"
	"github.com/stretchr/testify/assert"
)

func TestNewBaseExperimentManager(t *testing.T) {
	em := &BaseExperimentManager{}
	// Check that BaseExperimentManager implements ExperimentManager
	_, ok := interface{}(em).(ExperimentManager)
	assert.True(t, ok)
}

func TestBaseExperimentManagerMethods(t *testing.T) {
	em := &BaseExperimentManager{}
	assert.Equal(t, true, em.IsCacheEnabled())
	assert.Equal(t, Engine{}, em.GetEngineInfo())

	// Get clients
	clients, err := em.ListClients()
	assert.Equal(t, []Client{}, clients)
	assert.NoError(t, err)

	// Get experiments
	experiments, err := em.ListExperiments()
	assert.Equal(t, []Experiment{}, experiments)
	assert.NoError(t, err)
	experiments, err = em.ListExperimentsForClient(Client{})
	assert.Equal(t, []Experiment{}, experiments)
	assert.NoError(t, err)

	// Get variables
	variables, err := em.ListVariablesForClient(Client{})
	assert.Equal(t, []Variable{}, variables)
	assert.NoError(t, err)
	expVars, err := em.ListVariablesForExperiments([]Experiment{})
	assert.Equal(t, make(map[string][]Variable), expVars)
	assert.NoError(t, err)

	// Get Fiber config
	rawMessage, err := em.GetExperimentRunnerConfig(TuringExperimentConfig{})
	assert.Equal(t, json.RawMessage{}, rawMessage)
	assert.NoError(t, err)
}

func TestValidateExperimentConfig(t *testing.T) {
	em := NewBaseExperimentManager()

	// Define tests
	tests := map[string]struct {
		engine Engine
		cfg    TuringExperimentConfig
		err    string
	}{
		"failure | missing client info": {
			engine: Engine{
				ClientSelectionEnabled:     true,
				ExperimentSelectionEnabled: false,
			},
			cfg: TuringExperimentConfig{},
			err: strings.Join([]string{
				"Key: 'TuringExperimentConfig.Client.ID' Error:",
				"Field validation for 'ID' failed on the 'required' tag\n",
				"Key: 'TuringExperimentConfig.Client.Username' Error:",
				"Field validation for 'Username' failed on the 'required' tag",
			}, ""),
		},
		"failure | no experiment": {
			engine: Engine{
				ClientSelectionEnabled:     false,
				ExperimentSelectionEnabled: true,
			},
			cfg: TuringExperimentConfig{},
			err: strings.Join([]string{
				"Key: 'TuringExperimentConfig.experiments' Error:",
				"Field validation for 'experiments' failed on the 'no-experiment-selected' tag",
			}, ""),
		},
		"failure | client ID mismatch": {
			engine: Engine{
				ClientSelectionEnabled:     true,
				ExperimentSelectionEnabled: true,
			},
			cfg: TuringExperimentConfig{
				Client: Client{
					ID:       "1",
					Username: "client-a",
				},
				Experiments: []Experiment{
					{
						ID:       "1",
						Name:     "test-exp",
						ClientID: "2",
					},
				},
			},
			err: strings.Join([]string{
				"Key: 'TuringExperimentConfig.experiments' Error:",
				"Field validation for 'experiments' failed on the 'client-id-mismatch' tag",
			}, ""),
		},
		"failure | missing experiment info": {
			engine: Engine{
				ClientSelectionEnabled:     false,
				ExperimentSelectionEnabled: true,
			},
			cfg: TuringExperimentConfig{
				Experiments: []Experiment{
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
			engine: Engine{
				ClientSelectionEnabled:     false,
				ExperimentSelectionEnabled: false,
			},
			cfg: TuringExperimentConfig{
				Variables: Variables{
					Config: []VariableConfig{
						{
							Name:        "a",
							FieldSource: common.HeaderFieldSource,
						},
						{
							Name:        "b",
							Required:    true,
							FieldSource: common.HeaderFieldSource,
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
			engine: Engine{
				ClientSelectionEnabled:     false,
				ExperimentSelectionEnabled: false,
			},
			cfg: TuringExperimentConfig{
				Variables: Variables{
					Config: []VariableConfig{
						{
							Name:        "a",
							FieldSource: common.FieldSource("unknown"),
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
			err := em.ValidateExperimentConfig(data.engine, data.cfg)
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

func TestValidateTuringExperimentConfigMissingContextKey(t *testing.T) {
	em := NewBaseExperimentManager()
	err := em.validate.StructFilteredCtx(context.Background(), TuringExperimentConfig{}, func(ns []byte) bool {
		return true
	})
	assert.Error(t, err)
	if err != nil {
		assert.Equal(t, strings.Join([]string{
			"Key: 'TuringExperimentConfig.experiment_config' Error:",
			"Field validation for 'experiment_config' failed on the 'missing-context-info' tag",
		}, ""), err.Error())
	}
}
