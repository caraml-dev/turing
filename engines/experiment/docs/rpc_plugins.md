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

## Packaging 

## Deployment 


