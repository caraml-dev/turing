package runner_test

import (
	"context"
	"encoding/json"
	v1 "github.com/gojek/turing/engines/experiment/plugin/inproc/runner"
	"net/http"
	"reflect"
	"testing"

	"github.com/gojek/turing/engines/experiment/runner"
	"github.com/gojek/turing/engines/experiment/runner/mocks"
)

type fakeRunner struct {
	config json.RawMessage
}

func newFakeRunner(config json.RawMessage) (runner.ExperimentRunner, error) {
	return fakeRunner{config: config}, nil
}

func (runner fakeRunner) GetTreatmentForRequest(
	context.Context,
	runner.Logger,
	http.Header, []byte) (*runner.Treatment, error) {
	return nil, nil
}

func TestRegisterAndGet(t *testing.T) {
	tests := []struct {
		name            string
		runnerName      string
		runnerFactory   v1.Factory
		runnerConfig    json.RawMessage
		interceptors    []runner.Interceptor
		want            runner.ExperimentRunner
		wantRegisterErr bool
		wantGetErr      bool
		skipRegister    bool
	}{
		{
			name:          "successful registration and retrieval of fakeRunner",
			runnerName:    "fakeRunner",
			runnerFactory: newFakeRunner,
			runnerConfig:  []byte(`{"foo":"bar"}`),
			interceptors:  []runner.Interceptor{&mocks.Interceptor{}},
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
				err := v1.Register(tt.runnerName, tt.runnerFactory)

				if (err != nil) != tt.wantRegisterErr {
					t.Errorf("Register() error = %v, wantErr %v", err, tt.wantRegisterErr)
					return
				}
			}

			got, err := v1.Get(tt.runnerName, tt.runnerConfig)
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
