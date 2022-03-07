# Create a Router

The behaviour of a Turing router is defined by its configuration. To create a router, it is necessary to specify its 
configuration, which itself is made up of various components. Some of these components must be defined, while others 
such as the use of experiment engines, enrichers (pre-processors) or ensemblers (post-processors) are optional.  

Hence to build a router using Turing SDK, you would need to incrementally define these components and build a 
`RouterConfig` object (you can find more information on how these individual components need to be built in the 
XXX section).

Using the example shown in the `README` of the Turing SDK documentation main page, you need to construct a 
`RouterConfig` instance by specifying the various components as arguments:

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

Once this `RouterConfig` instance is defined, you can simply create a new Router by running:

```python
# 1. Create a new router using the RouterConfig object
new_router = turing.Router.create(router_config)
```

The return value would be a `Router` object representing the router that has been created if Turing API has 
created it successfully. Note that this does not necessarily mean that the router has been succesfully *deployed*.

