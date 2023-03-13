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
from turing.router.config.enricher import Enricher
from turing.router.config.common.env_var import EnvVar
from turing.router.config.router_ensembler_config import DockerRouterEnsemblerConfig
from turing.router.config.router_config import RouterConfig, Protocol
from turing.router.config.router_version import RouterStatus


def test_create_router():
    # assert 0 routers are present before creating a new router
    assert len(turing.Router.list()) == 0

    # set up route
    routes = [
        Route(
            id="control",
            endpoint=f'{os.getenv("MOCKSERVER_HTTP_ENDPOINT")}/control',
            timeout="5s",
        )
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
        min_replica=1, max_replica=1, cpu_request="100m", memory_request="250Mi"
    )

    # set up log config for the router
    log_config = LogConfig(result_logger_type=ResultLoggerType.NOP)

    # set up enricher for the router
    enricher = Enricher(
        image=os.getenv("TEST_ECHO_IMAGE"),
        resource_request=ResourceRequest(
            min_replica=1, max_replica=1, cpu_request="100m", memory_request="1Gi"
        ),
        endpoint="anything",
        timeout="2s",
        port=80,
        env=[EnvVar(name="TEST_ENV", value="enricher")],
    )

    # set up ensembler for the router
    ensembler = DockerRouterEnsemblerConfig(
        image=os.getenv("TEST_ECHO_IMAGE"),
        resource_request=ResourceRequest(
            min_replica=2, max_replica=2, cpu_request="100m", memory_request="256Mi"
        ),
        endpoint="anything",
        timeout="3s",
        port=80,
        env=[EnvVar(name="TEST_ENV", value="ensembler")],
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
    retrieved_router = turing.Router.get(1)
    assert retrieved_router.version == 1
    assert retrieved_router.status == RouterStatus.DEPLOYED
    assert retrieved_router.config.protocol == Protocol.HTTP
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
    expected_response = {
        "experiment": {
            "configuration": {"foo": "foo", "route_name": "control"},
        },
        "route_responses": [
            {"data": {"version": "control"}, "is_default": False, "route": "control"}
        ],
    }
    assert response.json()["response"] == expected_response

    # post batch request to turing router
    logging.info("Testing router batch endpoint...")
    response = requests.post(
        url=router.endpoint.replace("/predict", "/batch_predict"),
        headers={
            "Content-Type": "application/json",
            "X-Mirror-Body": "true",
        },
        json=[{"client": {"id": 4}}, {"client": {"id": 7}}],
    )
    assert response.status_code == 200
    expected_response = [
        {
            "code": 200,
            "data": {
                "request": {"client": {"id": 4}},
                "response": {
                    "experiment": {
                        "configuration": {"foo": "foo", "route_name": "control"},
                    },
                    "route_responses": [
                        {
                            "data": {"version": "control"},
                            "is_default": False,
                            "route": "control",
                        }
                    ],
                },
            },
        },
        {
            "code": 200,
            "data": {
                "request": {"client": {"id": 7}},
                "response": {
                    "experiment": {
                        "configuration": {"bar": "bar", "route_name": "treatment-a"},
                    },
                    "route_responses": [
                        {
                            "data": {"version": "control"},
                            "is_default": False,
                            "route": "control",
                        }
                    ],
                },
            },
        },
    ]
    assert response.json() == expected_response

    # test endpoint for router logs
    logging.info("Testing endpoint for router logs...")
    router_component_types = ["router", "ensembler", "enricher"]
    base_url = f'{os.getenv("API_BASE_PATH")}/projects/{os.getenv("PROJECT_ID")}/routers/{router.id}/logs'
    for component in router_component_types:
        response = requests.get(
            url=base_url + "?component_type=" + component,
        )
    assert response.status_code == 200
    assert len(response.content) > 0
