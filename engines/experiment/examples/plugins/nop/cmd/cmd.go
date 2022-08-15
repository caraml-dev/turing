package main

import (
	"github.com/caraml-dev/turing/engines/experiment/examples/plugins/nop"
	"github.com/caraml-dev/turing/engines/experiment/plugin/rpc"
)

func main() {
	rpc.Serve(&rpc.ClientServices{
		Manager: &nop.ExperimentManager{},
		Runner:  &nop.ExperimentRunner{},
	})
}
