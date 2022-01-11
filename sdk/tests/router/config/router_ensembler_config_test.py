import turing.generated.models
import turing.router.config.router_ensembler_config
import turing.router.config.common.common
import turing.router.config.common.schemas
import turing.router.config.resource_request
import pytest


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
    actual = turing.router.config.router_ensembler_config.RouterEnsemblerConfig(
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
            turing.router.config.resource_request.ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/ensembler_endpoint",
            "500ms",
            5120,
            [
                turing.router.config.common.common.EnvVar(
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
    actual = turing.router.config.router_ensembler_config.DockerRouterEnsemblerConfig(
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
            turing.router.config.resource_request.ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/ensembler_endpoint",
            "500ms",
            5120,
            [
                turing.router.config.common.common.EnvVar(
                    name="env_name",
                    value="env_val")
            ],
            "secret-name-for-google-service-account",
            turing.router.config.common.schemas.InvalidImageException
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
        turing.router.config.router_ensembler_config.DockerRouterEnsemblerConfig(
            id=id,
            image=image,
            resource_request=resource_request,
            endpoint=endpoint,
            timeout=timeout,
            port=port,
            env=env,
            service_account=service_account
        )


@pytest.mark.parametrize(
    "new_image,id,image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            "#@!#!@#@!",
            1,
            "test.io/gods-test/turing-ensembler:0.0.0-build.0",
            turing.router.config.resource_request.ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/ensembler_endpoint",
            "500ms",
            5120,
            [
                turing.router.config.common.common.EnvVar(
                    name="env_name",
                    value="env_val")
            ],
            "secret-name-for-google-service-account",
            turing.router.config.common.schemas.InvalidImageException
        )
    ])
def test_set_docker_router_ensembler_config_with_invalid_image(
        new_image,
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
    actual = turing.router.config.router_ensembler_config.DockerRouterEnsemblerConfig(
        id=id,
        image=image,
        resource_request=resource_request,
        endpoint=endpoint,
        timeout=timeout,
        port=port,
        env=env,
        service_account=service_account
    )
    with pytest.raises(expected):
        actual.image = new_image


@pytest.mark.parametrize(
    "new_image,id,image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            "test.io/gods-test/turing-ensembler:0.0.0-build.0",
            1,
            "test.io/gods-test/not-the-right-ensembler:0.0.0-build.1",
            turing.router.config.resource_request.ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/ensembler_endpoint",
            "500ms",
            5120,
            [
                turing.router.config.common.common.EnvVar(
                    name="env_name",
                    value="env_val")
            ],
            "secret-name-for-google-service-account",
            "generic_docker_router_ensembler_config"
        )
    ])
def test_set_docker_router_ensembler_config_with_valid_image(
        new_image,
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
    actual = turing.router.config.router_ensembler_config.DockerRouterEnsemblerConfig(
        id=id,
        image=image,
        resource_request=resource_request,
        endpoint=endpoint,
        timeout=timeout,
        port=port,
        env=env,
        service_account=service_account
    )
    actual.image = new_image
    assert actual.to_open_api() == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "id,image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            1,
            "test.io/gods-test/turing-ensembler:0.0.0-build.0",
            turing.router.config.resource_request.ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/ensembler_endpoint",
            "500ks",
            5120,
            [
                turing.router.config.common.common.EnvVar(
                    name="env_name",
                    value="env_val")
            ],
            "secret-name-for-google-service-account",
            turing.router.config.common.schemas.InvalidTimeoutException
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
        turing.router.config.router_ensembler_config.DockerRouterEnsemblerConfig(
            id=id,
            image=image,
            resource_request=resource_request,
            endpoint=endpoint,
            timeout=timeout,
            port=port,
            env=env,
            service_account=service_account
        )


@pytest.mark.parametrize(
    "new_timeout,id,image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            "500ks",
            1,
            "test.io/gods-test/turing-ensembler:0.0.0-build.0",
            turing.router.config.resource_request.ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/ensembler_endpoint",
            "500ms",
            5120,
            [
                turing.router.config.common.common.EnvVar(
                    name="env_name",
                    value="env_val")
            ],
            "secret-name-for-google-service-account",
            turing.router.config.common.schemas.InvalidTimeoutException
        )
    ])
def test_set_docker_router_ensembler_config_with_invalid_timeout(
        new_timeout,
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
    actual = turing.router.config.router_ensembler_config.DockerRouterEnsemblerConfig(
        id=id,
        image=image,
        resource_request=resource_request,
        endpoint=endpoint,
        timeout=timeout,
        port=port,
        env=env,
        service_account=service_account
    )
    with pytest.raises(expected):
        actual.timeout = new_timeout


@pytest.mark.parametrize(
    "new_timeout,id,image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            "500ms",
            1,
            "test.io/gods-test/turing-ensembler:0.0.0-build.0",
            turing.router.config.resource_request.ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/ensembler_endpoint",
            "1000ms",
            5120,
            [
                turing.router.config.common.common.EnvVar(
                    name="env_name",
                    value="env_val")
            ],
            "secret-name-for-google-service-account",
            "generic_docker_router_ensembler_config"
        )
    ])
def test_set_docker_router_ensembler_config_with_valid_timeout(
        new_timeout,
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
    actual = turing.router.config.router_ensembler_config.DockerRouterEnsemblerConfig(
        id=id,
        image=image,
        resource_request=resource_request,
        endpoint=endpoint,
        timeout=timeout,
        port=port,
        env=env,
        service_account=service_account
    )
    actual.timeout = new_timeout
    assert actual.to_open_api() == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "new_env,id,image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            [
                turing.router.config.common.common.EnvVar(
                    name="env_name",
                    value="env_val")
            ],
            1,
            "test.io/gods-test/turing-ensembler:0.0.0-build.0",
            turing.router.config.resource_request.ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            f"http://localhost:5000/ensembler_endpoint",
            "500ms",
            5120,
            [
                turing.router.config.common.common.EnvVar(
                    name="env_not_the_right_name",
                    value="env_val")
            ],
            "secret-name-for-google-service-account",
            "generic_docker_router_ensembler_config"
        )
    ])
def test_set_docker_router_ensembler_config_with_valid_env(
        new_env,
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
    actual = turing.router.config.router_ensembler_config.DockerRouterEnsemblerConfig(
        id=id,
        image=image,
        resource_request=resource_request,
        endpoint=endpoint,
        timeout=timeout,
        port=port,
        env=env,
        service_account=service_account
    )
    actual.env = new_env
    assert actual.to_open_api() == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "id,experiment_mappings,expected", [
        pytest.param(
            1,
            [
                {
                    "experiment":"experiment-1",
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
    actual = turing.router.config.router_ensembler_config.StandardRouterEnsemblerConfig(
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
            turing.router.config.router_ensembler_config.InvalidExperimentMappingException
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
            turing.router.config.router_ensembler_config.InvalidExperimentMappingException
        )
    ])
def test_set_standard_router_ensembler_config_with_invalid_experiment_mappings(
        new_experiment_mappings,
        id,
        experiment_mappings,
        expected):
    actual = turing.router.config.router_ensembler_config.StandardRouterEnsemblerConfig(
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
    actual = turing.router.config.router_ensembler_config.StandardRouterEnsemblerConfig(
        id=id,
        experiment_mappings=experiment_mappings
    )
    actual.experiment_mappings = new_experiment_mappings
    assert actual.to_open_api() == request.getfixturevalue(expected)