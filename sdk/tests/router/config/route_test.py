import pytest
from turing.router.config.route import MissingServiceMethodException, Route, InvalidUrlException, RouteProtocol


@pytest.mark.parametrize(
    "id,endpoint,timeout, protocol, service_method, expected",
    [
        pytest.param(
            "model-a", "http://predict_this.io/model-a", "100ms", RouteProtocol.HTTP, None, "generic_route"
        ),
        pytest.param(
            "model-a-grpc", "grpc_host:80", "100ms", RouteProtocol.GRPC, "package/method", "generic_route_grpc"
        ),
    ],
)
def test_create_route_with_valid_endpoint(id, endpoint, timeout, protocol, service_method, expected, request):
    actual = Route(id, endpoint, timeout, protocol, service_method).to_open_api()
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "id,endpoint,timeout,protocol,service_method,expected",
    [
        pytest.param(
            "route_test_1", "http//test_this_route.com/", 100, RouteProtocol.HTTP, None, InvalidUrlException
        ),
        pytest.param(
            "route_test_2", "http//test_this_route.com/", 100, RouteProtocol.GRPC, None, MissingServiceMethodException
        )
    ],
)
def test_create_route_with_invalid_endpoint(id, endpoint, timeout, protocol, service_method, expected):
    with pytest.raises(expected):
        Route(id, endpoint, timeout, protocol=protocol, service_method=service_method)


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
    assert actual.to_open_api() == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "id,endpoint,timeout,expected",
    [
        pytest.param(
            "model-a", "http://predict_this.io/model-a", "100ms", InvalidUrlException
        )
    ],
)
def test_set_route_with_invalid_endpoint(id, endpoint, timeout, expected):
    actual = Route(id, endpoint, timeout)
    with pytest.raises(expected):
        actual.endpoint = "http//test_this_route.com/"
