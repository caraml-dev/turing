# Introduction

## What is Turing?

Turing is a fast, scalable and extensible system that can be used to design, deploy and evaluate ML experiments in production. It takes care of the core Engineering aspects of experimentation such as traffic routing, outcome logging, system monitoring, etc. and is designed to work with pluggable Experiment Engines, pre and post processors. It is backed by existing systems like Merlin for model endpoints and Litmus / XP (WIP) for experiment configuration.

## Concepts
**Project**: Holds all MLP resources that belong to a specific team such as service accounts, Merlin models, etc.

**Router**: The router is the nucleus of the Turing system. It is responsible for coordinating the traffic routing to multiple model endpoints, invoking the pre and post processors, incorporating the response from the Experiment engine and logging of these responses.  

**Request**: Incoming message from the client to the Turing system.

**Response**: The Turing workflow involves the pre-processor (Enricher), the model endpoints, the Experiment engine and the post-processor (Ensembler), some of which are optional. Each component creates a response which becomes the request to the next component in the workflow (refer to How It Works below). In general, the Response refers to the final response from the Turing system, after passing through all stages.

**Route**: Model endpoint which may be a Merlin model or any arbitrary URL that can be reached from the Turing infrastructure.

**Experiment**: An application of rules, filters and configurations that determine how the traffic is routed and responses are combined to create the final response to the Turing request and enables evaluation of different models and parameters.

**Treatment**: The set of configurations and actions to be applied to the current request which results in an outcome that can be evaluated.

**Unit**: Smallest entity that can receive different treatments.

**Rule**: Conditions determining which treatment to apply to a specific unit.

**Enricher**: An optional service to perform arbitrary transformations on the incoming request or supplementing the request with data from external sources.

**Ensembler**: An optional external service that accepts responses from the model endpoints altogether with the experiment configuration and responds back to the Turing router with a final response. Exploration policies or combining responses from multiple models into one can be implemented here.

## Features
* Low-latency, high-throughput traffic routing to an unlimited number of ML models.

* Experimentation rules based on incoming requests to determine the treatment to be applied. The experiment engines currently supported are Litmus and XP (WIP).

* Feature enrichment of incoming requests through Feast (planned) and arbitrary pre-processors.

* Dynamic ensembling of models for each treatment. This could be selecting one of the models' response, a custom ensembling of responses from two or more models or any other arbitrary post-processing.

* Reliable and safe fallbacks in case of timeouts.

* Simple response and outcome tracking.

## How It Works

1. The Turing router receives incoming requests from the client.

2. Enrichment of request with features from external sources can be done if required by the Enricher. 

3. The unit ID is extracted and passed to a Experiment Engine to determine the treatment. Simultaneously, the request is forwarded to all model endpoints via the configured routes.

4. The Ensembler is called with the original Turing request, the implementation of exploration policies from the experiment engine response and the model responses.

5. A tracking ID is appended to the ensembled response and it is logged (together with the individual model responses and the original request) as it is returned to the client.

6. The client will then be able to log the outcome with the tracking ID.

## When To Use Turing
* You need to send traffic to multiple model endpoints

* You need to ensemble the resulting response based on the experiment configuration in Litmus / XP

* You want request-response pairs to be logged.

* You require out of the box monitoring of response times, errors and other router standard metrics.
