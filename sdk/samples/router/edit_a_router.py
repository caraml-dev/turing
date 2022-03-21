import turing
import turing.batch
import turing.batch.config
import turing.router.config.router_config
from turing.router.config.route import Route
from turing.router.config.router_config import RouterConfig
from turing.router.config.router_version import RouterStatus
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.log_config import LogConfig, ResultLoggerType
from turing.router.config.traffic_rule import TrafficRule, HeaderTrafficRuleCondition
from turing.router.config.enricher import Enricher
from turing.router.config.router_ensembler_config import DockerRouterEnsemblerConfig
from turing.router.config.common.env_var import EnvVar
from turing.router.config.experiment_config import ExperimentConfig


def main(turing_api: str, project: str):
    # Initialize Turing client
    turing.set_url(turing_api)
    turing.set_project(project)

    # Build a router for the sake of showing how you can retrieve one from the API
    # Create some routes
    routes = [
        Route(
            id='route-1',
            endpoint='http://paths.co/route-1',
            timeout='20ms'
        ),
        Route(
            id='route-2',
            endpoint='http://paths.co/route-2',
            timeout='20ms'
        )
    ]

    # Create some traffic rules
    rules = [
        TrafficRule(
            conditions=[
                HeaderTrafficRuleCondition(
                    field='turns',
                    values=['left']
                )
            ],
            routes=[
                'route-1'
            ]
        ),
        TrafficRule(
            conditions=[
                HeaderTrafficRuleCondition(
                    field='turns',
                    values=['right']
                )
            ],
            routes=[
                'route-2'
            ]
        )
    ]

    # Create an experiment config (
    experiment_config = ExperimentConfig(
        type="test-exp",
        config={
            'variables':
                [
                    {'name': 'latitude', 'field': 'farm_lat', 'field_source': 'header'},
                    {'name': 'longitude', 'field': 'farm_long', 'field_source': 'header'}
                ],
            'project_id': 102
        }
    )

    # Create a resource request config for the router
    resource_request = ResourceRequest(
        min_replica=0,
        max_replica=2,
        cpu_request="500m",
        memory_request="512Mi"
    )

    # Create a log config for the router
    log_config = LogConfig(
        result_logger_type=ResultLoggerType.NOP
    )

    # Create an enricher for the router
    enricher = Enricher(
        image="ealen/echo-server:0.5.1",
        resource_request=ResourceRequest(
            min_replica=0,
            max_replica=2,
            cpu_request="500m",
            memory_request="512Mi"
        ),
        endpoint="/",
        timeout="60ms",
        port=3000,
        env=[
            EnvVar(
                name="NODES",
                value="2"
            )
        ]
    )

    # Create an ensembler for the router
    ensembler = DockerRouterEnsemblerConfig(
        image="ealen/echo-server:0.5.1",
        resource_request=ResourceRequest(
            min_replica=1,
            max_replica=3,
            cpu_request="500m",
            memory_request="512Mi"
        ),
        endpoint="/echo",
        timeout="60ms",
        port=3000,
        env=[],
    )

    # Create the RouterConfig instance
    router_config = RouterConfig(
        environment_name="id-dev",
        name="my-router-1",
        routes=routes,
        rules=rules,
        default_route_id="test",
        experiment_engine=experiment_config,
        resource_request=resource_request,
        timeout="100ms",
        log_config=log_config,
        enricher=enricher,
        ensembler=ensembler
    )

    # Create a router using the RouterConfig object
    router = turing.Router.create(router_config)
    print(f"You have created a router with id: {router.id}")

    # Wait for the router to get deployed; note that a router that is PENDING will have None as its router_config
    try:
        router.wait_for_status(RouterStatus.DEPLOYED)
    except TimeoutError:
        raise Exception(f"Turing API is taking too long for router {router.id} to get deployed.")

    # Imagine we only have the router's id, and would like to retrieve it
    router_1 = turing.Router.get(router.id)

    # Now we'd like to update the config of router_1, but with some fields modified
    # Reminder: When trying to replicate configuration from an existing router, always retrieve the underlying
    # `RouterConfig` from the `Router` instance by accessing its `config` attribute.

    # Get the router config from router_1
    new_router_config_to_deploy = router_1.config

    # Make your desired changes to the config
    # Note that router_config.enricher.env is a regular Python list; so you can use methods such as append or extend
    new_router_config_to_deploy.enricher.env.append(
        EnvVar(
            name="WORKERS",
            value="2"
        )
    )

    new_router_config_to_deploy.resource_request.max_replica = 5

    # When editing a router, you can either 1. UPDATE the router, which would create a new router version and deploy it
    # immediately, or 2. SAVE the router version, which would only create a new router version without deploying it

    # 1. When you UPDATE a router, Turing will save the new version and attempt to deploy it immediately
    router_1.update(new_router_config_to_deploy)

    # Notice that the latest router version is pending deployment while the current router version is still active
    versions = router.list_versions()
    for v in versions:
        print(v)

    # 2. When you SAVE a router, Turing will save the new version, but not deploy it.
    new_router_config_to_save = router_config
    new_router_config_to_save.resource_request.min_replica = 0

    router_1.save_version(new_router_config_to_deploy)

    # Notice that the latest router version is undeployed (Turing has created the new version without deploying it)
    versions = router.list_versions()
    for v in versions:
        print(v)


if __name__ == '__main__':
    import fire
    fire.Fire(main)
