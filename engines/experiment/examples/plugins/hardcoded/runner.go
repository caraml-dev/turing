package hardcoded

import (
	"encoding/json"
	"errors"
	"github.com/gojek/turing/engines/experiment/examples/plugins/hardcoded/utils"
	"github.com/gojek/turing/engines/experiment/pkg/request"
	"github.com/gojek/turing/engines/experiment/runner"
	"net/http"
)

type ExperimentRunner struct {
	experiments []Experiment
}

func (e *ExperimentRunner) Configure(cfg json.RawMessage) error {
	var config RunnerConfig
	err := json.Unmarshal(cfg, &config)
	if err != nil {
		return err
	}

	e.experiments = config.Experiments
	return nil
}

func (e *ExperimentRunner) GetTreatmentForRequest(
	header http.Header,
	payload []byte,
	_ runner.GetTreatmentOptions,
) (*runner.Treatment, error) {
	for _, exp := range e.experiments {
		segmentationUnit, err := request.GetValueFromRequest(
			header,
			payload,
			exp.SegmentationConfig.SegmenterSource,
			exp.SegmentationConfig.SegmenterValue)
		if err != nil {
			continue
		}

		bucket := utils.Hash(segmentationUnit) % 10000

		var total uint32 = 0
		for name, variant := range exp.VariantsConfig {
			total += uint32(variant.Traffic * 10000)
			if bucket < total {
				return &runner.Treatment{
					ExperimentName: exp.Name,
					Name:           name,
					Config:         variant.Data,
				}, nil
			}
		}
	}

	return nil, errors.New("no experiment configured for the unit")
}
