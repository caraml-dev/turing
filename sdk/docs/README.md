# Introduction
The Turing SDK is a Python tool for interacting with the Turing API, and complements the existing Turing UI available 
for managing router creation, deployment, versioning, etc.

It not only allows you to build your routers in an incremental and configurable manner, it 
also gives you the opportunity to write imperative scripts to automate various router modification and deployment 
processes, hence simplifying your workflow when interacting with Turing API.

## What is the Turing SDK?
The Turing SDK is entirely written in Python and acts as a wrapper, around the classes automatically generated (by 
[OpenAPI Generator](https://github.com/OpenAPITools/openapi-generator)) from the OpenAPI specs written for the Turing 
API. These generated classes in turn act as an intermediary between raw JSON objects that are passed in HTTP 
requests/responses made to/received from the Turing API.

![Turing SDK Classes](https://github.com/caraml-dev/turing/blob/main/sdk/docs/assets/turing-sdk-classes.png?raw=true)

If you're someone who has used Turing/the Turing UI and would like more control and power over router 
management, the Turing SDK fits perfectly for your needs.

Note that using the Turing SDK assumes that you have basic knowledge of what Turing does and how Turing routers 
operate. If you are unsure of these, refer to the Turing UI [docs](https://github.com/caraml-dev/turing/tree/main/docs/how-to) and 
familiarise yourself with them first. A list of useful and important concepts used in Turing can also be found 
[here](https://github.com/caraml-dev/turing/blob/main/docs/concepts.md). 

Note that some functionalities available with the UI are not available with the Turing SDK, e.g. creating new projects.

## Samples
Samples of how the Turing SDK can be used to manage routers can be found 
[here](https://github.com/caraml-dev/turing/tree/main/sdk/samples).