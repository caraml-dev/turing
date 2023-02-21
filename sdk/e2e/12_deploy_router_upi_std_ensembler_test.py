import os
import logging

from caraml.upi.v1 import upi_pb2_grpc, upi_pb2, variable_pb2
from caraml.upi.v1 import table_pb2, type_pb2
import grpc

import turing
import turing.batch
import turing.batch.config
import turing.router.config.router_config

from turing.router.config.route import Route
from turing.router.config.experiment_config import ExperimentConfig
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.log_config import LogConfig, ResultLoggerType
from turing.router.config.router_config import RouterConfig, Protocol
from turing.router.config.router_ensembler_config import StandardRouterEnsemblerConfig
from turing.router.config.router_version import RouterStatus
from turing.router.config.traffic_rule import (
    DefaultTrafficRule,
    FieldSource,
    TrafficRule,
    TrafficRuleCondition,
)


def test_deploy_router_upi_std_ensembler():
    control_endpoint = f'{os.getenv("MOCKSERVER_UPI_CONTROL_ENDPOINT")}'
    treatment_endpoint = f'{os.getenv("MOCKSERVER_UPI_A_ENDPOINT")}'

    logging.info(f"control endpoint: {control_endpoint}")
    logging.info(f"treatment endpoint: {treatment_endpoint}")

    # set up route
    routes = [
        Route(
            id="control",
            endpoint=control_endpoint,
            service_method="caraml.upi.v1.UniversalPredictionService/PredictValues",
            timeout="5s",
        ),
        Route(
            id="treatment-a",
            endpoint=treatment_endpoint,
            service_method="caraml.upi.v1.UniversalPredictionService/PredictValues",
            timeout="5s",
        ),
    ]

    # set up resource request config for the router
    resource_request = ResourceRequest(
        min_replica=1, max_replica=1, cpu_request="100m", memory_request="250Mi"
    )

    # set up log config for the router
    log_config = LogConfig(result_logger_type=ResultLoggerType.NOP)

    # setup experiment engine
    experiment_engine = ExperimentConfig(
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
                        "field": "client_id",
                        "field_source": "payload",
                    }
                ],
            },
        },
    )

    ensembler = StandardRouterEnsemblerConfig(
        route_name_path="route_name", fallback_response_route_id="control"
    )

    # create the RouterConfig instance
    router_config = RouterConfig(
        environment_name=os.getenv("MODEL_CLUSTER_NAME"),
        name=f'e2e-sdk-upi-std-ensembler-{os.getenv("TEST_ID")}',
        routes=routes,
        default_traffic_rule=DefaultTrafficRule(routes=["control"]),
        experiment_engine=experiment_engine,
        ensembler=ensembler,
        default_route_id="control",
        resource_request=resource_request,
        protocol=Protocol.UPI,
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
        == f'{retrieved_router.name}-turing-router.{os.getenv("PROJECT_NAME")}.{os.getenv("KSERVICE_DOMAIN")}:80'
    )

    # get router version with id 1
    router_version_1 = retrieved_router.get_version(1)
    assert router_version_1.status == RouterStatus.DEPLOYED

    channel = grpc.insecure_channel(retrieved_router.endpoint)
    stub = upi_pb2_grpc.UniversalPredictionServiceStub(channel)

    logging.info("send request that satisfy control")

    # proprietary uses hashing, 4 will return control and 7 will return treatment in this setup
    request = upi_pb2.PredictValuesRequest(
        prediction_table=table_pb2.Table(
            columns=[table_pb2.Column(name="col1", type=type_pb2.TYPE_DOUBLE)],
            rows=[table_pb2.Row(values=[table_pb2.Value(double_value=12.2)])],
        ),
        prediction_context=[
            variable_pb2.Variable(
                name="client_id",
                type=type_pb2.TYPE_STRING,
                string_value="4",
            )
        ],
    )

    response: upi_pb2.PredictValuesResponse = stub.PredictValues(request)
    logging.info(f"received response {response}")

    assert response.prediction_result_table == request.prediction_table
    assert response.metadata.models[0].name == "control"

    logging.info("send request that goes to treatment-a")

    request = upi_pb2.PredictValuesRequest(
        prediction_table=table_pb2.Table(
            columns=[table_pb2.Column(name="col1", type=type_pb2.TYPE_DOUBLE)],
            rows=[table_pb2.Row(values=[table_pb2.Value(double_value=12.2)])],
        ),
        prediction_context=[
            variable_pb2.Variable(
                name="client_id",
                type=type_pb2.TYPE_STRING,
                string_value="7",
            )
        ],
    )

    response: upi_pb2.PredictValuesResponse = stub.PredictValues(request)
    logging.info(f"received response {response}")

    assert response.prediction_result_table == request.prediction_table
    assert response.metadata.models[0].name == "treatment-a"
