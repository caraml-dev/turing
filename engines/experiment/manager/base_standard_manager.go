package manager

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/go-playground/validator"

	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
)

// BaseStandardExperimentManager provides dummy implementations for the optional
// StandardExperimentManager methods and can be composed into other concrete
// implementations of the interface to provide the base behavior.
type BaseStandardExperimentManager struct {
	validate *validator.Validate
	info     Engine
}

func (em *BaseStandardExperimentManager) GetEngineInfo() (Engine, error) {
	return em.info, nil
}

func (*BaseStandardExperimentManager) IsCacheEnabled() (bool, error) {
	return true, nil
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

func (em *BaseStandardExperimentManager) ValidateExperimentConfig(cfg json.RawMessage) error {
	engineCfg := em.info.StandardExperimentManagerConfig
	if engineCfg == nil {
		return errors.New("Missing Standard Engine configuration")
	}

	var experimentCfg TuringExperimentConfig
	if err := json.Unmarshal(cfg, &experimentCfg); err != nil {
		return err
	}

	if engineCfg.ExperimentSelectionEnabled {
		// Check that there is at least 1 experiment
		if len(experimentCfg.Experiments) < 1 {
			return errors.New("Expected at least 1 experiment in the configuration")
		}
		// If Client Selection is enabled, check that the ClientID in each experiment matches the
		// client info passed in
		if engineCfg.ClientSelectionEnabled {
			for _, e := range experimentCfg.Experiments {
				if e.ClientID != experimentCfg.Client.ID {
					return errors.New("Client information does not match with the experiment")
				}
			}
		}
	}

	return em.validate.StructFilteredCtx(context.Background(), experimentCfg, func(ns []byte) bool {
		// Determine the fields for validation
		validateFields := []string{"TuringExperimentConfig.Variables"}
		if engineCfg.ClientSelectionEnabled {
			validateFields = append(validateFields, "TuringExperimentConfig.Client")
		}
		if engineCfg.ExperimentSelectionEnabled {
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
func NewBaseStandardExperimentManager(info Engine) *BaseStandardExperimentManager {
	return &BaseStandardExperimentManager{
		info:     info,
		validate: newExperimentConfigValidator(),
	}
}

// newExperimentConfigValidator returns a default validator for the TuringExperimentConfig
func newExperimentConfigValidator() *validator.Validate {
	v := validator.New()
	// Field Source validation for expected values
	_ = v.RegisterValidation("field-src", func(fl validator.FieldLevel) bool {
		stringSrc := fl.Field().String()
		_, err := request.GetFieldSource(stringSrc)
		return err == nil
	})
	return v
}
