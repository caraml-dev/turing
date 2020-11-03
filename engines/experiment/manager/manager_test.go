package manager_test

import (
	"encoding/json"
	"reflect"
	"testing"

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
