import logging

import turing

from turing.router.config.router_version import RouterStatus


def test_deploy_router_invalid_config():
    # get the existing router that has been created in 01_create_router_test.py
    logging.info("Retrieving router...")
    router = turing.Router.get(1)
    assert router is not None

    # deploy router version that corresponds to a failed deployment
    logging.info("Deploying router version...")
    response = router.deploy_version(2)
    assert response["router_id"] == router.id
    assert response["version"] == 2

    # wait for the router version deployment to fail
    try:
        router.wait_for_version_status(RouterStatus.FAILED, 2)
    except TimeoutError:
        raise Exception(
            f"Turing API is taking too long for router {router.id} with version 2 to fail."
        )

    # get the failed router version
    logging.info("Retrieving router version...")
    router_version = router.get_version(2)
    assert router_version.status == RouterStatus.FAILED

    # get the router again
    logging.info("Retrieving router...")
    router = turing.Router.get(1)
    assert router is not None
    assert router.status == RouterStatus.UNDEPLOYED
    assert router.version == 1
    assert router.get_version(1).status == RouterStatus.UNDEPLOYED
