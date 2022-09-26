import os
import logging

import requests
import turing
import turing.batch
import turing.batch.config
import turing.router.config.router_config

from turing.router.config.route import Route
from turing.router.config.experiment_config import ExperimentConfig
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.log_config import LogConfig, ResultLoggerType
from turing.router.config.router_ensembler_config import StandardRouterEnsemblerConfig
from turing.router.config.router_config import RouterConfig
from turing.router.config.router_version import RouterStatus


def test_deploy_router_with_std_ensembler():
    # set up route
    routes = [
        Route(
            id="control",
            endpoint=f'{os.getenv("MOCKSERVER_ENDPOINT")}/control',
            timeout="5s",
        ),
        Route(
            id="treatment-a",
            endpoint=f'{os.getenv("MOCKSERVER_ENDPOINT")}/treatment-a',
            timeout="5s",
        ),
    ]

    # set up experiment config
    experiment_config = ExperimentConfig(
        type="proprietary",
        config={
            "client": {"id": "1", "username": "test", "passkey": "test"},
            "experiments": [{"id": "001", "name": "exp_1"}],
            "variables": {
                "experiment_variables": {
                    "001": [{"name": "client_id", "type": "unit", "required": True}]
                },
                "config": [
                    {
                        "name": "client_id",
                        "required": True,
                        "field": "client.id",
                        "field_source": "payload",
                    }
                ],
            },
        },
    )

    # set up resource request config for the router
    resource_request = ResourceRequest(
        min_replica=1, max_replica=1, cpu_request="200m", memory_request="250Mi"
    )

    # set up log config for the router
    log_config = LogConfig(result_logger_type=ResultLoggerType.NOP)

    # set up standard ensembler
    ensembler = StandardRouterEnsemblerConfig(
        route_name_path="route_name", fallback_response_route_id="control"
    )

    # create the RouterConfig instance
    router_config = RouterConfig(
        environment_name=os.getenv("MODEL_CLUSTER_NAME"),
        name=f'e2e-sdk-std-ensembler-test-{os.getenv("TEST_ID")}',
        routes=routes,
        rules=[],
        experiment_engine=experiment_config,
        resource_request=resource_request,
        timeout="5s",
        log_config=log_config,
        ensembler=ensembler,
    )

    # create a router using the RouterConfig object
    router = turing.Router.create(router_config)
    assert router.status == RouterStatus.PENDING

    # wait for the router to get deployed
    try:
        router.wait_for_status(RouterStatus.DEPLOYED)
    except TimeoutError:
        raise Exception(
            f"Turing API is taking too long for router {router.id} to get deployed."
        )
    assert router.status == RouterStatus.DEPLOYED

    # get router with id 1
    retrieved_router = turing.Router.get(router.id)
    assert retrieved_router.version == 1
    assert retrieved_router.status == RouterStatus.DEPLOYED
    assert (
        retrieved_router.endpoint
        == f'http://{retrieved_router.name}-turing-router.{os.getenv("PROJECT_NAME")}.{os.getenv("KSERVICE_DOMAIN")}/v1/predict'
    )

    # get router version with id 1
    router_version_1 = retrieved_router.get_version(1)
    assert router_version_1.status == RouterStatus.DEPLOYED

    # post single request to turing router
    logging.info("Testing router endpoint...")
    response = requests.post(
        url=router.endpoint,
        headers={
            "Content-Type": "application/json",
            "X-Mirror-Body": "true",
        },
        json={"client": {"id": 4}},
    )
    assert response.status_code == 200
    expected_response = {"version": "treatment-a"}
    assert response.json() == expected_response
