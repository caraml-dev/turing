import pytest

from turing.router.config.route import (
    Route,
    DuplicateRouteException,
    InvalidRouteException,
)
from turing.router.config.router_config import RouterConfig
from turing.router.config.log_config import (
    LogConfig,
    ResultLoggerType,
    InvalidResultLoggerTypeAndConfigCombination,
)
from turing.router.config.autoscaling_policy import DEFAULT_AUTOSCALING_POLICY
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.router_ensembler_config import (
    DockerRouterEnsemblerConfig,
    NopRouterEnsemblerConfig,
    StandardRouterEnsemblerConfig,
    PyfuncRouterEnsemblerConfig,
    RouterEnsemblerConfig,
    InvalidEnsemblerTypeException,
)


@pytest.mark.parametrize(
    "actual,new_routes,expected",
    [
        pytest.param(
            "generic_router_config",
            [
                Route("model-a", "http://predict-this.io/model-a", "100ms"),
                Route("model-a", "http://predict-this.io/model-b", "100ms"),
            ],
            DuplicateRouteException,
        )
    ],
)
def test_set_router_config_with_invalid_routes(actual, new_routes, expected, request):
    actual = request.getfixturevalue(actual)
    actual.routes = new_routes
    with pytest.raises(expected):
        actual.to_open_api()


@pytest.mark.parametrize(
    "actual,invalid_route_id,expected",
    [
        pytest.param(
            "generic_router_config", "test-route-not-exists", InvalidRouteException
        )
    ],
)
def test_set_router_config_with_invalid_default_route(
    actual, invalid_route_id, expected, request
):
    actual = request.getfixturevalue(actual)
    actual.ensembler = None
    actual.default_route_id = invalid_route_id
    with pytest.raises(expected):
        actual.to_open_api()


@pytest.mark.parametrize("actual,expected", [pytest.param("generic_router_config", {})])
def test_remove_router_config_default_route(actual, expected, request):
    actual = request.getfixturevalue(actual)
    assert "default_route_id" not in actual.to_open_api()


@pytest.mark.parametrize(
    "router,type,nop_config,standard_config,docker_config,pyfunc_config,expected_class",
    [
        pytest.param(
            "generic_router_config",
            "nop",
            "nop_router_ensembler_config",
            None,
            None,
            None,
            NopRouterEnsemblerConfig,
        ),
        pytest.param(
            "generic_router_config",
            "standard",
            None,
            "standard_router_ensembler_config_with_experiment_mappings",
            None,
            None,
            StandardRouterEnsemblerConfig,
        ),
        pytest.param(
            "generic_router_config",
            "standard",
            None,
            "standard_router_ensembler_config_with_route_name_path",
            None,
            None,
            StandardRouterEnsemblerConfig,
        ),
        pytest.param(
            "generic_router_config",
            "docker",
            None,
            None,
            "generic_ensembler_docker_config",
            None,
            DockerRouterEnsemblerConfig,
        ),
        pytest.param(
            "generic_router_config",
            "pyfunc",
            None,
            None,
            None,
            "generic_ensembler_pyfunc_config",
            PyfuncRouterEnsemblerConfig,
        ),
    ],
)
def test_set_router_config_base_ensembler(
    router,
    type,
    nop_config,
    standard_config,
    docker_config,
    pyfunc_config,
    expected_class,
    request,
):
    actual = request.getfixturevalue(router)
    ensembler = RouterEnsemblerConfig(
        type=type,
        nop_config=None if nop_config is None else request.getfixturevalue(nop_config),
        standard_config=(
            None
            if standard_config is None
            else request.getfixturevalue(standard_config)
        ),
        docker_config=(
            None if docker_config is None else request.getfixturevalue(docker_config)
        ),
        pyfunc_config=(
            None if pyfunc_config is None else request.getfixturevalue(pyfunc_config)
        ),
    )
    actual.ensembler = ensembler
    assert isinstance(actual.ensembler, expected_class)


@pytest.mark.parametrize(
    "router_config,default_route_id,ensembler,expected",
    [
        pytest.param(
            "generic_router_config",
            "model-a",
            StandardRouterEnsemblerConfig(
                experiment_mappings=[], fallback_response_route_id="model-b"
            ),
            "model-b",
        ),
        pytest.param(
            "generic_router_config",
            "model-a",
            NopRouterEnsemblerConfig(final_response_route_id="model-b"),
            "model-b",
        ),
        pytest.param(
            "generic_router_config",
            "model-a",
            "docker_router_ensembler_config",
            None,
        ),
        pytest.param(
            "generic_router_config",
            "model-a",
            "pyfunc_router_ensembler_config",
            None,
        ),
    ],
)
def test_default_route_id_by_ensembler_config(
    router_config, default_route_id, ensembler, expected, request
):
    router = request.getfixturevalue(router_config)
    router.default_route_id = default_route_id
    router.ensembler = (
        request.getfixturevalue(ensembler) if isinstance(ensembler, str) else ensembler
    )
    if expected:
        assert router.to_open_api().to_dict()["config"]["default_route_id"] == expected
    else:
        assert "default_route_id" not in router.to_open_api().to_dict()["config"]


def test_default_router_autoscaling_policy(request):
    router_config = request.getfixturevalue("generic_router_config")
    config = router_config.to_dict()
    del config["autoscaling_policy"]
    assert RouterConfig(**config).autoscaling_policy == DEFAULT_AUTOSCALING_POLICY


@pytest.mark.parametrize(
    "base_config, ensembler_config, expected_err",
    [
        pytest.param(
            "minimal_upi_router_config",
            NopRouterEnsemblerConfig(final_response_route_id="control"),
            None,
        ),
        pytest.param(
            "minimal_upi_router_config",
            StandardRouterEnsemblerConfig(fallback_response_route_id="control"),
            None,
        ),
        pytest.param(
            "minimal_upi_router_config",
            DockerRouterEnsemblerConfig(
                image="test.io/just-a-test/turing-ensembler:0.0.0-build.0",
                resource_request=ResourceRequest(
                    min_replica=1,
                    max_replica=3,
                    cpu_request="500m",
                    memory_request="512Mi",
                ),
                endpoint=f"http://localhost:5000/ensembler_endpoint",
                timeout="500ms",
                port=5120,
                env=[],
            ),
            InvalidEnsemblerTypeException,
        ),
        pytest.param(
            "minimal_upi_router_config",
            PyfuncRouterEnsemblerConfig(
                project_id=1,
                ensembler_id=1,
                resource_request=ResourceRequest(
                    min_replica=0,
                    max_replica=2,
                    cpu_request="500m",
                    memory_request="512Mi",
                ),
                timeout="60ms",
                env=[],
            ),
            InvalidEnsemblerTypeException,
        ),
    ],
)
def test_upi_router_ensembler_config_constraint(
    base_config, ensembler_config, expected_err, request
):
    router_config = request.getfixturevalue(base_config)

    if expected_err:
        with pytest.raises(expected_err):
            router_config.ensembler = ensembler_config
    else:
        router_config.ensembler = ensembler_config
        router_config.to_open_api()
