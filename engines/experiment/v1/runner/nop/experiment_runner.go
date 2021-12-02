package nop

import (
	"log"

	"github.com/gojek/turing/engines/experiment/runner/nop"
	v1 "github.com/gojek/turing/engines/experiment/v1/runner"
)

// init ensures this runner is registered when the package is imported.
func init() {
	err := v1.Register("nop", nop.NewExperimentRunner)
	if err != nil {
		log.Fatal(err)
	}
}
