# Building Routers with Turing SDK

Turing SDK offers users a way to build router configuration incrementally as independent `python` objects: 

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

The SDK classes also double as objects that automatically get created (as properties of Router instances) when users 
receive responses from Turing API containing Router information, allow configuration stored in Turing API to be 
readily reused for new router configuration:

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
