import pytest
from turing.generated.exceptions import ApiValueError
from turing.router.config.resource_request import (
    ResourceRequest,
    InvalidReplicaCountException,
)


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected",
    [pytest.param(1, 3, "100m", "512Mi", "generic_resource_request")],
)
def test_create_resource_request_with_valid_params(
    min_replica, max_replica, cpu_request, memory_request, expected, request
):
    actual = ResourceRequest(
        min_replica, max_replica, cpu_request, memory_request
    ).to_open_api()
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected",
    [
        pytest.param(
            ResourceRequest.min_allowed_replica - 1,
            3,
            "100m",
            "512Mi",
            InvalidReplicaCountException,
        )
    ],
)
def test_create_resource_request_with_min_replica_below_min_allowed(
    min_replica,
    max_replica,
    cpu_request,
    memory_request,
    expected,
):
    with pytest.raises(expected):
        ResourceRequest(min_replica, max_replica, cpu_request, memory_request)


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected",
    [
        pytest.param(
            1,
            ResourceRequest.max_allowed_replica + 1,
            "100m",
            "512Mi",
            InvalidReplicaCountException,
        )
    ],
)
def test_create_resource_request_with_max_replica_above_max_allowed(
    min_replica,
    max_replica,
    cpu_request,
    memory_request,
    expected,
):
    with pytest.raises(expected):
        ResourceRequest(min_replica, max_replica, cpu_request, memory_request)


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected",
    [pytest.param(5, 4, "100m", "512Mi", InvalidReplicaCountException)],
)
def test_create_resource_request_with_min_replica_greater_than_max_replica(
    min_replica,
    max_replica,
    cpu_request,
    memory_request,
    expected,
):
    with pytest.raises(expected):
        ResourceRequest(min_replica, max_replica, cpu_request, memory_request)


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected",
    [pytest.param(2, 5, "3m", "512Mi", ApiValueError)],
)
def test_create_resource_request_with_invalid_cpu_request_string(
    min_replica,
    max_replica,
    cpu_request,
    memory_request,
    expected,
):
    with pytest.raises(expected):
        ResourceRequest(
            min_replica, max_replica, cpu_request, memory_request
        ).to_open_api()


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected",
    [pytest.param(2, 5, "33m", "512Ri", ApiValueError)],
)
def test_create_resource_request_with_invalid_memory_request_string(
    min_replica,
    max_replica,
    cpu_request,
    memory_request,
    expected,
):
    with pytest.raises(expected):
        ResourceRequest(
            min_replica, max_replica, cpu_request, memory_request
        ).to_open_api()


@pytest.mark.parametrize(
    "new_min_replica,min_replica,max_replica,cpu_request,memory_request,expected",
    [pytest.param(1, 2, 3, "100m", "512Mi", "generic_resource_request")],
)
def test_create_resource_request_with_valid_min_replica(
    new_min_replica,
    min_replica,
    max_replica,
    cpu_request,
    memory_request,
    expected,
    request,
):
    actual = ResourceRequest(min_replica, max_replica, cpu_request, memory_request)
    actual.min_replica = new_min_replica
    assert actual.to_open_api() == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "new_max_replica,min_replica,max_replica,cpu_request,memory_request,expected",
    [pytest.param(3, 1, 6, "100m", "512Mi", "generic_resource_request")],
)
def test_create_resource_request_with_valid_max_replica(
    new_max_replica,
    min_replica,
    max_replica,
    cpu_request,
    memory_request,
    expected,
    request,
):
    actual = ResourceRequest(min_replica, max_replica, cpu_request, memory_request)
    actual.max_replica = new_max_replica
    assert actual.to_open_api() == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "new_min_replica,min_replica,max_replica,cpu_request,memory_request,expected",
    [pytest.param(-1, 1, 3, "100m", "512Mi", InvalidReplicaCountException)],
)
def test_set_resource_request_with_min_replica_below_min_allowed(
    new_min_replica, min_replica, max_replica, cpu_request, memory_request, expected
):
    actual = ResourceRequest(min_replica, max_replica, cpu_request, memory_request)
    with pytest.raises(expected):
        actual.min_replica = new_min_replica


@pytest.mark.parametrize(
    "new_max_replica,min_replica,max_replica,cpu_request,memory_request,expected",
    [pytest.param(50, 1, 3, "100m", "512Mi", InvalidReplicaCountException)],
)
def test_set_resource_request_with_max_replica_above_max_allowed(
    new_max_replica, min_replica, max_replica, cpu_request, memory_request, expected
):
    actual = ResourceRequest(min_replica, max_replica, cpu_request, memory_request)
    with pytest.raises(expected):
        actual.max_replica = new_max_replica


@pytest.mark.parametrize(
    "new_min_replica,min_replica,max_replica,cpu_request,memory_request,expected",
    [pytest.param(5, 1, 3, "100m", "512Mi", InvalidReplicaCountException)],
)
def test_set_resource_request_with_min_replica_greater_than_max_replica(
    new_min_replica, min_replica, max_replica, cpu_request, memory_request, expected
):
    actual = ResourceRequest(min_replica, max_replica, cpu_request, memory_request)
    with pytest.raises(expected):
        actual.min_replica = new_min_replica


@pytest.mark.parametrize(
    "new_max_replica,min_replica,max_replica,cpu_request,memory_request,expected",
    [pytest.param(4, 5, 10, "100m", "512Mi", InvalidReplicaCountException)],
)
def test_set_resource_request_with_max_replica_below_min_replica(
    new_max_replica, min_replica, max_replica, cpu_request, memory_request, expected
):
    actual = ResourceRequest(min_replica, max_replica, cpu_request, memory_request)
    with pytest.raises(expected):
        actual.max_replica = new_max_replica


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected",
    [pytest.param(1, 3, "3m", "512Mi", ApiValueError)],
)
def test_create_resource_request_with_invalid_cpu_request_string(
    min_replica, max_replica, cpu_request, memory_request, expected
):
    actual = ResourceRequest(min_replica, max_replica, cpu_request, memory_request)
    with pytest.raises(expected):
        actual.to_open_api()


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected",
    [pytest.param(1, 3, "100m", "512Ri", ApiValueError)],
)
def test_create_resource_request_with_invalid_memory_request_string(
    min_replica, max_replica, cpu_request, memory_request, expected
):
    actual = ResourceRequest(min_replica, max_replica, cpu_request, memory_request)
    with pytest.raises(expected):
        actual.to_open_api()
