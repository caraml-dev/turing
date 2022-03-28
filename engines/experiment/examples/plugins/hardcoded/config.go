package hardcoded

import (
	"encoding/json"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/pkg/request"
)

type TreatmentConfig struct {
	Traffic float32         `json:"traffic"`
	Data    json.RawMessage `json:"treatment_configuration"`
}

type SegmenterConfig struct {
	Name            string              `json:"name"`
	SegmenterSource request.FieldSource `json:"source"`
	SegmenterValue  string              `json:"value"`
}

type Experiment struct {
	manager.Experiment
	SegmentationConfig SegmenterConfig            `json:"segmentation_configuration"`
	VariantsConfig     map[string]TreatmentConfig `json:"variants_configuration"`
}

type ManagerConfig struct {
	Engine      manager.Engine      `json:"engine"`
	Experiments []Experiment        `json:"experiments"`
	Variables   []manager.Variables `json:"variables"`
}

type RunnerConfig struct {
	Experiments []Experiment `json:"experiments"`
}
