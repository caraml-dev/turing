import os
import logging

import requests
import turing

from turing.router.config.experiment_config import ExperimentConfig
from turing.router.config.autoscaling_policy import AutoscalingPolicy, AutoscalingMetric
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.enricher import Enricher
from turing.router.config.common.env_var import EnvVar
from turing.router.config.router_ensembler_config import DockerRouterEnsemblerConfig
from turing.router.config.router_version import RouterStatus


def test_create_router_version():
    # get the existing router that has been created in 01_create_router_test.py
    logging.info("Retrieving router...")
    router = turing.Router.get(1)
    assert router is not None

    # get the router config from the deployed router
    new_router_config = router.config

    # Set autoscaling policy
    new_router_config.autoscaling_policy = AutoscalingPolicy(
        metric=AutoscalingMetric.RPS, target="100"
    )

    # set up the new experiment config
    new_router_config.experiment_config = ExperimentConfig(
        type="nop",
    )

    # set up the new senricher for the router
    new_router_config.enricher = Enricher(
        image=os.getenv("TEST_ECHO_IMAGE"),
        resource_request=ResourceRequest(
            min_replica=1, max_replica=1, cpu_request="10", memory_request="1Gi"
        ),
        autoscaling_policy=AutoscalingPolicy(metric=AutoscalingMetric.CPU, target="80"),
        endpoint="anything",
        timeout="2s",
        port=80,
        env=[EnvVar(name="TEST_ENV", value="enricher")],
    )

    # set up the new ensembler for the router
    new_router_config.ensembler = DockerRouterEnsemblerConfig(
        image=os.getenv("TEST_ECHO_IMAGE"),
        resource_request=ResourceRequest(
            min_replica=2, max_replica=2, cpu_request="200m", memory_request="256Mi"
        ),
        autoscaling_policy=AutoscalingPolicy(
            metric=AutoscalingMetric.MEMORY, target="220"
        ),
        endpoint="anything",
        timeout="3s",
        port=80,
        env=[EnvVar(name="TEST_ENV", value="ensembler")],
    )

    # update router
    logging.info("Updating router with new config...")
    router.create_version(new_router_config)

    # get router version 3
    logging.info("Retrieving new router version...")
    router_ver_3 = router.get_version(3)
    assert router_ver_3.status == RouterStatus.UNDEPLOYED

    # ensure router is not yet updated (i.e. router version remains as 1)
    logging.info(
        "Ensuring the router is not yet updated (i.e. the version number remains as 1)..."
    )
    router = turing.Router.get(1)
    assert router.version == 1

    # test router endpoint by posting a single request
    assert (
        router.endpoint
        == f'http://{router.name}-turing-router.{os.getenv("PROJECT_NAME")}.{os.getenv("KSERVICE_DOMAIN")}/v1/predict'
    )
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
