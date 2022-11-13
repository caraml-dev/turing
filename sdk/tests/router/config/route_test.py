import pytest
from turing.router.config.route import (
    MissingServiceMethodException,
    Route,
    InvalidUrlException,
)
from turing.router.config.router_config import RouterConfig, Protocol


@pytest.mark.parametrize(
    "id,endpoint,timeout,protocol,service_method,expected",
    [
        pytest.param(
            "model-a",
            "http://predict_this.io/model-a",
            "100ms",
            None,
            None,
            "generic_route",
        ),
        pytest.param(
            "model-a",
            "http://predict_this.io/model-a",
            "100ms",
            Protocol.HTTP,
            None,
            "generic_route",
        ),
        pytest.param(
            "model-a-grpc",
            "grpc_host:80",
            "100ms",
            Protocol.UPI,
            "package/method",
            "generic_route_grpc",
        ),
    ],
)
def test_create_route_with_valid_endpoint(
    id, endpoint, timeout, protocol, service_method, expected, request
):
    actual = Route(id, endpoint, timeout, service_method).to_open_api()
    RouterConfig(routes=[actual], protocol=protocol)
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "id,endpoint,timeout,protocol,expected",
    [
        pytest.param(
            "route_test_1",
            "http//test_this_route.com/",
            100,
            Protocol.HTTP,
            InvalidUrlException,
        ),
        pytest.param(
            "route_test_2",
            "http//test_this_route.com/",
            100,
            Protocol.UPI,
            MissingServiceMethodException,
        ),
    ],
)
def test_create_route_with_invalid_endpoint(id, endpoint, timeout, protocol, expected):
    with pytest.raises(expected):
        RouterConfig(routes=[Route(id, endpoint, timeout)], protocol=protocol)


@pytest.mark.parametrize(
    "id,endpoint,timeout,expected",
    [
        pytest.param(
            "model-a", "http://predict_this.io/model-b", "100ms", "generic_route"
        )
    ],
)
def test_set_route_with_valid_endpoint(id, endpoint, timeout, expected, request):
    actual = Route(id, endpoint, timeout)
    actual.endpoint = "http://predict_this.io/model-a"
    RouterConfig(routes=[actual])
    assert actual.to_open_api() == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "id,endpoint,timeout,expected",
    [
        pytest.param(
            "model-a", "http//predict_this.io/model-a", "100ms", InvalidUrlException
        )
    ],
)
def test_set_route_with_invalid_endpoint(id, endpoint, timeout, expected):
    actual = Route(id, endpoint, timeout)
    with pytest.raises(expected):
        RouterConfig(routes=[actual])
