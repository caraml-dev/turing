 # RPC Experiment Engine Plugins

It's possible to integrate external experiment engines with Turing by implementing experiment engine plugin 
and deploying it together with Turing app. Turing supports RPC experiment engine plugins via [Hashicorp go-plugin](
https://github.com/hashicorp/go-plugin) library.


## Implement a plugin

---
[Experiment Engine](./developer_guide.md#adding-a-new-experiment-engine) plugin serves the role of an adapter 
between the Experiment Engine backend and Turing Server/Turing Router. It's also possible to support an 
Experiment Engine, that doesn't have a backend at all. In this case, the plugin must implement the logic for 
managing experiments and assigning experiment treatments to requests internally.

Currently, Turing supports the Experiment Engine plugins written in Golang. Experiment Engine plugin is an 
executable Golang binary that is launched by the Turing Server/Router as a child process. Inter-process 
communication (IPC) between the parent (Turing Server/Router) and the child (Experiment Engine plugin) processes
is done through the local unix socket. For more information about the `hashicorp/go-plugin`'s internals, please 
check its [official documentation](https://github.com/hashicorp/go-plugin#readme).

For the correct work, the Experiment Engine plugin must implement both [Experiment Manager](developer_guide.md#experiment-manager)
and [Experiment Runner](./developer_guide.md#experiment-runner) interfaces.

Turing Server interacts with the Experiment Manager interface of the plugin to fetch the information about 
experiments configuration, which is used to deploy Turing router.
![experiment_manager_data_types](./assets/exp_manager_plugin_diagram.png)

Turing Router interacts with the Experiment Runner interface of the plugin to "run" experiments and assign an 
experiment treatment to each incoming request to the router.
![experiment_manager_data_types](./assets/exp_runner_plugin_diagram.png)

### Pre-requisites 
 * Golang 1.14
 * Git
 * Docker

### Implementation
This repository contains a simple serverless Experiment Engine plugin, which can be used as a reference for the plugin
implementation. 

First clone the Turing repo and change the workdir to the reference plugin root directory: 
```shell
$ git clone https://github.com/gojek/turing.git
$ cd turing/engines/experiment/examples/plugins/hardcoded
```

A Go plugin is a standalone application which serves the Experiment Engine implementation by calling 
`rpc.Serve(&rpc.ClientServices{})`. The entrypoint of the application is [`cmd/main.go`](../examples/plugins/hardcoded/cmd/main.go):

```go
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
```
Note that the main function is intentionally simplified to just a call to `rpc.Serve` function. It's entirely 
possible, that for the more complex plugin implementations it would be required to do some extra setup of the plugin 
before it gets served. 

All the bootstrapping code, required to serve the Experiment Engine plugin via the RPC, is provided by the Turing
library, which is added into the plugin module as the dependency in [`go.mod`](../examples/plugins/hardcoded/go.mod):
```shell
module github.com/gojek/turing/engines/experiment/examples/plugins/hardcoded

go 1.14

require github.com/gojek/turing/engines/experiment v1.0.0

replace github.com/gojek/turing/engines/experiment => ../../../
```
NOTE: Since this example plugin module and the `engines/experiment` library belong to the same repository, the plugin
resolves `engines/experiment` package locally and that's the reason why the `replace ...` line in added to the 
bottom of `go.mod`.

A plugin needs to serve the implementation of [Experiment Manager](./developer_guide.md#experiment-manager) (either 
[Standard](./developer_guide.md#standard-experiment-manager) or [Custom](./developer_guide.md#custom-experiment-manager)) 
and [Experiment Runner](./developer_guide.md#experiment-runner). 

#### Experiment Manager

Both Standard and Custom experiment managers should implement the `ConfigurableExperimentManager` interface of:
```go
// ConfigurableExperimentManager interface of an ExperimentManager, that can be configured
// with an arbitrary configuration passed as a JSON data
type ConfigurableExperimentManager interface {
    shared.Configurable
    manager.ExperimentManager
}
```

Which is composed of the `ExperimentManager` interface of:
```go
type ExperimentManager interface {
    // GetEngineInfo returns the configuration of the experiment engine
    GetEngineInfo() (Engine, error)
    
    // ValidateExperimentConfig validates the given Turing experiment config for the expected data and format
    ValidateExperimentConfig(cfg json.RawMessage) error
    
    // GetExperimentRunnerConfig converts the given config (as retrieved from the DB) into a format suitable
    // for the Turing router (i.e., to be passed to the Experiment Runner). This interface method will be
    // called at the time of router deployment.
    //
    // cfg holds the experiment configuration in a format that is suitable for use with the Turing UI and
    // this is the data that is saved to the Turing DB.
    //
    // In case of StandardExperimentManager, cfg is expected to be unmarshalled into TuringExperimentConfig
    GetExperimentRunnerConfig(cfg json.RawMessage) (json.RawMessage, error)
}
```
And `Configurable` interface of:
```go
type Configurable interface {
	Configure(cfg json.RawMessage) error
}
```
More details about the plugin configuration and the role of `Configurable` method in it can be found in the 
[Plugin Configuration](./rpc_plugins.md#plugin-configuration) section.

Additionally, the Standard experiment manager should implement the StandardExperimentManager interface of:
```go
type StandardExperimentManager interface {
	ExperimentManager
	// IsCacheEnabled returns whether the experiment engine wants to cache its responses in the Turing API cache
	IsCacheEnabled() (bool, error)
	// ListClients returns a list of the clients registered on the experiment engine
	ListClients() ([]Client, error)
	// ListExperiments returns a list of the experiments registered on the experiment engine
	ListExperiments() ([]Experiment, error)
	// ListExperimentsForClient returns a list of the experiments registered on the experiment engine,
	// for the given client
	ListExperimentsForClient(Client) ([]Experiment, error)
	// ListVariablesForClient returns a list of the variables registered on the given client
	ListVariablesForClient(Client) ([]Variable, error)
	// ListVariablesForExperiments returns a list of the variables registered on the given experiments
	ListVariablesForExperiments([]Experiment) (map[string][]Variable, error)
}
```

A simple serverless `ExperimentManager` implementation, that receives the static experiment configuration 
at the initialization (via `Configure(...)` method) and serves this data to the Turing Server, can be found 
in the [`manager.go`](../examples/plugins/hardcoded/manager.go).

#### Experiment Runner

A plugin should also serve the implementation of ExperimentRunner, that implements `ConfigurableExperimentRunner` 
interface of:
```go
// ConfigurableExperimentRunner interface of an ExperimentRunner, that can be configured
// with an arbitrary configuration passed as a JSON data
type ConfigurableExperimentRunner interface {
	shared.Configurable
	runner.ExperimentRunner
}
```
Which is composed of the `ExperimentRunner` interface of:
```go
type ExperimentRunner interface {
	// GetTreatmentForRequest assigns a treatment to the given request
	GetTreatmentForRequest(
		header http.Header,
		payload []byte,
		options GetTreatmentOptions,
	) (*Treatment, error)
}
```
And `Configurable` interface of:
```go
type Configurable interface {
	Configure(cfg json.RawMessage) error
}
```
More details about the plugin configuration and the role of `Configurable` method in it can be found in the
[Plugin Configuration](./rpc_plugins.md#plugin-configuration) section.

A simple `ExperimentRunner` implementation that assigns a treatment to the request based on the hardcoded experiment 
configuration and the traffic weights configured for each treatment can be found in the 
[`runner.go`](../examples/plugins/hardcoded/runner.go).

### Plugin Configuration

During the initialization, Turing Server/Router configures the plugin with the configuration data. 

More specifically, Turing Server passes the arbitrary JSON configuration, defined in Turing config file during 
the deployment, to the ExperimentManager's `Configure(cfg json.RawMessage) error` method. The specific implementation
of the plugin can parse this JSON data into the expected data structure and use it to control ExperimentManager's logic.
In the [provided example](../examples/plugins/hardcoded/manager.go), passed JSON configuration is parsed as an 
instance of [ManagerConfig](../examples/plugins/hardcoded/config.go), which consists of the engine metadata and a
list of experiments that this Experiment Manager should be aware of: 

```go
func (e *ExperimentManager) Configure(cfg json.RawMessage) error {
	var config ManagerConfig

	err := json.Unmarshal(cfg, &config)
	if err != nil {
		return err
	}

	e.BaseStandardExperimentManager = manager.NewBaseStandardExperimentManager(config.Engine)
	e.experiments = make(map[string]Experiment)
	for _, exp := range config.Experiments {
		e.experiments[exp.Name] = exp
	}
	return nil
}
```
Example configuration of this ExperimentManager can be found in the [`configs/plugin_config_example.yaml`](../examples/plugins/hardcoded/configs/plugin_config_example.yaml).
Note that it's plugin's responsibility to set the expectations and define the contract for the plugin's configuration data.
It's also possible that Experiment Engine plugin does not require any external configuration. In this case, `Configure` 
method can be implemented as:
```go
func (e *ExperimentManager) Configure(json.RawMessage) error {
	return nil
}
```

ExperimentRunner 

### Logging

## Packaging 

## Deployment 


