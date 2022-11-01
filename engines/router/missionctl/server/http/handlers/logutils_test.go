package handlers

import (
	"context"
	"fmt"
	"testing"

	fiberProtocol "github.com/gojek/fiber/protocol"
	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/router/missionctl/errors"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
)

// TestCopyResponseToLogChannel tests the copyResponseToLogChannel method in logutils.
// Verify that when an error is set, the error message is copied, response is set to null;
// when the error is empty, the response is copied an error field in the log is empty.
// Additionally, verify that the response body is still open for reading after the operation.
func TestCopyResponseToLogChannel(t *testing.T) {
	tests := map[string]*errors.TuringError{
		"success": nil,
		"error":   errors.NewTuringError(fmt.Errorf("test error"), fiberProtocol.HTTP),
	}

	for name, httpErr := range tests {
		t.Run(name, func(t *testing.T) {
			// Make test response
			resp := tu.MakeTestMisisonControlResponse()
			// Capture expected body, for validation
			expectedRespBody := resp.Body()

			// Make response channel
			respCh := make(chan routerResponse, 1)

			// Push message to channel and close the channel
			copyResponseToLogChannel(context.Background(), respCh, "test", resp, httpErr)
			close(respCh)

			// Read from the channel and validate
			data := <-respCh
			// Check key
			assert.Equal(t, data.key, "test")
			// Check error and response
			if httpErr == nil {
				assert.Empty(t, data.err)
				assert.Equal(t, expectedRespBody, data.body)
			} else {
				assert.Equal(t, data.err, httpErr.Error())
				assert.Empty(t, data.body)
			}
			// Check that the original response body is still readable
			assert.Equal(t, expectedRespBody, resp.Body())
		})
	}
}
