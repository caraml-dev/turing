import os
import logging

from caraml.upi.v1 import upi_pb2_grpc, upi_pb2
from caraml.upi.v1 import table_pb2, type_pb2
import grpc

import turing
import turing.batch
import turing.batch.config
import turing.router.config.router_config

from turing.router.config.route import Route, RouteProtocol
from turing.router.config.experiment_config import ExperimentConfig
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.log_config import LogConfig, ResultLoggerType
from turing.router.config.router_config import RouterConfig
from turing.router.config.router_version import RouterStatus


def test_deploy_router_upi():
    # set up route
    routes = [
        Route(
            id="control",
            endpoint=f'{os.getenv("MOCKSERVER_UPI_ENDPOINT")}',
            protocol=RouteProtocol.GRPC,
            service_method="caraml.upi.v1.UniversalPredictionService/PredictValues",
            timeout="5s",
        ),
        Route(
            id="treatment-a",
            endpoint=f'{os.getenv("MOCKSERVER_UPI_ENDPOINT")}',
            protocol=RouteProtocol.GRPC,
            service_method="caraml.upi.v1.UniversalPredictionService/PredictValues",
            timeout="5s",
        ),
    ]

    # set up resource request config for the router
    resource_request = ResourceRequest(
        min_replica=1, max_replica=1, cpu_request="200m", memory_request="250Mi"
    )

    # set up log config for the router
    log_config = LogConfig(result_logger_type=ResultLoggerType.NOP)

    # create the RouterConfig instance
    router_config = RouterConfig(
        environment_name=os.getenv("MODEL_CLUSTER_NAME"),
        name=f'e2e-sdk-upi-test-{os.getenv("TEST_ID")}',
        routes=routes,
        rules=[],
        experiment_engine=ExperimentConfig(type="nop"),
        resource_request=resource_request,
        timeout="5s",
        log_config=log_config,
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
        == f'{retrieved_router.name}-turing-router.{os.getenv("PROJECT_NAME")}.{os.getenv("KSERVICE_DOMAIN")}'
    )

    # get router version with id 1
    router_version_1 = retrieved_router.get_version(1)
    assert router_version_1.status == RouterStatus.DEPLOYED

    channel = grpc.insecure_channel(retrieved_router.endpoint)
    stub = upi_pb2_grpc.UniversalPredictionServiceStub(channel)
    response = stub.PredictValues(upi_pb2.PredictValuesRequest(
        prediction_table=table_pb2.Table(
            columns=[table_pb2.Column(name="col1",type=type_pb2.TYPE_DOUBLE)],
            rows=[table_pb2.Row(values=[table_pb2.Value(double_value=12.2)])],
        )
    ))
    logging.info(response)
