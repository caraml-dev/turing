import turing.generated.models
import turing.router.config.route
import pytest


@pytest.mark.parametrize(
    "id,endpoint,timeout,expected", [
        pytest.param(
            "route_test_1",
            "https://test_this_route.com/",
            100,
            "generic_route"
        )
    ])
def test_create_route_with_valid_endpoint(id, endpoint, timeout, expected, request):
    actual = turing.router.config.route.Route(id, endpoint, timeout).to_open_api()
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "id,endpoint,timeout,expected", [
        pytest.param(
            "route_test_1",
            "http//test_this_route.com/",
            100,
            turing.router.config.route.InvalidUrlException
        )
    ])
def test_create_route_with_invalid_endpoint(id, endpoint, timeout, expected):
    with pytest.raises(expected):
        turing.router.config.route.Route(id, endpoint, timeout)
