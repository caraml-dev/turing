import pytest
from turing.mounted_mlp_secret import MountedMLPSecret
from turing.router.config.autoscaling_policy import DEFAULT_AUTOSCALING_POLICY
from turing.router.config.enricher import Enricher
from turing.router.config.common.env_var import EnvVar
from turing.router.config.autoscaling_policy import AutoscalingPolicy
from turing.router.config.resource_request import ResourceRequest


@pytest.mark.parametrize(
    "id,image,resource_request,autoscaling_policy,endpoint,timeout,port,env,secrets,service_account,expected",
    [
        pytest.param(
            1,
            "test.io/just-a-test/turing-enricher:0.0.0-build.0",
            ResourceRequest(
                min_replica=1, max_replica=3, cpu_request="100m", memory_request="512Mi"
            ),
            AutoscalingPolicy(metric="rps", target="100"),
            f"http://localhost:5000/enricher_endpoint",
            "500ms",
            5180,
            [EnvVar(name="env_name", value="env_val")],
            [
                MountedMLPSecret(
                    mlp_secret_name="mlp_secret_name", env_var_name="env_var_name"
                )
            ],
            "service-account",
            "generic_enricher",
        )
    ],
)
def test_create_enricher(
    id,
    image,
    resource_request,
    autoscaling_policy,
    endpoint,
    timeout,
    port,
    env,
    secrets,
    service_account,
    expected,
    request,
):
    actual = Enricher(
        id=id,
        image=image,
        resource_request=resource_request,
        autoscaling_policy=autoscaling_policy,
        endpoint=endpoint,
        timeout=timeout,
        port=port,
        env=env,
        secrets=secrets,
        service_account=service_account,
    ).to_open_api()
    assert actual == request.getfixturevalue(expected)


def test_default_enricher_autoscaling_policy():
    assert (
        Enricher(
            id=id,
            image="image",
            resource_request=ResourceRequest(
                min_replica=1, max_replica=3, cpu_request="100m", memory_request="512Mi"
            ),
            endpoint="endpoint",
            timeout="1s",
            port=8080,
            env=EnvVar(name="env_name", value="env_val"),
            secrets=[
                MountedMLPSecret(
                    mlp_secret_name="mlp_secret_name", env_var_name="env_var_name"
                )
            ],
            service_account="service_account",
        ).autoscaling_policy
        == DEFAULT_AUTOSCALING_POLICY
    )
