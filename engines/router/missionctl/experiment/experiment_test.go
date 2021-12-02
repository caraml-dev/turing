package experiment

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	_ "github.com/gojek/turing/engines/experiment/v1/runner/nop"
	"github.com/gojek/turing/engines/router/missionctl/turingctx"
	"github.com/stretchr/testify/assert"
)

// TestTreatment satisfies the Treatment interface
type TestTreatment struct {
	name      string
	treatment string
	raw       json.RawMessage // raw experiment config from the experiment engine
}

// GetExperimentName returns the name of the experiment
func (ex TestTreatment) GetExperimentName() string {
	return ex.name
}

// GetName retrives the treatment (or control) name
func (ex TestTreatment) GetName() string {
	return ex.treatment
}

// GetConfig returs the raw experiment config from the experiment engine
func (ex TestTreatment) GetConfig() json.RawMessage {
	return ex.raw
}

type testSuiteExperimentResponse struct {
	treatment      TestTreatment
	err            error
	expectedConfig json.RawMessage
	expectedError  string
}

func TestNewExperimentRunner(t *testing.T) {
	tests := map[string]struct {
		engineName               string
		config                   json.RawMessage
		litmusEnvValueForPassKey string // empty env value means the environment variable is unset
		xpEnvValueForPasskey     string // empty env value means the environment variable is unset
		wantErr                  bool
	}{
		"nop | success": {
			engineName: "nop",
		},
		"unsupported | false": {
			engineName: "unsupported",
			wantErr:    true,
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewExperimentRunner(data.engineName, data.config)
			if (err != nil) != data.wantErr {
				t.Errorf("NewExperimentRunner() error= %v, wantErr %v", err, data.wantErr)
			}
		})
	}
}

func TestNewResponse(t *testing.T) {
	tests := map[string]testSuiteExperimentResponse{
		"success": {
			treatment: TestTreatment{
				raw: []byte(`{"key": "value"}`),
			},
			err:            nil,
			expectedConfig: []byte(`{"key": "value"}`),
			expectedError:  "",
		},
		"failure_error": {
			treatment: TestTreatment{
				raw: []byte(`{"key": "value"}`),
			},
			err:            fmt.Errorf("Test Error"),
			expectedConfig: nil,
			expectedError:  "Test Error",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp := NewResponse(data.treatment, data.err)

			// Validate
			assert.Equal(t, data.expectedConfig, resp.Configuration)
			assert.Equal(t, []byte(data.expectedConfig), resp.Body())
			assert.Equal(t, data.expectedError, resp.Error)
			assert.Nil(t, resp.Header())
		})
	}
}

func TestWithExperimentResponseChannel(t *testing.T) {
	ch := make(chan *Response)
	ctx := WithExperimentResponseChannel(context.Background(), ch)
	respCh, err := GetExperimentResponseChannel(ctx)

	// Validate
	assert.NotNil(t, ctx.Value(turingctx.TuringTreatmentChannelKey))
	assert.Equal(t, ch, respCh)
	assert.Nil(t, err)
}

func TestGetExperimentResponseChannelError(t *testing.T) {
	respCh, err := GetExperimentResponseChannel(context.Background())
	assert.Nil(t, respCh)
	assert.NotNil(t, err)
}
