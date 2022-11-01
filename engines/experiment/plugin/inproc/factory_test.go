package inproc_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/caraml-dev/turing/engines/experiment/config"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/experiment/manager"
	mocksManager "github.com/caraml-dev/turing/engines/experiment/manager/mocks"
	plugin "github.com/caraml-dev/turing/engines/experiment/plugin/inproc"
	managerPlugin "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/manager"
	runnerPlugin "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner"
	"github.com/caraml-dev/turing/engines/experiment/runner"
	mocksRunner "github.com/caraml-dev/turing/engines/experiment/runner/mocks"
)

func Test_NewEngineFactory(t *testing.T) {
	suite := map[string]struct {
		engine string
		cfg    config.EngineConfig
	}{
		"success": {
			engine: "engine-1",
			cfg: config.EngineConfig{
				EngineConfiguration: map[string]interface{}{
					"my_config": "my_value",
				},
			},
		},
	}
	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			factory, err := plugin.NewEngineFactory(tt.engine, tt.cfg)
			assert.NoError(t, err)
			assert.Equal(t, tt.engine, factory.EngineName)

			rawConfig, _ := tt.cfg.RawEngineConfig()
			assert.Equal(t, rawConfig, factory.EngineConfig)
		})
	}
}

func withPatchedManagerRegistry(em manager.ExperimentManager, err string, fn func()) {
	monkey.Patch(managerPlugin.Get,
		func(name string, config json.RawMessage) (manager.ExperimentManager, error) {
			if err != "" {
				return em, errors.New(err)
			}
			return em, nil
		},
	)
	defer monkey.Unpatch(managerPlugin.Get)
	fn()
}

func TestEngineFactory_GetExperimentManager(t *testing.T) {
	suite := map[string]struct {
		engine   string
		expected manager.ExperimentManager
		err      string
	}{
		"success | experiment manager exists in the registry": {
			engine:   "engine-1",
			expected: &mocksManager.ExperimentManager{},
		},
		"failure | experiment manager doesn't exists in the registry": {
			engine: "engine-1",
			err:    "no experiment manager found for name \"engine-1\"",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			factory, _ := plugin.NewEngineFactory(tt.engine, config.EngineConfig{})
			withPatchedManagerRegistry(tt.expected, tt.err, func() {
				actual, err := factory.GetExperimentManager()
				if tt.err != "" {
					assert.EqualError(t, err, tt.err)
				} else {
					assert.NoError(t, err)
					assert.Same(t, tt.expected, actual)
				}
			})
		})
	}
}

func withPatchedRunnerRegistry(er runner.ExperimentRunner, err string, fn func()) {
	monkey.Patch(runnerPlugin.Get,
		func(name string, config json.RawMessage) (runner.ExperimentRunner, error) {
			if err != "" {
				return er, errors.New(err)
			}
			return er, nil
		},
	)
	defer monkey.Unpatch(runnerPlugin.Get)
	fn()
}

func TestEngineFactory_GetExperimentRunner(t *testing.T) {
	suite := map[string]struct {
		engine   string
		expected runner.ExperimentRunner
		err      string
	}{
		"success | experiment runner exists in the registry": {
			engine:   "engine-1",
			expected: &mocksRunner.ExperimentRunner{},
		},
		"failure | experiment runner doesn't exists in the registry": {
			engine: "engine-1",
			err:    "no experiment runner found for name \"engine-1\"",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			factory, _ := plugin.NewEngineFactory(tt.engine, config.EngineConfig{})
			withPatchedRunnerRegistry(tt.expected, tt.err, func() {
				actual, err := factory.GetExperimentRunner()
				if tt.err != "" {
					assert.EqualError(t, err, tt.err)
				} else {
					assert.NoError(t, err)
					assert.Same(t, tt.expected, actual)
				}
			})
		})
	}
}
