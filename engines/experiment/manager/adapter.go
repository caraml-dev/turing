package manager

/* The functions in this file can be used to safely invoke the underlying methods
   on the experiment managers, regardless of their type (standard vs custom).
*/

import (
	"errors"
)

const (
	standardMethodErr = "Method is only supported by standard experiment managers"
)

func IsStandardExperimentManager(expManager ExperimentManager) bool {
	engineInfo, err := expManager.GetEngineInfo()
	return err == nil && engineInfo.Type == StandardExperimentManagerType
}

// StandardExperimentManager methods ******************************************

func IsCacheEnabled(expManager ExperimentManager) bool {
	if IsStandardExperimentManager(expManager) {
		if stdMgr, ok := expManager.(StandardExperimentManager); ok {
			cacheEnabled, err := stdMgr.IsCacheEnabled()
			if err != nil {
				return false
			}
			return cacheEnabled
		}
	}
	return false
}

func ListClients(expManager ExperimentManager) ([]Client, error) {
	if IsStandardExperimentManager(expManager) {
		if stdMgr, ok := expManager.(StandardExperimentManager); ok {
			return stdMgr.ListClients()
		}
	}
	return []Client{}, errors.New(standardMethodErr)
}

func ListExperiments(expManager ExperimentManager) ([]Experiment, error) {
	if IsStandardExperimentManager(expManager) {
		if stdMgr, ok := expManager.(StandardExperimentManager); ok {
			return stdMgr.ListExperiments()
		}
	}
	return []Experiment{}, errors.New(standardMethodErr)
}

func ListExperimentsForClient(expManager ExperimentManager, client Client) ([]Experiment, error) {
	if IsStandardExperimentManager(expManager) {
		if stdMgr, ok := expManager.(StandardExperimentManager); ok {
			return stdMgr.ListExperimentsForClient(client)
		}
	}
	return []Experiment{}, errors.New(standardMethodErr)
}

func ListVariablesForClient(expManager ExperimentManager, client Client) ([]Variable, error) {
	if IsStandardExperimentManager(expManager) {
		if stdMgr, ok := expManager.(StandardExperimentManager); ok {
			return stdMgr.ListVariablesForClient(client)
		}
	}
	return []Variable{}, errors.New(standardMethodErr)
}

func ListVariablesForExperiments(expManager ExperimentManager, exps []Experiment) (map[string][]Variable, error) {
	if IsStandardExperimentManager(expManager) {
		if stdMgr, ok := expManager.(StandardExperimentManager); ok {
			return stdMgr.ListVariablesForExperiments(exps)
		}
	}
	return map[string][]Variable{}, errors.New(standardMethodErr)
}
