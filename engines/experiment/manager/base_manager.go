package manager

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-playground/validator"
)

// BaseExperimentManager provides dummy implementations for the ExperimentManager methods
// and can be composed into other concrete implementations of the interface to provide the
// base behavior.
type BaseExperimentManager struct {
	validate *validator.Validate
}

func (*BaseExperimentManager) IsCacheEnabled() bool {
	return true
}

func (*BaseExperimentManager) GetEngineInfo() Engine {
	return Engine{}
}

func (*BaseExperimentManager) ListClients() ([]Client, error) {
	return []Client{}, nil
}

func (*BaseExperimentManager) ListExperiments() ([]Experiment, error) {
	return []Experiment{}, nil
}

func (*BaseExperimentManager) ListExperimentsForClient(Client) ([]Experiment, error) {
	return []Experiment{}, nil
}

func (*BaseExperimentManager) ListVariablesForClient(Client) ([]Variable, error) {
	return []Variable{}, nil
}

func (*BaseExperimentManager) ListVariablesForExperiments([]Experiment) (map[string][]Variable, error) {
	return make(map[string][]Variable), nil
}

func (*BaseExperimentManager) GetExperimentRunnerConfig(TuringExperimentConfig) (json.RawMessage, error) {
	return json.RawMessage{}, nil
}

func (em *BaseExperimentManager) ValidateExperimentConfig(engine Engine, cfg TuringExperimentConfig) error {
	ctx := context.WithValue(context.Background(), engineValidatorCtxKey, engine)

	return em.validate.StructFilteredCtx(ctx, cfg, func(ns []byte) bool {
		// Determine the fields for validation
		validateFields := []string{"TuringExperimentConfig.Variables"}
		if engine.ClientSelectionEnabled {
			validateFields = append(validateFields, "TuringExperimentConfig.Client")
		}
		if engine.ExperimentSelectionEnabled {
			validateFields = append(validateFields, "TuringExperimentConfig.Experiments")
		}

		// If the field's fully qualified name starts with the name of any of the chosen fields,
		// do not filter it from validation (return false will pick up the field for validation).
		for _, field := range validateFields {
			if strings.HasPrefix(string(ns), field) {
				return false
			}
		}
		return true
	})
}

// NewBaseExperimentManager is a constructor for the base experiment manager
func NewBaseExperimentManager() *BaseExperimentManager {
	return &BaseExperimentManager{
		validate: newExperimentConfigValidator(),
	}
}
