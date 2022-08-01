import logging

import turing

from turing.router.config.router_version import RouterStatus


def test_delete_router():
    # get the existing router that has been created in 01_create_router_test.py
    logging.info("Retrieving router...")
    router = turing.Router.get(1)
    assert router is not None

    # undeploy router
    logging.info("Undeploying router...")
    response = router.undeploy()
    assert response['router_id'] == router.id

    # wait for the router to get undeployed
    try:
        router.wait_for_status(RouterStatus.UNDEPLOYED)
    except TimeoutError:
        raise Exception(f"Turing API is taking too long for router {router.id} to get undeployed.")

    # delete router
    logging.info("Deleting router...")
    deleted_router_id = turing.Router.delete(router.id)
    assert deleted_router_id == router.id

    # check that the router no longer exists
    for r in turing.Router.list():
        assert r.name != router.name
