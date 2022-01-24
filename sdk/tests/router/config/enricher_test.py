import pytest
from turing.router.config.enricher import Enricher
from turing.router.config.common.env_var import EnvVar
from turing.router.config.resource_request import ResourceRequest


@pytest.mark.parametrize(
    "id,image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            1,
            "test.io/just-a-test/turing-enricher:0.0.0-build.0",
            ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/enricher_endpoint",
            "500ms",
            5180,
            [
                EnvVar(
                        name="env_name",
                        value="env_val"
                )
            ],
            "service-account",
            "generic_enricher"
        )
    ])
def test_create_enricher(id, image, resource_request, endpoint, timeout, port, env, service_account, expected, request):
    actual = Enricher(
        id=id,
        image=image,
        resource_request=resource_request,
        endpoint=endpoint,
        timeout=timeout,
        port=port,
        env=env,
        service_account=service_account
    ).to_open_api()
    assert actual == request.getfixturevalue(expected)
