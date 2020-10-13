package mocks

import (
	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/client"
	"github.com/stretchr/testify/mock"
)

// MLPService implements the MLPService interface
type MLPService struct {
	mock.Mock
}

// GetEnvironments satisfies the MLPService interface
func (m *MLPService) GetEnvironments() ([]merlin.Environment, error) {
	ret := m.Called()

	if ret[1] != nil {
		return nil, ret[1].(error)
	}

	return (ret[0]).([]merlin.Environment), nil
}

// GetEnvironment satisfies the MLPService interface
func (m *MLPService) GetEnvironment(name string) (*merlin.Environment, error) {
	ret := m.Called(name)

	if ret[1] != nil {
		return nil, ret[1].(error)
	}

	return (ret[0]).(*merlin.Environment), nil
}

// GetProject satisfies the MLPService interface
func (m *MLPService) GetProject(id int) (*mlp.Project, error) {
	ret := m.Called(id)

	if ret[1] != nil {
		return nil, ret[1].(error)
	}

	return (ret[0]).(*mlp.Project), nil
}

// GetSecret satisfies the MLPService interface
func (m *MLPService) GetSecret(projectID int, name string) (string, error) {
	ret := m.Called(projectID, name)

	if ret[1] != nil {
		return "", ret[1].(error)
	}

	return (ret[0]).(string), nil
}
