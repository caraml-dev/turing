package runner

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/caraml-dev/turing/engines/experiment/runner"
)

// All registered experiment runners.
var runnersLock sync.Mutex
var runners = make(map[string]Factory)

// Factory creates an experiment runner with the provided config.
// Config is a raw encoded JSON value. The runner implementation
// should provide a schema and example of the JSON value to explain the usage.
type Factory func(config json.RawMessage) (runner.ExperimentRunner, error)

// Register a runner with the provided name and factory function.
// For registration to be properly recorded, this should be called in the init
// phase of runtime. The init function is usually defined in the package where
// the runner is implemented. The name of the runner should be unique
// across all runner implementations. Registering multiple runners with the
// same name will return an error.
func Register(name string, factory Factory) error {
	runnersLock.Lock()
	defer runnersLock.Unlock()

	name = strings.ToLower(name)
	if _, found := runners[name]; found {
		return fmt.Errorf("experiment runner %q was registered twice", name)
	}

	runners[name] = factory
	return nil
}

// Get a runner that has been registered. Retrieving a runner not yet registered
// will return an error.
func Get(name string, config json.RawMessage) (runner.ExperimentRunner, error) {
	runnersLock.Lock()
	defer runnersLock.Unlock()

	name = strings.ToLower(name)
	r, ok := runners[name]
	if !ok {
		return nil, fmt.Errorf("no experiment runner found for name %q", name)
	}

	return r(config)
}
