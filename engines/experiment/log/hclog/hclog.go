package hclog

import (
	"github.com/caraml-dev/turing/engines/experiment/log"
)

// When this package is imported, the global logger is replaced with hcLogger,
// which is recommended to be used in RPC plugins as it provides a structured
// logging output in the host application, that calls the plugin implementation
func init() {
	log.SetGlobalLogger(log.DefaultHCLogger())
}
