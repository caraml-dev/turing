import os
import turing

from turing.router.config.experiment_config import ExperimentConfig
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.enricher import Enricher
from turing.router.config.common.env_var import EnvVar
from turing.router.config.router_version import RouterStatus


def test_update_router_invalid_config():
    # get existing router that has been created in 01_create_router_test.py
    router = turing.Router.get(1)
    assert router is not None

    # get the router config from the deployed router
    new_router_config = router.config

    # set up the new experiment config
    new_router_config.experiment_config = ExperimentConfig(
        type="nop",
    )

    # set up the new senricher for the router
    new_router_config.enricher = Enricher(
        image=os.getenv("TEST_ECHO_IMAGE"),
        resource_request=ResourceRequest(
            min_replica=1,
            max_replica=1,
            cpu_request="10",
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

    # update router
    router.update(new_router_config)

    # get router version 2
    router_ver_2 = router.get_version(2)
    assert router_ver_2.status == RouterStatus.FAILED
