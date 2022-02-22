package hardcoded

import (
	"encoding/json"
	"github.com/gojek/turing/engines/experiment/manager"
)

type ExperimentManager struct {
	*manager.BaseStandardExperimentManager
	experiments map[string]Experiment
}

func (e *ExperimentManager) Configure(cfg json.RawMessage) error {
	var config ManagerConfig

	err := json.Unmarshal(cfg, &config)
	if err != nil {
		return err
	}

	e.BaseStandardExperimentManager = manager.NewBaseStandardExperimentManager(config.Engine)
	e.experiments = make(map[string]Experiment)
	for _, exp := range config.Experiments {
		e.experiments[exp.Name] = exp
	}
	return nil
}

func (e *ExperimentManager) ListClients() ([]manager.Client, error) {
	return []manager.Client{}, nil
}

func (e *ExperimentManager) ListExperiments() ([]manager.Experiment, error) {
	var experiments []manager.Experiment
	for _, exp := range e.experiments {
		experiments = append(experiments, exp.Experiment)
	}

	return experiments, nil
}

func (e *ExperimentManager) ListExperimentsForClient(manager.Client) ([]manager.Experiment, error) {
	return e.ListExperiments()
}

func (e ExperimentManager) ListVariablesForClient(manager.Client) ([]manager.Variable, error) {
	return []manager.Variable{}, nil
}

func (e ExperimentManager) ListVariablesForExperiments(
	experiments []manager.Experiment,
) (map[string][]manager.Variable, error) {
	return map[string][]manager.Variable{}, nil
}

func (e ExperimentManager) GetExperimentRunnerConfig(cfg json.RawMessage) (json.RawMessage, error) {
	standardExpCfg, err := manager.ParseStandardExperimentConfig(cfg)
	if err != nil {
		return nil, err
	}

	runnerExperimentConfigs := make([]Experiment, len(standardExpCfg.Experiments))
	for idx, exp := range standardExpCfg.Experiments {
		runnerExperimentConfigs[idx] = e.experiments[exp.Name]

		for _, variable := range standardExpCfg.Variables.ExperimentVariables[exp.ID] {
			if variable.Type == manager.UnitVariableType {
				for _, varConfig := range standardExpCfg.Variables.Config {
					if varConfig.Name == variable.Name {
						runnerExperimentConfigs[idx].SegmentationConfig = SegmenterConfig{
							Name:            variable.Name,
							SegmenterSource: varConfig.FieldSource,
							SegmenterValue:  varConfig.Field,
						}
						break
					}
				}
				break
			}
		}
	}

	return json.Marshal(
		RunnerConfig{
			Experiments: runnerExperimentConfigs,
		})
}
