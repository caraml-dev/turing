package main

import (
	"github.com/gojek/turing/engines/experiment/examples/plugins/nop"
	"github.com/gojek/turing/engines/experiment/plugin"
)

func main() {
	plugin.Serve(&plugin.ClientServices{
		Manager: &nop.ExperimentManager{},
		Runner:  &nop.ExperimentRunner{},
	})
}
