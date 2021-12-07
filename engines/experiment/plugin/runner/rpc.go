package runner

import (
	"context"
	"encoding/json"
	"net/http"
	"net/rpc"

	"github.com/gojek/turing/engines/experiment/runner"
)

type rpcClient struct {
	*rpc.Client
}

func (c *rpcClient) Configure(cfg json.RawMessage) error {
	return c.Call("Plugin.Configure", cfg, new(interface{}))
}

func (c *rpcClient) GetTreatmentForRequest(
	ctx context.Context,
	log runner.Logger,
	header http.Header,
	payload []byte,
) (runner.Treatment, error) {
	req := getTreatmentRequest{
		ctx:     ctx,
		log:     log,
		header:  header,
		payload: payload,
	}
	var resp runner.Treatment

	err := c.Call("Plugin.GetTreatmentForRequest", &req, &resp)
	return resp, err
}

type rpcServer struct {
	Impl ConfigurableExperimentRunner
}

func (s *rpcServer) Configure(cfg json.RawMessage, _ *interface{}) error {
	return s.Impl.Configure(cfg)
}

func (s *rpcServer) GetTreatmentForRequest(req *getTreatmentRequest, resp *runner.Treatment) (err error) {
	resp, err = s.Impl.GetTreatmentForRequest(
		req.ctx, req.log, req.header, req.payload,
	)
	return
}
