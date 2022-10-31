package experiment

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	_ "github.com/caraml-dev/turing/engines/experiment/plugin/inproc/runner/nop"
	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/caraml-dev/turing/engines/router/missionctl/turingctx"
)

type testSuiteExperimentResponse struct {
	treatment      runner.Treatment
	err            error
	expectedConfig json.RawMessage
	expectedError  string
}

func TestNewExperimentRunner(t *testing.T) {
	tests := map[string]struct {
		engineName string
		config     map[string]interface{}
		wantErr    bool
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
			treatment: runner.Treatment{
				Config: []byte(`{"key": "value"}`),
			},
			err:            nil,
			expectedConfig: []byte(`{"key": "value"}`),
			expectedError:  "",
		},
		"failure_error": {
			treatment: runner.Treatment{
				Config: []byte(`{"key": "value"}`),
			},
			err:            fmt.Errorf("Test Error"),
			expectedConfig: nil,
			expectedError:  "Test Error",
		},
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			resp := NewResponse(&data.treatment, data.err)

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
