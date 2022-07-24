import os

import turing
import turing.batch
import turing.batch.config
import turing.router.config.router_config

from turing.router.config.route import Route
from turing.router.config.experiment_config import ExperimentConfig
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.log_config import LogConfig, ResultLoggerType
from turing.router.config.enricher import Enricher
from turing.router.config.common.env_var import EnvVar
from turing.router.config.router_ensembler_config import DockerRouterEnsemblerConfig
from turing.router.config.router_config import RouterConfig
from turing.router.config.router_version import RouterStatus


def setup_module():
    turing.set_url(os.getenv('API_BASE_PATH'), use_google_oauth=False)
    turing.set_project(os.getenv('PROJECT_NAME'))


def test_create_router():
    # assert 0 routers are present before creating a new router
    assert len(turing.Router.list()) == 0

    # set up route
    routes = [
        Route(
            id='control',
            endpoint=f'{os.getenv("MOCKSERVER_ENDPOINT")}/control',
            timeout='5ms'
        )
    ]

    # set up experiment config (
    experiment_config = ExperimentConfig(
        type="nop",
        config={
            "client": {
                "id": "1",
                "username": "test",
                "passkey": "test"
            },
            "experiments": [
                {
                    "id": "001",
                    "name": "exp_1"
                }
            ],
            "variables": {
                "experiment_variables": {
                    "001": [
                        {
                            "name": "client_id",
                            "type": "unit",
                            "required": True
                        }
                    ]
                },
                "config": [
                    {
                        "name": "client_id",
                        "required": True,
                        "field": "client.id",
                        "field_source": "payload"
                    }
                ]
            }
        }
    )

    # set up resource request config for the router
    resource_request = ResourceRequest(
        min_replica=1,
        max_replica=1,
        cpu_request="200m",
        memory_request="250Mi"
    )

    # set up log config for the router
    log_config = LogConfig(
        result_logger_type=ResultLoggerType.NOP
    )

    # set up enricher for the router
    enricher = Enricher(
        image=os.getenv("TEST_ECHO_IMAGE"),
        resource_request=ResourceRequest(
            min_replica=1,
            max_replica=1,
            cpu_request="200m",
            memory_request="1Gi"
        ),
        endpoint="anything",
        timeout="2s",
        port=80,
        env=[
            EnvVar(
                name="TEST_ENV",
                value="enricher"
            )
        ]
    )

    # set up ensembler for the router
    ensembler = DockerRouterEnsemblerConfig(
        image=os.getenv("TEST_ECHO_IMAGE"),
        resource_request=ResourceRequest(
            min_replica=2,
            max_replica=2,
            cpu_request="250m",
            memory_request="256Mi"
        ),
        endpoint="anything",
        timeout="3s",
        port=80,
        env=[
            EnvVar(
                name="TEST_ENV",
                value="ensembler"
            )
        ],
    )

    # create the RouterConfig instance
    router_config = RouterConfig(
        environment_name=os.getenv("MODEL_CLUSTER_NAME"),
        name=f'e2e-sdk-experiment-{os.getenv("TEST_ID")}',
        routes=routes,
        rules=[],
        experiment_engine=experiment_config,
        resource_request=resource_request,
        timeout="5s",
        log_config=log_config,
        enricher=enricher,
        ensembler=ensembler
    )

    # create a router using the RouterConfig object
    router = turing.Router.create(router_config)
    print(f"You have created a router with id: {router.id}")

    # Wait for the router to get deployed; note that a router that is PENDING will have None as its router_config
    try:
        router.wait_for_status(RouterStatus.DEPLOYED)
    except TimeoutError:
        raise Exception(f"Turing API is taking too long for router {router.id} to get deployed.")
