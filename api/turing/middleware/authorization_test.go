package middleware

import (
	"testing"

	"github.com/gojek/mlp/api/pkg/authz/enforcer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/turing/api/turing/middleware/mocks"
)

func TestBootstrapTuringPolicies(t *testing.T) {
	authzEnforcer := &mocks.Enforcer{}
	authzEnforcer.On("UpsertPolicy", mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Return(nil, nil)

	// Create new authorizer
	_, err := NewAuthorizer(authzEnforcer, "/")

	// Validate
	assert.NoError(t, err)
	authzEnforcer.AssertCalled(t, "UpsertPolicy", "allow-all-list-experiment-engines",
		[]string{},
		[]string{"**"},
		[]string{"experiment-engines", "experiment-engines:**"},
		[]string{enforcer.ActionRead},
	)
}
