package main

import (
	"github.com/gojek/turing/engines/experiment/examples/plugins/hardcoded"
	"github.com/gojek/turing/engines/experiment/log"
	"github.com/gojek/turing/engines/experiment/plugin/rpc"
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
