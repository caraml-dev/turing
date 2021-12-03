package manager

import (
	"encoding/json"
	"github.com/gojek/turing/engines/experiment/manager"
)

type rpcServer struct {
	Impl ConfigurableExperimentManager
}

func (s *rpcServer) Configure(cfg json.RawMessage, _ *interface{}) error {
	return s.Impl.Configure(cfg)
}

func (s *rpcServer) GetEngineInfo(_ interface{}, resp *manager.Engine) error {
	*resp = s.Impl.GetEngineInfo()
	return nil
}
