import os
import logging

import requests
import turing
import turing.batch
import turing.batch.config
import turing.router.config.router_config

from turing.router.config.route import Route
from turing.router.config.traffic_rule import (
    TrafficRule,
    HeaderTrafficRuleCondition,
    PayloadTrafficRuleCondition,
)
from turing.router.config.experiment_config import ExperimentConfig
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.log_config import LogConfig, ResultLoggerType
from turing.router.config.common.env_var import EnvVar
from turing.router.config.router_ensembler_config import DockerRouterEnsemblerConfig
from turing.router.config.router_config import RouterConfig
from turing.router.config.router_version import RouterStatus


def test_deploy_router_with_traffic_rules():
    # set up routes
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
        Route(
            id="treatment-b",
            endpoint=f'{os.getenv("MOCKSERVER_ENDPOINT")}/treatment-b',
            timeout="5s",
        ),
    ]

    # set up traffic rules
    rules = [
        TrafficRule(
            name="rule-1",
            conditions=[
                HeaderTrafficRuleCondition(field="X-Region", values=["region-a"])
            ],
            routes=["treatment-a"],
        ),
        TrafficRule(
            name="rule-2",
            conditions=[
                PayloadTrafficRuleCondition(
                    field="service_type.id", values=["service-type-b"]
                )
            ],
            routes=["treatment-b"],
        ),
    ]

    # set up experiment engine
    experiment_config = ExperimentConfig(
        type="nop",
    )

    # set up resource request config for the router
    resource_request = ResourceRequest(
        min_replica=1, max_replica=1, cpu_request="200m", memory_request="250Mi"
    )

    # set up log config for the router
    log_config = LogConfig(result_logger_type=ResultLoggerType.NOP)

    # set up ensembler for the router
    ensembler = DockerRouterEnsemblerConfig(
        image=os.getenv("TEST_ECHO_IMAGE"),
        resource_request=ResourceRequest(
            min_replica=2, max_replica=2, cpu_request="200m", memory_request="256Mi"
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
        rules=rules,
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

    # post single request to turing router that satisfies the first traffic rule
    logging.info(
        "Testing router endpoint with a request that satisfies the first traffic rule..."
    )
    response = requests.post(
        url=router.endpoint,
        headers={
            "Content-Type": "application/json",
            "X-Region": "region-a",
            "X-Mirror-Body": "true",
        },
        json={},
    )
    assert response.status_code == 200
    expected_response = {
        "experiment": {},
        "route_responses": [
            {"data": {"version": "control"}, "is_default": False, "route": "control"},
            {
                "data": {"version": "treatment-a"},
                "is_default": False,
                "route": "treatment-a",
            },
        ],
    }
    actual_response = response.json()["response"]
    actual_response["route_responses"] = sorted(
        actual_response["route_responses"], key=lambda x: x["data"]["version"]
    )
    assert actual_response["experiment"] == expected_response["experiment"]
    assert actual_response["route_responses"] == expected_response["route_responses"]

    # post single request to turing router that satisfies the second traffic rule
    logging.info(
        "Testing router endpoint with a request that satisfies the second traffic rule..."
    )
    response = requests.post(
        url=router.endpoint,
        headers={
            "Content-Type": "application/json",
            "X-Mirror-Body": "true",
        },
        json={"service_type": {"id": "service-type-b"}},
    )
    assert response.status_code == 200
    expected_response = {
        "experiment": {},
        "route_responses": [
            {"data": {"version": "control"}, "is_default": False, "route": "control"},
            {
                "data": {"version": "treatment-b"},
                "is_default": False,
                "route": "treatment-b",
            },
        ],
    }
    actual_response = response.json()["response"]
    actual_response["route_responses"] = sorted(
        actual_response["route_responses"], key=lambda x: x["data"]["version"]
    )
    assert actual_response["experiment"] == expected_response["experiment"]
    assert actual_response["route_responses"] == expected_response["route_responses"]

    # post single request to turing router that satisfies neither traffic rule
    logging.info(
        "Testing router endpoint with a request that satisfies neither traffic rule..."
    )
    response = requests.post(
        url=router.endpoint,
        headers={
            "Content-Type": "application/json",
            "X-Mirror-Body": "true",
        },
        json={"service_type": {"id": "service-type-c"}},
    )
    assert response.status_code == 200
    expected_response = {
        "experiment": {},
        "route_responses": [
            {"data": {"version": "control"}, "is_default": False, "route": "control"},
        ],
    }
    assert response.json()["response"] == expected_response
