package manager

/* The functions in this file can be used to safely invoke the underlying methods
   on the experiment managers, regardless of their type (standard vs custom).
*/

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	experimentManagerCastingErr = "Error casting %s to %s experiment manager"
	standardExperimentConfigErr = "Unable to parse standard experiment config: %v"
	standardMethodErr           = "Method is only supported by standard experiment managers"
)

func IsCacheEnabled(expManager ExperimentManager) bool {
	if expManager.GetEngineInfo().Type == StandardExperimentManagerType {
		if stdMgr, ok := expManager.(StandardExperimentManager); ok {
			return stdMgr.IsCacheEnabled()
		}
	}
	return false
}

func ListClients(expManager ExperimentManager) ([]Client, error) {
	if expManager.GetEngineInfo().Type == StandardExperimentManagerType {
		if stdMgr, ok := expManager.(StandardExperimentManager); ok {
			return stdMgr.ListClients()
		}
	}
	return []Client{}, errors.New(standardMethodErr)
}

func ListExperiments(expManager ExperimentManager) ([]Experiment, error) {
	if expManager.GetEngineInfo().Type == StandardExperimentManagerType {
		if stdMgr, ok := expManager.(StandardExperimentManager); ok {
			return stdMgr.ListExperiments()
		}
	}
	return []Experiment{}, errors.New(standardMethodErr)
}

func ListExperimentsForClient(expManager ExperimentManager, client Client) ([]Experiment, error) {
	if expManager.GetEngineInfo().Type == StandardExperimentManagerType {
		if stdMgr, ok := expManager.(StandardExperimentManager); ok {
			return stdMgr.ListExperimentsForClient(client)
		}
	}
	return []Experiment{}, errors.New(standardMethodErr)
}

func ListVariablesForClient(expManager ExperimentManager, client Client) ([]Variable, error) {
	if expManager.GetEngineInfo().Type == StandardExperimentManagerType {
		if stdMgr, ok := expManager.(StandardExperimentManager); ok {
			return stdMgr.ListVariablesForClient(client)
		}
	}
	return []Variable{}, errors.New(standardMethodErr)
}

func ListVariablesForExperiments(expManager ExperimentManager, exps []Experiment) (map[string][]Variable, error) {
	if expManager.GetEngineInfo().Type == StandardExperimentManagerType {
		if stdMgr, ok := expManager.(StandardExperimentManager); ok {
			return stdMgr.ListVariablesForExperiments(exps)
		}
	}
	return map[string][]Variable{}, errors.New(standardMethodErr)
}

func GetExperimentRunnerConfig(expManager ExperimentManager, expCfg interface{}) (json.RawMessage, error) {
	engineInfo := expManager.GetEngineInfo()
	// Call the appropriate validator method based on the type of the experiment manager
	switch engineInfo.Type {
	case StandardExperimentManagerType:
		stdMgr, ok := expManager.(StandardExperimentManager)
		if !ok {
			return json.RawMessage{}, fmt.Errorf(experimentManagerCastingErr, engineInfo.Name, "standard")
		}
		stdExpConfig, err := GetStandardExperimentConfig(expCfg)
		if err != nil {
			return json.RawMessage{}, fmt.Errorf(standardExperimentConfigErr, err)
		}
		return stdMgr.GetExperimentRunnerConfig(stdExpConfig)
	default:
		customMgr, ok := expManager.(CustomExperimentManager)
		if !ok {
			return json.RawMessage{}, fmt.Errorf(experimentManagerCastingErr, engineInfo.Name, "custom")
		}
		return customMgr.GetExperimentRunnerConfig(expCfg)
	}
}

func ValidateExperimentConfig(expManager ExperimentManager, expCfg interface{}) error {
	engineInfo := expManager.GetEngineInfo()
	// Call the appropriate validator method based on the type of the experiment manager
	switch engineInfo.Type {
	case StandardExperimentManagerType:
		stdMgr, ok := expManager.(StandardExperimentManager)
		if !ok {
			return fmt.Errorf(experimentManagerCastingErr, engineInfo.Name, "standard")
		}
		stdExpConfig, err := GetStandardExperimentConfig(expCfg)
		if err != nil {
			return fmt.Errorf(standardExperimentConfigErr, err)
		}
		return stdMgr.ValidateExperimentConfig(engineInfo.StandardExperimentManagerConfig, stdExpConfig)
	default:
		customMgr, ok := expManager.(CustomExperimentManager)
		if !ok {
			return fmt.Errorf(experimentManagerCastingErr, engineInfo.Name, "custom")
		}
		return customMgr.ValidateExperimentConfig(expCfg)
	}
}
