import turing.generated.models
import turing.router.config.resource_request
import pytest


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            1,
            3,
            "100m",
            "512Mi",
            "generic_resource_request"
        )
    ])
def test_create_resource_request_with_valid_params(
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected,
        request
):
    actual = turing.router.config.resource_request.ResourceRequest(
        min_replica,
        max_replica,
        cpu_request,
        memory_request
    ).to_open_api()
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            turing.router.config.resource_request.ResourceRequest.min_allowed_replica - 1,
            3,
            "100m",
            "512Mi",
            turing.router.config.resource_request.InvalidReplicaCountException
        )
    ])
def test_create_resource_request_with_min_replica_below_min_allowed(
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected,
):
    with pytest.raises(expected):
        turing.router.config.resource_request.ResourceRequest(
            min_replica,
            max_replica,
            cpu_request,
            memory_request
        ).to_open_api()


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            1,
            turing.router.config.resource_request.ResourceRequest.max_allowed_replica + 1,
            "100m",
            "512Mi",
            turing.router.config.resource_request.InvalidReplicaCountException
        )
    ])
def test_create_resource_request_with_max_replica_above_max_allowed(
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected,
):
    with pytest.raises(expected):
        turing.router.config.resource_request.ResourceRequest(
            min_replica,
            max_replica,
            cpu_request,
            memory_request
        ).to_open_api()


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            5,
            5,
            "100m",
            "512Mi",
            turing.router.config.resource_request.InvalidReplicaCountException
        )
    ])
def test_create_resource_request_with_min_replica_geq_max_replica(
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected,
):
    with pytest.raises(expected):
        turing.router.config.resource_request.ResourceRequest(
            min_replica,
            max_replica,
            cpu_request,
            memory_request
        ).to_open_api()


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            2,
            5,
            "3m",
            "512Mi",
            turing.router.config.resource_request.InvalidCPURequestException
        )
    ])
def test_create_resource_request_with_invalid_cpu_request_string(
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected,
):
    with pytest.raises(expected):
        turing.router.config.resource_request.ResourceRequest(
            min_replica,
            max_replica,
            cpu_request,
            memory_request
        ).to_open_api()


@pytest.mark.parametrize(
    "min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            2,
            5,
            "33m",
            "512Ri",
            turing.router.config.resource_request.InvalidMemoryRequestException
        )
    ])
def test_create_resource_request_with_invalid_memory_request_string(
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected,
):
    with pytest.raises(expected):
        turing.router.config.resource_request.ResourceRequest(
            min_replica,
            max_replica,
            cpu_request,
            memory_request
        ).to_open_api()


@pytest.mark.parametrize(
    "new_min_replica,min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            1,
            2,
            3,
            "100m",
            "512Mi",
            "generic_resource_request"
        )
    ])
def test_create_resource_request_with_valid_min_replica(
        new_min_replica,
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected,
        request
):
    actual = turing.router.config.resource_request.ResourceRequest(
        min_replica,
        max_replica,
        cpu_request,
        memory_request
    )
    actual.min_replica = new_min_replica
    assert actual.to_open_api() == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "new_max_replica,min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            3,
            1,
            6,
            "100m",
            "512Mi",
            "generic_resource_request"
        )
    ])
def test_create_resource_request_with_valid_max_replica(
        new_max_replica,
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected,
        request
):
    actual = turing.router.config.resource_request.ResourceRequest(
        min_replica,
        max_replica,
        cpu_request,
        memory_request
    )
    actual.max_replica = new_max_replica
    assert actual.to_open_api() == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "new_min_replica,min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            -1,
            1,
            3,
            "100m",
            "512Mi",
            turing.router.config.resource_request.InvalidReplicaCountException
        )
    ])
def test_set_resource_request_with_min_replica_below_min_allowed(
        new_min_replica,
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected
):
    actual = turing.router.config.resource_request.ResourceRequest(
        min_replica,
        max_replica,
        cpu_request,
        memory_request
    )
    with pytest.raises(expected):
        actual.min_replica = new_min_replica


@pytest.mark.parametrize(
    "new_max_replica,min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            50,
            1,
            3,
            "100m",
            "512Mi",
            turing.router.config.resource_request.InvalidReplicaCountException
        )
    ])
def test_set_resource_request_with_max_replica_above_max_allowed(
        new_max_replica,
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected
):
    actual = turing.router.config.resource_request.ResourceRequest(
        min_replica,
        max_replica,
        cpu_request,
        memory_request
    )
    with pytest.raises(expected):
        actual.max_replica = new_max_replica


@pytest.mark.parametrize(
    "new_min_replica,min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            3,
            1,
            3,
            "100m",
            "512Mi",
            turing.router.config.resource_request.InvalidReplicaCountException
        )
    ])
def test_set_resource_request_with_min_replica_geq_max_replica(
        new_min_replica,
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected
):
    actual = turing.router.config.resource_request.ResourceRequest(
        min_replica,
        max_replica,
        cpu_request,
        memory_request
    )
    with pytest.raises(expected):
        actual.min_replica = new_min_replica


@pytest.mark.parametrize(
    "new_max_replica,min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            4,
            5,
            10,
            "100m",
            "512Mi",
            turing.router.config.resource_request.InvalidReplicaCountException
        )
    ])
def test_set_resource_request_with_max_replica_below_min_replica(
        new_max_replica,
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected
):
    actual = turing.router.config.resource_request.ResourceRequest(
        min_replica,
        max_replica,
        cpu_request,
        memory_request
    )
    with pytest.raises(expected):
        actual.max_replica = new_max_replica


@pytest.mark.parametrize(
    "new_cpu_request,min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            "100m",
            1,
            3,
            "150m",
            "512Mi",
            "generic_resource_request"
        )
    ])
def test_create_resource_request_with_valid_cpu_request(
        new_cpu_request,
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected,
        request
):
    actual = turing.router.config.resource_request.ResourceRequest(
        min_replica,
        max_replica,
        cpu_request,
        memory_request
    )
    actual.cpu_request = new_cpu_request
    assert actual.to_open_api() == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "new_cpu_request,min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            "3m",
            1,
            3,
            "100m",
            "512Mi",
            turing.router.config.resource_request.InvalidCPURequestException
        )
    ])
def test_set_resource_request_with_invalid_cpu_request_string(
        new_cpu_request,
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected
):
    actual = turing.router.config.resource_request.ResourceRequest(
        min_replica,
        max_replica,
        cpu_request,
        memory_request
    )
    with pytest.raises(expected):
        actual.cpu_request = new_cpu_request


@pytest.mark.parametrize(
    "new_memory_request, min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            "512Mi",
            1,
            3,
            "100m",
            "512Ki",
            "generic_resource_request"
        )
    ])
def test_create_resource_request_with_valid_memory_request(
        new_memory_request,
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected,
        request
):
    actual = turing.router.config.resource_request.ResourceRequest(
        min_replica,
        max_replica,
        cpu_request,
        memory_request
    )
    actual.memory_request = new_memory_request
    assert actual.to_open_api() == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "new_memory_request,min_replica,max_replica,cpu_request,memory_request,expected", [
        pytest.param(
            "512Ri",
            1,
            3,
            "100m",
            "512Mi",
            turing.router.config.resource_request.InvalidMemoryRequestException
        )
    ])
def test_set_resource_request_with_invalid_memory_request_string(
        new_memory_request,
        min_replica,
        max_replica,
        cpu_request,
        memory_request,
        expected
):
    actual = turing.router.config.resource_request.ResourceRequest(
        min_replica,
        max_replica,
        cpu_request,
        memory_request
    )
    with pytest.raises(expected):
        actual.memory_request = new_memory_request
