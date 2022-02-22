package main

import (
	"github.com/gojek/turing/engines/experiment/examples/plugins/hardcoded"
	"github.com/gojek/turing/engines/experiment/plugin/rpc"
)

func main() {
	rpc.Serve(&rpc.ClientServices{
		Manager: &hardcoded.ExperimentManager{},
		Runner:  &hardcoded.ExperimentRunner{},
	})
}
