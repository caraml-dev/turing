import turing
import turing.batch
import turing.batch.config
import turing.router.config.router_config
from turing.router.config.route import Route
from turing.router.config.router_config import RouterConfig
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.log_config import LogConfig, ResultLoggerType
from turing.router.config.traffic_rule import TrafficRule, HeaderTrafficRuleCondition, PayloadTrafficRuleCondition
from turing.router.config.enricher import Enricher
from turing.router.config.router_ensembler_config import DockerRouterEnsemblerConfig
from turing.router.config.common.env_var import EnvVar
from turing.router.config.experiment_config import ExperimentConfig


def main(turing_api: str, project: str):
    # Initialize Turing client
    turing.set_url(turing_api)
    turing.set_project(project)

    # Build a router config in order to create a router
    # # Create some routes
    routes = [
        Route(
            id='meow',
            endpoint='http://fox-says.meow',
            timeout='20ms'
        ),
        Route(
            id='woof',
            endpoint='http://fox-says.woof',
            timeout='20ms'
        ),
        Route(
            id='baaa',
            endpoint='http://fox-says.baa',
            timeout='20ms'
        ),
        Route(
            id='oink',
            endpoint='http://fox-says.oink',
            timeout='20ms'
        ),
        Route(
            id='ring-ding-ding',
            endpoint='http://fox-says.ring-ding-ding',
            timeout='20ms'
        )
    ]

    # # Create some traffic rules
    rules = [
        TrafficRule(
            conditions=[
                HeaderTrafficRuleCondition(
                    field='cat',
                    values=['name']
                )
            ],
            routes=[
                'meow'
            ]
        ),
        TrafficRule(
            conditions=[
                HeaderTrafficRuleCondition(
                    field='dog',
                    values=['name']
                )
            ],
            routes=[
                'woof'
            ]
        ),
        TrafficRule(
            conditions=[
                HeaderTrafficRuleCondition(
                    field='sheep',
                    values=['name']
                )
            ],
            routes=[
                'baaa'
            ]
        ),
        TrafficRule(
            conditions=[
                HeaderTrafficRuleCondition(
                    field='pig',
                    values=['name']
                )
            ],
            routes=[
                'oink'
            ]
        ),
        TrafficRule(
            conditions=[
                PayloadTrafficRuleCondition(
                    field='sus',
                    values=['body']
                )
            ],
            routes=[
                'meow',
                'woof',
                'baaa',
                'oink',
                'ring-ding-ding'
            ]
        )
    ]

    # # Create an experiment config (
    experiment_config = ExperimentConfig(
        type="xp",
        config={
            'variables':
                [
                    {'name': 'farm_id', 'field': 'farm_id', 'field_source': 'header'},
                    {'name': 'country_code', 'field': 'country', 'field_source': 'header'},
                    {'name': 'latitude', 'field': 'farm_lat', 'field_source': 'header'},
                    {'name': 'longitude', 'field': 'farm_long', 'field_source': 'header'}
                ],
            'project_id': 102
        }
    )

    # # Create a resource request config for the router
    resource_request = ResourceRequest(
        min_replica=0,
        max_replica=2,
        cpu_request="500m",
        memory_request="512Mi"
    )

    # # Create a log config for the router
    log_config = LogConfig(
        result_logger_type=ResultLoggerType.NOP
    )

    # # Create an enricher for the router
    enricher = Enricher(
        image="asia.gcr.io/gods-dev/echo:1.0.2",
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
                name="humans",
                value="farmer-joe"
            )
        ]
    )

    # # Create an ensembler for the router
    ensembler = DockerRouterEnsemblerConfig(
        id=1,
        image="asia.gcr.io/gods-dev/echo:1.0.2",
        resource_request=ResourceRequest(
            min_replica=1,
            max_replica=3,
            cpu_request="500m",
            memory_request="512Mi"
        ),
        endpoint="/echo",
        timeout="60ms",
        port=8080,
        env=[],
    )

    # # Create the RouterConfig instance
    router_config = RouterConfig(
        environment_name="id-dev",
        name="what-does-the-fox-say-1",
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

    # 1. Create a new router using the RouterConfig object
    new_router = turing.Router.create(router_config)

    # 2. List all routers
    routers = turing.Router.list()
    for r in routers:
        if r.name == new_router.name:
            my_router_id = r.id
        print(r)

    # 3. Get the router you just created using the router_id obtained
    my_router = turing.Router.get(my_router_id)

    # Access the router config from the returned Router object directly
    my_router_config = my_router.config

    # Modify something in the router config
    my_router_config.routes.append(
        Route(
            id='fee-fi-fo-fum',
            endpoint='http://fox-says.fee-fi-fo-fum',
            timeout='20ms'
        )
    )

    # 4. Update the router with the new router config
    my_router.update(my_router_config)

    # 5. List all the router config versions of your router
    my_router_versions = my_router.list_versions()
    for ver in my_router_versions:
        if ver.status == 'undeployed':
            first_ver_no = ver.version
        if ver.status == "deployed":
            latest_ver_no = ver.version
        print(ver)

    # 6. Deploy a specific router config version (the first one we created)
    my_router.deploy_version(first_ver_no)

    # 7. Undeploy the current active router configuration
    my_router.undeploy()

    # 8. Deploy the router's *current* configuration (notice how it still deploys the *first* version)
    my_router.deploy()

    # # Undeploy the router
    my_router.undeploy()

    # 9. Get a specific router version of the router
    my_router_ver = my_router.get_version(first_ver_no)
    print(my_router_ver)

    # 10. Delete a specific router version of the router
    my_router.delete_version(latest_ver_no)

    # 11. Get all deployment events associated with this router
    events = my_router.get_events()
    for e in events:
        print(e)

    # 12. Delete this router (using its router_id)
    turing.Router.delete(my_router_id)
    # # Check if the router still exists
    for r in turing.Router.list():
        if r.name == my_router_config.name:
            print("Oh my, this router still exists!")
            break


if __name__ == '__main__':
    import fire
    fire.Fire(main)
