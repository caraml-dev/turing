package manager

import (
	common "github.com/gojek/turing/engines/experiment/common"
)

type ExperimentManagerType string

var (
	StandardExperimentManagerType ExperimentManagerType = "standard"
	CustomExperimentManagerType   ExperimentManagerType = "custom"
)

type StandardExperimentManagerConfig struct {
	// ClientSelectionEnabled is set to true if the experiment engine has the concept of
	// clients, that can be used to authenticate the experiment run requests.
	ClientSelectionEnabled bool `json:"client_selection_enabled"`
	// ExperimentSelectionEnabled is set to true if the experiment engine allows selecting
	// individual experiments configured elsewhere
	ExperimentSelectionEnabled bool `json:"experiment_selection_enabled"`
	// HomePageURL is an optional string which, if set, will be used by the UI to redirect
	// to the experiment engine's home page, to view more details on the experiment
	// configured in Turing.
	HomePageURL string `json:"home_page_url"`
}

type RemoteUI struct {
	// Name is the name of the remote app declared in the Module Federation plugin
	Name string `json:"name"`
	// URL is the Host + Remote Entry file at which the remote UI can be found
	URL string `json:"url"`
}

type CustomExperimentManagerConfig struct {
	// RemoteUI specifies the information for the custom experiment engine UI to be
	// consumed by the Turing app, using Module Federation
	RemoteUI RemoteUI `json:"remote_ui"`
}

// Engine describes the properties of an experiment engine
type Engine struct {
	// Name is the display name used for the experiment engine.
	Name string `json:"name"`
	// Type describes the class of the experiment engine manager
	Type ExperimentManagerType `json:"type"`
	// StandardExperimentManagerConfig is expected to be set by a "standard" experiment engine manager
	// and is used by the generic Turing experiment engine UI.
	StandardExperimentManagerConfig *StandardExperimentManagerConfig `json:"standard_experiment_manager_config,omitempty"`
	// CustomExperimentManagerConfig is expected to be set by a "custom" experiment engine manager
	// and is used to load the custom experiment engine UI.
	CustomExperimentManagerConfig *CustomExperimentManagerConfig `json:"custom_experiment_manager_config,omitempty"`
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

// TuringExperimentConfig is the saved experiment config on Turing, when using the generic UI,
// that captures the key pieces of info about the experiment engine
type TuringExperimentConfig struct {
	Client      Client       `json:"client" validate:"dive"`
	Experiments []Experiment `json:"experiments" validate:"dive"`
	Variables   Variables    `json:"variables" validate:"dive"`
}
