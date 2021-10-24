package manager

import (
	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator"
	common "github.com/gojek/turing/engines/experiment/common"
)

// BaseStandardExperimentManager provides dummy implementations for the optional
// StandardExperimentManager methods and can be composed into other concrete
// implementations of the interface to provide the base behavior.
type BaseStandardExperimentManager struct {
	validate *validator.Validate
}

func (*BaseStandardExperimentManager) IsCacheEnabled() bool {
	return true
}

func (*BaseStandardExperimentManager) ListClients() ([]Client, error) {
	return []Client{}, nil
}

func (*BaseStandardExperimentManager) ListExperiments() ([]Experiment, error) {
	return []Experiment{}, nil
}

func (*BaseStandardExperimentManager) ListExperimentsForClient(Client) ([]Experiment, error) {
	return []Experiment{}, nil
}

func (*BaseStandardExperimentManager) ListVariablesForClient(Client) ([]Variable, error) {
	return []Variable{}, nil
}

func (*BaseStandardExperimentManager) ListVariablesForExperiments([]Experiment) (map[string][]Variable, error) {
	return make(map[string][]Variable), nil
}

func (em *BaseStandardExperimentManager) ValidateExperimentConfig(engine Engine, cfg TuringExperimentConfig) error {
	if engine.StandardEngineConfig == nil {
		return errors.New("Missing Standard Engine configuration")
	}

	if engine.StandardEngineConfig.ExperimentSelectionEnabled {
		// Check that there is at least 1 experiment
		if len(cfg.Experiments) < 1 {
			return errors.New("Expected at least 1 experiment in the configuration")
		}
		// If Client Selection is enabled, check that the ClientID in each experiment matches the
		// client info passed in
		if engine.StandardEngineConfig.ClientSelectionEnabled {
			for _, e := range cfg.Experiments {
				if e.ClientID != cfg.Client.ID {
					return errors.New("Client information does not match with the experiment")
				}
			}
		}
	}

	return em.validate.StructFilteredCtx(context.Background(), cfg, func(ns []byte) bool {
		// Determine the fields for validation
		validateFields := []string{"TuringExperimentConfig.Variables"}
		if engine.StandardEngineConfig.ClientSelectionEnabled {
			validateFields = append(validateFields, "TuringExperimentConfig.Client")
		}
		if engine.StandardEngineConfig.ExperimentSelectionEnabled {
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

// NewBaseStandardExperimentManager is a constructor for the base experiment manager
func NewBaseStandardExperimentManager() *BaseStandardExperimentManager {
	return &BaseStandardExperimentManager{
		validate: newExperimentConfigValidator(),
	}
}

// newExperimentConfigValidator returns a default validator for the TuringExperimentConfig
func newExperimentConfigValidator() *validator.Validate {
	v := validator.New()
	// Field Source validation for expected values
	_ = v.RegisterValidation("field-src", func(fl validator.FieldLevel) bool {
		stringSrc := fl.Field().String()
		_, err := common.GetFieldSource(stringSrc)
		return err == nil
	})
	return v
}
