# Introduction
Turing SDK is a Python tool for interacting with Turing API, and complements the existing Turing UI available for 
managing 
router creation, deployment, versioning, etc. 

It aims to assist users in communicating with the API without having 
to manually build their own router configuration (which would normally be written in JSON), and also helps load JSON 
responses by the API into easily manageable Python objects.

By doing so, Turing SDK not only allows you to build your routers in an incremental and configurable manner, it also 
gives you the opportunity to write imperative scripts to automate various router modification and deployment 
processes, hence simplifying your workflow when interacting with Turing API.

## What is Turing SDK?
Turing SDK is entirely written in Python and acts as a wrapper, around the classes automatically generated (by 
[OpenAPI Generator](https://github.com/OpenAPITools/openapi-generator)) from the OpenAPI specs written for Turing API. 

While these auto-generated classes may appear sufficient for users to manipulate with, they expose unnecessary 
complexity to our users, especially the technical details of the underlying implementation of 
router configuration used by Turing API.

Hence, if you're someone who has used Turing/Turing UI and would like more control and power over router management, 
Turing SDK fits perfectly for your needs.

Note that using Turing SDK assumes that you have basic knowledge of what Turing does and how Turing routers operate. 
If you are unsure of these, refer to the Turing UI [docs](https://github.com/gojek/turing/tree/main/docs/how-to) and 
familiarise yourself with them first. A list of useful and important concepts used in Turing can also be found 
[here](https://github.com/gojek/turing/blob/main/docs/concepts.md). 

Note that some functionalities available with the UI are not available with Turing SDK, e.g. creating new projects.

## Features
- Incremental approach in building router configuration

Example:
```python
router_config = RouterConfig(
        environment_name="id-dev",
        name="router-1",
        routes=[
            Route(
                id="model-a",
                endpoint="http://predict-this.io/model-a",
                timeout="100ms"
            ),
            Route(
                id="model-b",
                endpoint="http://predict-this.io/model-b",
                timeout="100ms"
            )
        ],
        rules=None,
        default_route_id="test",
        experiment_engine=ExperimentConfig(
            type="nop",
            config={
                'variables':
                        [
                            {'name': 'order_id', 'field': 'fdsv', 'field_source': 'header'},
                            {'name': 'country_code', 'field': 'dcsd', 'field_source': 'header'},
                            {'name': 'latitude', 'field': 'd', 'field_source': 'header'},
                            {'name': 'longitude', 'field': 'sdSDa', 'field_source': 'header'}
                        ],
                'project_id': 102
            }
        ),
        resource_request=ResourceRequest(
            min_replica=0,
            max_replica=2,
            cpu_request="500m",
            memory_request="512Mi"
        ),
        timeout="100ms",
        log_config=LogConfig(
            result_logger_type=ResultLoggerType.NOP,
            table="abc.dataset.table",
            service_account_secret="not-a-secret"
        ),
        enricher=Enricher(
            image="asia.test.io/model-dev/echo:1.0.2",
            resource_request=ResourceRequest(
                min_replica=0,
                max_replica=2,
                cpu_request="500m",
                memory_request="512Mi"
            ),
            endpoint="/",
            timeout="60ms",
            port=8080,
            env=[
                EnvVar(
                    name="test",
                    value="abc"
                )
            ]
        ),
        ensembler=DockerRouterEnsemblerConfig(
            image="asia.test.io/gods-test/turing-ensembler:0.0.0-build.0",
            resource_request=ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request="500m",
                memory_request="512Mi"
            ),
            endpoint=f"http://localhost:5000/ensembler_endpoint",
            timeout="500ms",
            port=5120,
            env=[],
        )
    )
```
- SDK classes also double as objects that automatically get created (as properties of Router instances) when users 
receive responses from Turing API containing Router information

Example:
```python
# retrieving a router's version using the SDK get_version method
latest_version = my_router.get_version(10)

# extract the RouterConfig object beneath the returned RouterVersion object
latest_config = latest_version.get_config()

# manipulate the extracted RouterConfig object directly (from PR #152)
latest_config.routes[0].timeout = "50ms"

# reuse these RouterConfig objects with other SDK methods
my_router.update(latest_config)
```

## Samples
Samples of how Turing SDK can be used to manage routers can be found 
[here](https://github.com/gojek/turing/tree/main/sdk/samples).