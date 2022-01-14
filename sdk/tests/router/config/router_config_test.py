import pytest
from turing.router.config.route import Route, DuplicateRouteException


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
