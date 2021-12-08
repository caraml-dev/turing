package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/rpc"

	"github.com/gojek/turing/engines/experiment/manager"
	"github.com/gojek/turing/engines/experiment/runner"
	"github.com/hashicorp/go-plugin"
)

type rpcClient struct {
	*rpc.Client
}

func (c *rpcClient) Capabilities() *Capabilities {
	var resp Capabilities
	if err := c.Call("Plugin.Capabilities", new(interface{}), &resp); err != nil {
		panic(err)
	}
	return &resp
}

func (c *rpcClient) Configure(cfg json.RawMessage) error {
	return c.Call("Plugin.Configure", cfg, new(interface{}))
}

func (c *rpcClient) GetEngineInfo() manager.Engine {
	resp := manager.Engine{}
	err := c.Call("Plugin.GetEngineInfo", new(interface{}), &resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return resp
}

func (c *rpcClient) GetTreatmentForRequest(
	ctx context.Context,
	log runner.Logger,
	header http.Header,
	payload []byte,
) (*runner.Treatment, error) {
	req := GetTreatmentRequest{
		Context: ctx,
		//Logger:  log,
		Header:  header,
		Payload: payload,
	}
	var resp runner.Treatment

	err := c.Call("Plugin.GetTreatmentForRequest", &req, &resp)

	println(fmt.Sprintf("%v", err))
	return &resp, err
}

type rpcServer struct {
	ManagerImpl ConfigurableExperimentManager
	RunnerImpl  ConfigurableExperimentRunner
}

func (s *rpcServer) Capabilities(_ interface{}, resp *Capabilities) error {
	*resp = Capabilities{
		Manager: s.ManagerImpl != nil,
		Runner:  s.RunnerImpl != nil,
	}
	return nil
}

func (s *rpcServer) Configure(cfg json.RawMessage, _ *interface{}) (err error) {
	if s.ManagerImpl != nil {
		if err = s.ManagerImpl.Configure(cfg); err != nil {
			return
		}
	}
	if s.RunnerImpl != nil {
		return s.RunnerImpl.Configure(cfg)
	}
	return
}

func (s *rpcServer) GetEngineInfo(_ interface{}, resp *manager.Engine) error {
	*resp = s.ManagerImpl.GetEngineInfo()
	return nil
}

func (s *rpcServer) GetTreatmentForRequest(req *GetTreatmentRequest, resp *runner.Treatment) error {
	treatment, err := s.RunnerImpl.GetTreatmentForRequest(
		req.Context, req.Logger, req.Header, req.Payload,
	)
	println(fmt.Sprintf("%v", treatment))
	if err != nil {
		return err
	}

	*resp = *treatment
	return nil
}

func Serve(services *Services) {
	plugins := plugin.PluginSet{
		ExperimentsEnginePluginIdentifier: &ExperimentEnginePlugin{
			ManagerImpl: services.Manager,
			RunnerImpl:  services.Runner,
		},
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         plugins,
	})
}
