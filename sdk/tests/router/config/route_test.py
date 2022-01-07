import turing.generated.models
import turing.router.config.route
import pytest


@pytest.mark.parametrize(
    "id,endpoint,timeout,expected", [
        pytest.param(
            "route_test_1",
            "https://test_this_route.com/",
            100,
            turing.generated.models.Route(
                id="route_test_1",
                type="PROXY",
                endpoint="https://test_this_route.com/",
                timeout="100ms"
            )
        )
    ])
def test_create_route_with_valid_endpoint(id, endpoint, timeout, expected):
    actual = turing.router.config.route.Route(id, endpoint, timeout).to_open_api()
    assert actual == expected


@pytest.mark.parametrize(
    "id,endpoint,timeout,expected", [
        (
            "route_test_1",
            "http//test_this_route.com/",
            100,
            turing.router.config.route.InvalidUrlException
        )
    ])
def test_create_route_with_invalid_endpoint(id, endpoint, timeout, expected):
    with pytest.raises(expected):
        turing.router.config.route.Route(id, endpoint, timeout)
