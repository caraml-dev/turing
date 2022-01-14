import pytest
import turing.generated.models
from turing.generated.exceptions import ApiValueError
from turing.router.config.common.env_var import EnvVar
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.router_ensembler_config import (RouterEnsemblerConfig,
                                                          DockerRouterEnsemblerConfig,
                                                          StandardRouterEnsemblerConfig,
                                                          InvalidExperimentMappingException)


@pytest.mark.parametrize(
    "id,type,standard_config,docker_config,expected", [
        pytest.param(
            1,
            "standard",
            turing.generated.models.EnsemblerStandardConfig(
                experiment_mappings=[
                    turing.generated.models.EnsemblerStandardConfigExperimentMappings(
                        experiment="experiment-1",
                        treatment="treatment-1",
                        route="route-1"
                    ),
                    turing.generated.models.EnsemblerStandardConfigExperimentMappings(
                        experiment="experiment-2",
                        treatment="treatment-2",
                        route="route-2"
                    )
                ]
            ),
            None,
            "generic_standard_router_ensembler_config"
        ),
        pytest.param(
            1,
            "docker",
            None,
            turing.generated.models.EnsemblerDockerConfig(
                image="test.io/gods-test/turing-ensembler:0.0.0-build.0",
                resource_request=turing.generated.models.ResourceRequest(
                    min_replica=1,
                    max_replica=3,
                    cpu_request='100m',
                    memory_request='512Mi'
                ),
                endpoint=f"http://localhost:5000/ensembler_endpoint",
                timeout="500ms",
                port=5120,
                env=[
                    turing.generated.models.EnvVar(
                        name="env_name",
                        value="env_val")
                ],
                service_account="secret-name-for-google-service-account"
            ),
            "generic_docker_router_ensembler_config"
        )
    ])
def test_create_router_ensembler_config(id, type, standard_config, docker_config, expected, request):
    actual = RouterEnsemblerConfig(
        id=id,
        type=type,
        standard_config=standard_config,
        docker_config=docker_config
    ).to_open_api()
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "id,image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            1,
            "test.io/gods-test/turing-ensembler:0.0.0-build.0",
            ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/ensembler_endpoint",
            "500ms",
            5120,
            [
                EnvVar(
                    name="env_name",
                    value="env_val")
            ],
            "secret-name-for-google-service-account",
            "generic_docker_router_ensembler_config"
        )
    ])
def test_create_docker_router_ensembler_config(
        id,
        image,
        resource_request,
        endpoint,
        timeout,
        port,
        env,
        service_account,
        expected,
        request
):
    actual = DockerRouterEnsemblerConfig(
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


@pytest.mark.parametrize(
    "id,image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            1,
            "#@!#!@#@!",
            ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/ensembler_endpoint",
            "500ms",
            5120,
            [
                EnvVar(
                    name="env_name",
                    value="env_val")
            ],
            "secret-name-for-google-service-account",
            ApiValueError
        )
    ])
def test_create_docker_router_ensembler_config_with_invalid_image(
        id,
        image,
        resource_request,
        endpoint,
        timeout,
        port,
        env,
        service_account,
        expected
):
    with pytest.raises(expected):
        DockerRouterEnsemblerConfig(
            id=id,
            image=image,
            resource_request=resource_request,
            endpoint=endpoint,
            timeout=timeout,
            port=port,
            env=env,
            service_account=service_account
        ).to_open_api()


@pytest.mark.parametrize(
    "id,image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            1,
            "test.io/gods-test/turing-ensembler:0.0.0-build.0",
            ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/ensembler_endpoint",
            "500ks",
            5120,
            [
                EnvVar(
                    name="env_name",
                    value="env_val")
            ],
            "secret-name-for-google-service-account",
            ApiValueError
        )
    ])
def test_create_docker_router_ensembler_config_with_invalid_timeout(
        id,
        image,
        resource_request,
        endpoint,
        timeout,
        port,
        env,
        service_account,
        expected
):
    with pytest.raises(expected):
        DockerRouterEnsemblerConfig(
            id=id,
            image=image,
            resource_request=resource_request,
            endpoint=endpoint,
            timeout=timeout,
            port=port,
            env=env,
            service_account=service_account
        ).to_open_api()


@pytest.mark.parametrize(
    "id,image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            1,
            "test.io/gods-test/turing-ensembler:0.0.0-build.0",
            ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/ensembler_endpoint",
            "500ms",
            5120,
            [
                EnvVar(
                    name="env_!@#!@$!",
                    value="env_val")
            ],
            "secret-name-for-google-service-account",
            ApiValueError
        )
    ])
def test_create_docker_router_ensembler_config_with_invalid_env(
        id,
        image,
        resource_request,
        endpoint,
        timeout,
        port,
        env,
        service_account,
        expected,
):
    with pytest.raises(expected):
        DockerRouterEnsemblerConfig(
            id=id,
            image=image,
            resource_request=resource_request,
            endpoint=endpoint,
            timeout=timeout,
            port=port,
            env=env,
            service_account=service_account
        ).to_open_api()


@pytest.mark.parametrize(
    "id,experiment_mappings,expected", [
        pytest.param(
            1,
            [
                {
                    "experiment": "experiment-1",
                    "treatment": "treatment-1",
                    "route": "route-1"
                },
                {
                    "experiment": "experiment-2",
                    "treatment": "treatment-2",
                    "route": "route-2"
                },
            ],
            "generic_standard_router_ensembler_config"
        )
    ])
def test_create_standard_router_ensembler_config(id, experiment_mappings, expected, request):
    actual = StandardRouterEnsemblerConfig(
        id=id,
        experiment_mappings=experiment_mappings
    ).to_open_api()
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "new_experiment_mappings, id,experiment_mappings,expected", [
        pytest.param(
            [
                {
                    "experiment": "wrong-experiment"
                }
            ],
            1,
            [
                {
                    "experiment": "experiment-1",
                    "treatment": "treatment-1",
                    "route": "route-1"
                },
                {
                    "experiment": "experiment-2",
                    "treatment": "treatment-2",
                    "route": "route-2"
                },
            ],
            InvalidExperimentMappingException
        ),
        pytest.param(
            [
                {
                    "experiment": "experiment-1",
                    "treatment": "treatment-1",
                    "route": 313
                }
            ],
            1,
            [
                {
                    "experiment": "experiment-1",
                    "treatment": "treatment-1",
                    "route": "route-1"
                },
                {
                    "experiment": "experiment-2",
                    "treatment": "treatment-2",
                    "route": "route-2"
                },
            ],
            InvalidExperimentMappingException
        )
    ])
def test_set_standard_router_ensembler_config_with_invalid_experiment_mappings(
        new_experiment_mappings,
        id,
        experiment_mappings,
        expected):
    actual = StandardRouterEnsemblerConfig(
        id=id,
        experiment_mappings=experiment_mappings
    )
    with pytest.raises(expected):
        actual.experiment_mappings = new_experiment_mappings


@pytest.mark.parametrize(
    "new_experiment_mappings,id,experiment_mappings,expected", [
        pytest.param(
            [
                {
                    "experiment": "experiment-1",
                    "treatment": "treatment-1",
                    "route": "route-1"
                },
                {
                    "experiment": "experiment-2",
                    "treatment": "treatment-2",
                    "route": "route-2"
                },
            ],
            1,
            [
                {
                    "experiment": "wrong-experiment",
                    "treatment": "wrong-treatment",
                    "route": "wrong-route"
                }
            ],
            "generic_standard_router_ensembler_config"
        )
    ])
def test_set_standard_router_ensembler_config_with_valid_experiment_mappings(
        new_experiment_mappings,
        id,
        experiment_mappings,
        expected,
        request):
    actual = StandardRouterEnsemblerConfig(
        id=id,
        experiment_mappings=experiment_mappings
    )
    actual.experiment_mappings = new_experiment_mappings
    assert actual.to_open_api() == request.getfixturevalue(expected)
