package runner

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/gojek/turing/engines/experiment/runner/mocks"
)

type fakeRunner struct {
	config json.RawMessage
}

func newFakeRunner(config json.RawMessage) (ExperimentRunner, error) {
	return fakeRunner{config: config}, nil
}

func (runner fakeRunner) GetTreatmentForRequest(context.Context, Logger, http.Header, []byte) (Treatment, error) {
	return nil, nil
}

func TestRegisterAndGet(t *testing.T) {
	tests := []struct {
		name            string
		runnerName      string
		runnerFactory   Factory
		runnerConfig    json.RawMessage
		interceptors    []Interceptor
		want            ExperimentRunner
		wantRegisterErr bool
		wantGetErr      bool
		skipRegister    bool
	}{
		{
			name:          "successful registration and retrieval of fakeRunner",
			runnerName:    "fakeRunner",
			runnerFactory: newFakeRunner,
			runnerConfig:  []byte(`{"foo":"bar"}`),
			interceptors:  []Interceptor{&mocks.Interceptor{}},
			want:          fakeRunner{config: []byte(`{"foo":"bar"}`)},
		},
		{
			name:            "failed multiple registrations of fakeRunner",
			runnerName:      "fakeRunner",
			runnerFactory:   newFakeRunner,
			runnerConfig:    []byte(`{"foo":"bar"}`),
			want:            fakeRunner{config: []byte(`{"foo":"bar"}`)},
			wantRegisterErr: true,
		},
		{
			name:         "failed retrieval of non-registered runner",
			runnerName:   "nonRegisteredRunner",
			wantGetErr:   true,
			skipRegister: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.skipRegister {
				err := Register(tt.runnerName, tt.runnerFactory)

				if (err != nil) != tt.wantRegisterErr {
					t.Errorf("Register() error = %v, wantErr %v", err, tt.wantRegisterErr)
					return
				}
			}

			got, err := Get(tt.runnerName, tt.runnerConfig)
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
