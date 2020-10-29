package manager

import (
	"context"

	"github.com/go-playground/validator"
	common "github.com/gojek/turing/engines/experiment/common"
)

// Engine describes the properties of an experiment engine
type Engine struct {
	Name                       string `json:"name"`
	ClientSelectionEnabled     bool   `json:"client_selection_enabled"`
	ExperimentSelectionEnabled bool   `json:"experiment_selection_enabled"`
	HomePageURL                string `json:"home_page_url"`
}

// Client describes the properties of a client registered on an experiment engine
type Client struct {
	ID       string `json:"id" validate:"required"`
	Username string `json:"username" validate:"required"`
	Passkey  string `json:"passkey,omitempty"`
}

// Variant describes the properties of a variant registered on an experiment
type Variant struct {
	Name string `json:"name"`
}

// Experiment describes the properties of an experiment registered on an experiment engine
type Experiment struct {
	ID       string    `json:"id" validate:"required"`
	Name     string    `json:"name" validate:"required"`
	ClientID string    `json:"client_id"`
	Variants []Variant `json:"variants"`
}

// VariableType is used to describe experiment varibles
type VariableType string

const (
	// UnsupportedVariableType variable type is used when the experiment engine does not classify
	// the variable (None) or the type is not known or relevant to the API
	UnsupportedVariableType VariableType = "unsupported"
	// UnitVariableType variable type is used for variables that as as experiment units (typically,
	// their hash is used to randomly assign variants / parameter values)
	UnitVariableType VariableType = "unit"
	// FilterVariableType variable type is used for variables that act as filters (i.e., if the
	// incoming value for the variable in a request does nt fit certain criteria, the experiment is
	// considered to be not applicable)
	FilterVariableType VariableType = "filter"
)

// Variable descibes the properties of a variable registered on a client / experiment
type Variable struct {
	Name     string       `json:"name" validate:"required"`
	Required bool         `json:"required"`
	Type     VariableType `json:"type" validate:"required"`
}

// VariableConfig describes the request parsing configuration for a variable
type VariableConfig struct {
	Name        string             `json:"name" validate:"required"`
	Required    bool               `json:"required"`
	Field       string             `json:"field" validate:"required_with=Required"`
	FieldSource common.FieldSource `json:"field_source" validate:"field-src"`
}

// Variables represents the configuration of all experiment variables
type Variables struct {
	// ClientVariables represents the list of variables configured on the experiment engine's
	// client, if applicable
	ClientVariables []Variable `json:"client_variables" validate:"dive"`
	// ExperimentVariables is a map of experiment_id -> []Variable, representing the
	// variables configured on each experiment on the experiment engine
	ExperimentVariables map[string][]Variable `json:"experiment_variables" validate:"dive,dive"`
	// Config represents the request parsing configuration for all of the Turing experiment
	// variables
	Config []VariableConfig `json:"config" validate:"dive"`
}

// TuringExperimentConfig is the saved experiment config on Turing, that captures the key pieces
// of info on the experiment engine
type TuringExperimentConfig struct {
	Deployment struct {
		Endpoint string `json:"endpoint"`
		Timeout  string `json:"timeout"`
	} `json:"deployment"`
	Client      Client       `json:"client" validate:"dive"`
	Experiments []Experiment `json:"experiments" validate:"dive"`
	Variables   Variables    `json:"variables" validate:"dive"`
}

// ValidatorCtxKey is used to set values in the context object passed to the validator method,
// for context based validation.
type ValidatorCtxKey string

// engineValidatorCtxKey is used to pass in the Engine properties in the context
var engineValidatorCtxKey ValidatorCtxKey = "engine"

// newExperimentConfigValidator returns a default validator for the TuringExperimentConfig
func newExperimentConfigValidator() *validator.Validate {
	v := validator.New()
	// Register contextual validation method for TuringExperimentConfig
	v.RegisterStructValidationCtx(validateTuringExperimentConfig, TuringExperimentConfig{})
	// Field Source validation for expected values
	_ = v.RegisterValidation("field-src", validateFieldSource)

	return v
}

// validateTuringExperimentConfig is used to validate TuringExperimentConfig using the contextual
// information passed to it
func validateTuringExperimentConfig(ctx context.Context, sl validator.StructLevel) {
	config := sl.Current().Interface().(TuringExperimentConfig)
	engine, ok := ctx.Value(engineValidatorCtxKey).(Engine)
	if !ok {
		sl.ReportError("", "experiment_config", "Config", "missing-context-info", "")
	}

	if engine.ExperimentSelectionEnabled {
		// Check that there is at least 1 experiment
		if len(config.Experiments) < 1 {
			sl.ReportError(config.Experiments, "experiments", "Experiments", "no-experiment-selected", "")
		}
		// If Client Selection is enabled, check that the ClientID in each experiment matches the
		// client info passed in
		if engine.ClientSelectionEnabled {
			for _, e := range config.Experiments {
				if e.ClientID != config.Client.ID {
					sl.ReportError(config.Experiments, "experiments", "Experiments", "client-id-mismatch", "")
				}
			}
		}
	}
}

// validateFieldSource is used to check if the field source has an expected value
func validateFieldSource(fl validator.FieldLevel) bool {
	stringSrc := fl.Field().String()
	_, err := common.GetFieldSource(stringSrc)
	return err == nil
}
