package main

import (
	"github.com/gojek/turing/engines/experiment/examples/plugins/nop"
	"github.com/gojek/turing/engines/experiment/plugin/rpc"
)

func main() {
	rpc.Serve(&rpc.ClientServices{
		Manager: &nop.ExperimentManager{},
		Runner:  &nop.ExperimentRunner{},
	})
}
