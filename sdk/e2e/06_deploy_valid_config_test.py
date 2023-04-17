import os
import logging

import turing

from turing.router.config.router_version import RouterStatus


def test_deploy_router_valid_config():
    # get the existing router that has been created in 01_create_router_test.py
    logging.info("Retrieving router...")
    router = turing.Router.get(1)
    assert router is not None

    # deploy the router
    logging.info("Deploying router...")
    response = router.deploy()
    assert response["router_id"] == 1
    assert response["version"] == 1

    # check that the status of the router is pending
    try:
        router.wait_for_status(RouterStatus.PENDING)
    except TimeoutError:
        raise Exception(
            f"Turing API is taking too long for router {router.id} to start getting deployed."
        )

    # wait for the router to get deployed
    try:
        router.wait_for_status(RouterStatus.DEPLOYED)
    except TimeoutError:
        raise Exception(
            f"Turing API is taking too long for router {router.id} to get deployed."
        )

    # get the router again
    logging.info("Retrieving router...")
    router = turing.Router.get(1)
    assert router is not None
    assert router.version == 1
    assert (
        router.endpoint
        == f'http://{router.name}-turing-router.{os.getenv("PROJECT_NAME")}.{os.getenv("KSERVICE_DOMAIN")}/v1/predict'
    )
    assert router.status == RouterStatus.DEPLOYED
