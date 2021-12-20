package manager

import (
	"encoding/json"
	"errors"

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

// Methods of manager.StandardExperimentManager are served below this line

func (s *rpcServer) IsCacheEnabled(_ interface{}, resp *bool) error {
	return s.asStandardManager(func(sm manager.StandardExperimentManager) error {
		*resp = sm.IsCacheEnabled()
		return nil
	})
}

func (s *rpcServer) ListClients(_ interface{}, resp *[]manager.Client) error {
	return s.asStandardManager(func(sm manager.StandardExperimentManager) (err error) {
		*resp, err = sm.ListClients()
		return
	})
}

func (s *rpcServer) ListExperiments(_ interface{}, resp *[]manager.Experiment) error {
	return s.asStandardManager(func(sm manager.StandardExperimentManager) (err error) {
		*resp, err = sm.ListExperiments()
		return
	})
}

func (s *rpcServer) ListExperimentsForClient(client manager.Client, resp *[]manager.Experiment) error {
	return s.asStandardManager(func(sm manager.StandardExperimentManager) (err error) {
		*resp, err = sm.ListExperimentsForClient(client)
		return
	})
}

func (s *rpcServer) ListVariablesForClient(client manager.Client, resp *[]manager.Variable) error {
	return s.asStandardManager(func(sm manager.StandardExperimentManager) (err error) {
		*resp, err = sm.ListVariablesForClient(client)
		return
	})
}

func (s *rpcServer) ListVariablesForExperiments(
	experiments []manager.Experiment,
	resp *map[string][]manager.Variable,
) error {
	return s.asStandardManager(func(sm manager.StandardExperimentManager) (err error) {
		*resp, err = sm.ListVariablesForExperiments(experiments)
		return
	})
}

func (s *rpcServer) asStandardManager(fn func(manager.StandardExperimentManager) error) error {
	standardManager, ok := s.Impl.(manager.StandardExperimentManager)
	if !ok {
		return errors.New("not implemented")
	}

	return fn(standardManager)
}
