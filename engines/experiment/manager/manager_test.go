package manager

import (
	"encoding/json"
	"reflect"
	"testing"
)

type fakeManager struct {
	config json.RawMessage
}

func (fm fakeManager) IsCacheEnabled() bool {
	return false
}

func (fm fakeManager) GetEngineInfo() Engine {
	return Engine{Name: "fake"}
}

func (fm fakeManager) ListClients() ([]Client, error) {
	return []Client{}, nil
}

func (fm fakeManager) ListExperiments() ([]Experiment, error) {
	return []Experiment{}, nil
}

func (fm fakeManager) ListExperimentsForClient(client Client) ([]Experiment, error) {
	return []Experiment{}, nil
}

func (fm fakeManager) ListVariablesForClient(client Client) ([]Variable, error) {
	return []Variable{}, nil
}

func (fm fakeManager) ListVariablesForExperiments(experiments []Experiment) (map[string][]Variable, error) {
	return map[string][]Variable{}, nil
}

func (fm fakeManager) GetExperimentRunnerConfig(config TuringExperimentConfig) (json.RawMessage, error) {
	return nil, nil
}

func (fm fakeManager) ValidateExperimentConfig(engine Engine, config TuringExperimentConfig) error {
	return nil
}

func newFakeManager(config json.RawMessage) (ExperimentManager, error) {
	return fakeManager{config: config}, nil
}

func TestRegisterAndGet(t *testing.T) {
	tests := []struct {
		name            string
		managerName     string
		managerFactory  Factory
		managerConfig   json.RawMessage
		skipRegister    bool
		want            ExperimentManager
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
				err := Register(tt.managerName, tt.managerFactory)

				if (err != nil) != tt.wantRegisterErr {
					t.Errorf("Register() error = %v, wantErr %v", err, tt.wantRegisterErr)
					return
				}
			}

			got, err := Get(tt.managerName, tt.managerConfig)
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
