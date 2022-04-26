import pytest
from turing.router.config.route import Route, DuplicateRouteException, InvalidRouteException


@pytest.mark.parametrize(
    "actual,new_routes,expected", [
        pytest.param(
            "generic_router_config",
            [
                Route("model-a", "http://predict-this.io/model-a", "100ms"),
                Route("model-a", "http://predict-this.io/model-b", "100ms")
            ],
            DuplicateRouteException
        )
    ])
def test_set_router_config_with_invalid_routes(actual, new_routes, expected, request):
    actual = request.getfixturevalue(actual)
    actual.routes = new_routes
    with pytest.raises(expected):
        actual.to_open_api()

@pytest.mark.parametrize(
    "actual,invalid_route_id,expected", [
        pytest.param(
            "generic_router_config",
            "test-route-not-exists",
            InvalidRouteException
        )
    ])
def test_set_router_config_with_invalid_default_route(actual, invalid_route_id, expected, request):
    actual = request.getfixturevalue(actual)
    actual.default_route_id = invalid_route_id
    with pytest.raises(expected):
        actual.to_open_api()
