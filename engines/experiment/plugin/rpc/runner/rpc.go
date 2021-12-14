package runner

import (
	"context"
	"encoding/json"
	"github.com/gojek/turing/engines/experiment/plugin/rpc/shared"
	"net/http"

	"github.com/gojek/turing/engines/experiment/runner"
)

// rpcClient implements ConfigurableExperimentRunner interface
type rpcClient struct {
	shared.RPCClient
}

func (c *rpcClient) Configure(cfg json.RawMessage) error {
	return c.Call("Plugin.Configure", cfg, new(interface{}))
}

func (c *rpcClient) GetTreatmentForRequest(
	ctx context.Context,
	log runner.Logger,
	header http.Header,
	payload []byte,
) (*runner.Treatment, error) {
	req := GetTreatmentRequest{
		Context: ctx,
		Logger:  log,
		Header:  header,
		Payload: payload,
	}
	var resp runner.Treatment

	err := c.Call("Plugin.GetTreatmentForRequest", &req, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// rpcServer serves the implementation of a ConfigurableExperimentRunner
type rpcServer struct {
	Impl ConfigurableExperimentRunner
}

func (s *rpcServer) Configure(cfg json.RawMessage, _ *interface{}) (err error) {
	return s.Impl.Configure(cfg)
}

func (s *rpcServer) GetTreatmentForRequest(req *GetTreatmentRequest, resp *runner.Treatment) error {
	treatment, err := s.Impl.GetTreatmentForRequest(
		req.Context, req.Logger, req.Header, req.Payload,
	)
	if err != nil {
		return err
	}

	*resp = *treatment
	return nil
}
