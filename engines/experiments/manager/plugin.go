package manager

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

type ExperimentManagerPlugin struct {
	Impl ExperimentManager
}

func (p *ExperimentManagerPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &ExperimentManagerRPCServer{Impl: p.Impl}, nil
}

func (ExperimentManagerPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &ExperimentManagerRPC{client: c}, nil
}

type ExperimentManagerRPC struct{ client *rpc.Client }

func (g *ExperimentManagerRPC) GetEngineInfo() Engine {
	info := Engine{}
	err := g.client.Call("Plugin.GetEngineInfo", new(interface{}), &info)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		panic(err)
	}

	return info
}

type ExperimentManagerRPCServer struct {
	Impl ExperimentManager
}

func (s *ExperimentManagerRPCServer) GetEngineInfo(args interface{}, resp *Engine) error {
	*resp = s.Impl.GetEngineInfo()
	return nil
}
