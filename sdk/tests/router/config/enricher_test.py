import turing.generated.models
import turing.router.config.enricher
import turing.router.config.common.common
import turing.router.config.common.schemas
import turing.router.config.resource_request
import pytest


@pytest.mark.parametrize(
    "id,image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            1,
            "test.io/gods-test/turing-enricher:0.0.0-build.0",
            turing.router.config.resource_request.ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/enricher_endpoint",
            "500ms",
            5180,
            [
                turing.router.config.common.common.EnvVar(
                        name="env_name",
                        value="env_val"
                )
            ],
            "service-account",
            "generic_enricher"
        )
    ])
def test_create_enricher(id, image, resource_request, endpoint, timeout, port, env, service_account, expected, request):
    actual = turing.router.config.enricher.Enricher(
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
