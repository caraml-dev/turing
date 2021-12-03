package main

import "github.com/gojek/turing/engines/experiment/plugin"

func main() {
	plugin.Serve(&plugin.Services{
		Manager: &ExperimentManager{},
	})
}
