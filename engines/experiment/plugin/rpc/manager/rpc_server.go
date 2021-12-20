package manager

import (
	"encoding/json"

	"github.com/gojek/turing/engines/experiment/manager"
)

// rpcServer serves the implementation of a ConfigurableExperimentManager
type rpcServer struct {
	Impl ConfigurableExperimentManager
}

func (s *rpcServer) Configure(cfg json.RawMessage, _ *interface{}) (err error) {
	return s.Impl.Configure(cfg)
}

func (s *rpcServer) GetEngineInfo(_ interface{}, resp *manager.Engine) error {
	*resp = s.Impl.GetEngineInfo()
	return nil
}

func (s *rpcServer) ValidateExperimentConfig(cfg json.RawMessage, _ *interface{}) (err error) {
	return s.Impl.ValidateExperimentConfig(cfg)
}

func (s *rpcServer) GetExperimentRunnerConfig(inConfig json.RawMessage, outConfig *json.RawMessage) (err error) {
	*outConfig, err = s.Impl.GetExperimentRunnerConfig(inConfig)
	return
}
