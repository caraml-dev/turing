# Developer Guide

## Local Development

### Requirements
- Golang 1.14

## Adding a new Experiment Engine

Experiment Engines have 2 main responsibilities – management of experiments (to cater to requests such as list experiments, check if authorization is required to run experiments, etc.) and running experiments (this generates the Treatment for the given request). The former is used by the Turing app and the latter by the router. Thus, Experiment Engines for Turing must implement two interfaces, one for each type of functionality. 

### Experiment Manager

Experiment runners are required to implement the methods in the `ExperimentManager` interface. Not all methods in the interface may be required by all experiment engines. For example, an experiment engine that does not support authN/authZ need not implement `ListClients`. For this purpose, the `manager` package provides a `NewBaseExperimentManager()` constructor to create a base experiment manager with default implementations, that can be composed into other concrete implementations of the interface.

The primary data types associated with the management of experiments may be visualized with the following UML diagram.

![experiment_manager_data_types](./assets/experiments_data_model.png)

* **Engine** represents the properties of the experiment engine
* An **Experiment** may have an optional client. Experiments may be configured with a list of **Variants**.
* A **Client** may own 1 or more Experiments.
* Experiment **Variables** may be configured at the Client or Experiment level and may be required or optional. Some Experiment Engines distinguish between their type (Unit vs Filter) whereas it may not be important in other systems whose API takes in a single list of key-value pairs.
* The **VariableType** is an enumeration and can be extended with other classifications in the future, if required.

### Experiment Runner

Experiment runners are required to implement the methods in the `ExperimentRunner` interface. This interface contains a single method to retrieve the treatment for a given request.

### Turing API

Todo: Info on registering new plugins, structure of `engines/experiment`.

### Turing UI

The responses from the experiment manager APIs will be used to populate the UI components for Client, Experiment and Variables configuration. For example, the client selection panel will only be displayed if `client_selection_enabled` is set to true in the Experiment Engine properties.