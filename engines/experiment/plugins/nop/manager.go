package main

import "github.com/gojek/turing/engines/experiment/manager"

type ExperimentManager struct{}

func (ExperimentManager) GetEngineInfo() manager.Engine {
	return manager.Engine{
		Name:        "nop",
		DisplayName: "None",
		Type:        manager.StandardExperimentManagerType,
	}
}
