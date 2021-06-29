# Concepts

**Project**: Holds all [MLP](https://github.com/gojek/mlp) resources that belong to a specific team such as service accounts, Merlin models, etc.

**Router**: The router is the nucleus of the Turing system. It is responsible for coordinating the traffic routing to multiple model endpoints, invoking the pre and post processors, incorporating the response from the Experiment engine and logging of these responses.  

**Request**: Incoming message from the client to the Turing system.

**Response**: The Turing workflow involves the pre-processor (Enricher), the model endpoints, the Experiment engine and the post-processor (Ensembler), some of which are optional. Each component creates a response which becomes the request to the next component in the workflow. In general, the Response refers to the final response from the Turing system, after passing through all stages.

**Route**: Model endpoint which may be a Merlin model or any arbitrary URL that can be reached from the Turing infrastructure.

**Experiment**: An application of rules, filters and configurations that determine how the traffic is routed and responses are combined to create the final response to the Turing request and enables evaluation of different models and parameters.

**Treatment**: The set of configurations and actions to be applied to the current request which results in an outcome that can be evaluated.

**Unit**: Smallest entity that can receive different treatments.

**Rule**: Conditions determining which treatment to apply to a specific unit.

**Enricher**: An optional service to perform arbitrary transformations on the incoming request or supplementing the request with data from external sources.

**Ensembler**: An optional external service that accepts responses from the model endpoints altogether with the experiment configuration and responds back to the Turing router with a final response. Exploration policies or combining responses from multiple models into one can be implemented here.
