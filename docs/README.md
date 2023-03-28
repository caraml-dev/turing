# Introduction

## What is Turing?

Turing is a fast, scalable and extensible system that can be used to design, deploy and evaluate ML experiments in production. It takes care of the core Engineering aspects of experimentation such as traffic routing, outcome logging, system monitoring, etc. and is designed to work with pluggable Experiment Engines, pre and post processors. It is backed by existing systems like [Merlin](https://github.com/caraml-dev/merlin) for model endpoints.

## Features
* Low-latency, high-throughput traffic routing to an unlimited number of ML models.

* Experimentation rules based on incoming requests to determine the treatment to be applied. The experiment engines currently supported are closed source for now (we are working on this!).

* Feature enrichment of incoming requests through [Feast](https://github.com/feast-dev/feast) (planned) and arbitrary pre-processors.

* Dynamic ensembling of models for each treatment. This could be selecting one of the models' response, custom ensembling of responses from two or more models or any other arbitrary post-processing.

* Reliable and safe fallbacks in case of timeouts.

* Simple response and outcome tracking.

## How It Works

1. The Turing router receives incoming requests from the client.

2. Enrichment of request with features from external sources can be done if required by the Enricher. 

3. The unit ID is extracted and passed to an Experiment Engine to determine the treatment. Simultaneously, the request is forwarded to all model endpoints via the configured routes.

4. The Ensembler is called with the original Turing request, the implementation of exploration policies from the experiment engine response and the model responses.

5. A tracking ID is appended to the ensembled response and it is logged (together with the individual model responses and the original request) and it is returned to the client.

6. The client will then be able to log the outcome with the tracking ID.

## When To Use Turing
* You need to send traffic to multiple model endpoints

* You need to ensemble the resulting response based on the experiment configuration.

* You want request-response pairs to be logged.
