package manager

/* The functions in this file can be used to safely invoke the underlying methods
   on the experiment managers, regardless of their type (standard vs custom).
*/

import (
	"errors"
)

const (
	experimentManagerCastingErr = "Error casting %s to %s experiment manager"
	standardExperimentConfigErr = "Unable to parse standard experiment config: %v"
	standardMethodErr           = "Method is only supported by standard experiment managers"
	unknownManagerTypeErr       = "Experiment Manager type %s is not recognized"
)

// StandardExperimentManager methods ******************************************

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
