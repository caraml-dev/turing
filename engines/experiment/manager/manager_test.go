package manager_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/manager/mocks"
)

type fakeManager struct {
	*mocks.ExperimentManager
	config json.RawMessage
}

func newFakeManager(config json.RawMessage) (manager.ExperimentManager, error) {
	return fakeManager{config: config}, nil
}

func TestRegisterAndGet(t *testing.T) {
	tests := []struct {
		name            string
		managerName     string
		managerFactory  manager.Factory
		managerConfig   json.RawMessage
		skipRegister    bool
		want            manager.ExperimentManager
		wantRegisterErr bool
		wantGetErr      bool
	}{
		{
			name:           "successful registration and retrieval of fakeManager",
			managerName:    "fakeManager",
			managerFactory: newFakeManager,
			managerConfig:  []byte(`{"foo":"bar"}`),
			want:           fakeManager{config: []byte(`{"foo":"bar"}`)},
		},
		{
			name:            "failed multiple registrations of fakeManager",
			managerName:     "fakeManager",
			managerFactory:  newFakeManager,
			managerConfig:   []byte(`{"foo":"bar"}`),
			want:            fakeManager{config: []byte(`{"foo":"bar"}`)},
			wantRegisterErr: true,
		},
		{
			name:         "failed retrieval of non-registered manager",
			managerName:  "nonRegisteredManager",
			skipRegister: true,
			wantGetErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.skipRegister {
				err := manager.Register(tt.managerName, tt.managerFactory)

				if (err != nil) != tt.wantRegisterErr {
					t.Errorf("Register() error = %v, wantErr %v", err, tt.wantRegisterErr)
					return
				}
			}

			got, err := manager.Get(tt.managerName, tt.managerConfig)
			if (err != nil) != tt.wantGetErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantGetErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStandardExperimentConfig(t *testing.T) {
	tests := map[string]struct {
		cfg      interface{}
		expected manager.TuringExperimentConfig
		err      string
	}{
		"failure | invalid json": {
			cfg: func() {},
			err: "json: unsupported type: func()",
		},
		"failure | invalid standard config": {
			cfg: []int{1, 2},
			err: "json: cannot unmarshal array into Go value of type manager.TuringExperimentConfig",
		},
		"success": {
			cfg: struct {
				Client      manager.Client       `json:"client,omitempty"`
				Experiments []manager.Experiment `json:"experiments,omitempty"`
				Variables   manager.Variables    `json:"variables,omitempty"`
				Extra       int                  `json:"extra_value"`
			}{},
			expected: manager.TuringExperimentConfig{},
		},
	}

	// Run tests
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			stdCfg, err := manager.GetStandardExperimentConfig(data.cfg)
			if data.err != "" {
				assert.EqualError(t, err, data.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, data.expected, stdCfg)
			}
		})
	}
}
