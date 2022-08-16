package main

import (
	"github.com/caraml-dev/turing/engines/experiment/examples/plugins/hardcoded"
	"github.com/caraml-dev/turing/engines/experiment/log"
	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc"
	"github.com/hashicorp/go-hclog"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{
		Level:      hclog.Info,
		Name:       "example-plugin",
		JSONFormat: true,
	})
	log.SetGlobalLogger(logger)

	rpc.Serve(&rpc.ClientServices{
		Manager: &hardcoded.ExperimentManager{},
		Runner:  &hardcoded.ExperimentRunner{},
	})
}
