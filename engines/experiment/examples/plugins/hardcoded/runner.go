package hardcoded

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	"github.com/caraml-dev/turing/engines/experiment/examples/plugins/hardcoded/utils"
	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
	"github.com/caraml-dev/turing/engines/experiment/runner"
	"github.com/gojek/mlp/api/pkg/instrumentation/metrics"
)

type ExperimentRunner struct {
	experiments      []Experiment
	sortedTreatments map[string][]string
}

func (e *ExperimentRunner) Configure(cfg json.RawMessage) error {
	var config RunnerConfig
	err := json.Unmarshal(cfg, &config)
	if err != nil {
		return err
	}
	e.experiments = config.Experiments
	e.sortedTreatments = make(map[string][]string)

	for _, exp := range e.experiments {
		var variants []string

		for name := range exp.VariantsConfig {
			variants = append(variants, name)
		}

		sort.Slice(variants, func(i, j int) bool {
			return exp.VariantsConfig[variants[i]].Traffic > exp.VariantsConfig[variants[j]].Traffic
		})

		e.sortedTreatments[exp.ID] = variants
	}
	return nil
}

func (e *ExperimentRunner) GetTreatmentForRequest(
	header http.Header,
	payload []byte,
	_ runner.GetTreatmentOptions,
) (*runner.Treatment, error) {
	for _, exp := range e.experiments {
		segmentationUnit, err := request.GetValueFromHTTPRequest(
			header,
			payload,
			exp.SegmentationConfig.SegmenterSource,
			exp.SegmentationConfig.SegmenterValue)
		if err != nil {
			continue
		}

		bucket := utils.Hash(segmentationUnit) % 10000

		var total uint32 = 0
		for _, name := range e.sortedTreatments[exp.ID] {
			variant := exp.VariantsConfig[name]
			total += uint32(variant.Traffic * 10000)
			if bucket <= total {
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

func (r *ExperimentRunner) RegisterMetrics(_ metrics.Collector, _ runner.MetricsRegistrationHelper) error {
	return nil
}
