import pytest
import turing.generated.models
from turing.generated.exceptions import ApiValueError
from turing.router.config.common.env_var import EnvVar
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.route import InvalidRouteException
from turing.router.config.router_ensembler_config import (RouterEnsemblerConfig,
                                                          EnsemblerNopConfig,
                                                          NopRouterEnsemblerConfig,
                                                          PyfuncRouterEnsemblerConfig,
                                                          DockerRouterEnsemblerConfig,
                                                          StandardRouterEnsemblerConfig,
                                                          InvalidExperimentMappingException)

@pytest.mark.parametrize(
    "id,type,standard_config,docker_config,expected", [
        pytest.param(
            1,
            "standard",
            "standard_router_ensembler_config",
            None,
            "generic_standard_router_ensembler_config"
        ),
        pytest.param(
            1,
            "docker",
            None,
            "generic_ensembler_docker_config",
            "generic_docker_router_ensembler_config"
        )
    ])
def test_create_router_ensembler_config(id, type, standard_config, docker_config, expected, request):
    docker_config_data = None if docker_config is None else request.getfixturevalue(docker_config)
    standard_config_data = None if standard_config is None else request.getfixturevalue(standard_config)
    actual = RouterEnsemblerConfig(
        id=id,
        type=type,
        standard_config=standard_config_data,
        docker_config=docker_config_data
    ).to_open_api()
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "project_id,ensembler_id,resource_request,timeout,env,expected", [
        pytest.param(
            77,
            11,
            ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            "500ms",
            [
                EnvVar(
                    name="env_name",
                    value="env_val")
            ],
            "generic_pyfunc_router_ensembler_config"
        )
    ])
def test_create_pyfunc_router_ensembler_config(
        project_id,
        ensembler_id,
        resource_request,
        timeout,
        env,
        expected,
        request
):
    actual = PyfuncRouterEnsemblerConfig(
        project_id=project_id,
        ensembler_id=ensembler_id,
        resource_request=resource_request,
        timeout=timeout,
        env=env
    ).to_open_api()
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "project_id,ensembler_id,resource_request,timeout,env,expected", [
        pytest.param(
            77,
            11,
            ResourceRequest(
                min_replica=1,
                max_replica=3,
                cpu_request='100m',
                memory_request='512Mi'
            ),
            "500ks",
            [
                EnvVar(
                    name="env_name",
                    value="env_val")
            ],
            ApiValueError
        )
    ])
def test_create_pyfunc_router_ensembler_config_with_invalid_timeout(
        project_id,
        ensembler_id,
        resource_request,
        timeout,
        env,
        expected
):
    with pytest.raises(expected):
        PyfuncRouterEnsemblerConfig(
            project_id=project_id,
            ensembler_id=ensembler_id,
            resource_request=resource_request,
            timeout=timeout,
            env=env
        ).to_open_api()


@pytest.mark.parametrize(
    "image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            "test.io/just-a-test/turing-ensembler:0.0.0-build.0",
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
    "image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
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
            image=image,
            resource_request=resource_request,
            endpoint=endpoint,
            timeout=timeout,
            port=port,
            env=env,
            service_account=service_account
        ).to_open_api()


@pytest.mark.parametrize(
    "image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            "test.io/just-a-test/turing-ensembler:0.0.0-build.0",
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
            image=image,
            resource_request=resource_request,
            endpoint=endpoint,
            timeout=timeout,
            port=port,
            env=env,
            service_account=service_account
        ).to_open_api()


@pytest.mark.parametrize(
    "image,resource_request,endpoint,timeout,port,env,service_account,expected", [
        pytest.param(
            "test.io/just-a-test/turing-ensembler:0.0.0-build.0",
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
            image=image,
            resource_request=resource_request,
            endpoint=endpoint,
            timeout=timeout,
            port=port,
            env=env,
            service_account=service_account
        ).to_open_api()


@pytest.mark.parametrize(
    "experiment_mappings,fallback_response_route_id,expected", [
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
            "route-1",
            "generic_standard_router_ensembler_config"
        )
    ])
def test_create_standard_router_ensembler_config(experiment_mappings, fallback_response_route_id, expected, request):
    actual = StandardRouterEnsemblerConfig(
        experiment_mappings=experiment_mappings,
        fallback_response_route_id=fallback_response_route_id,
    )
    assert actual.to_open_api() == request.getfixturevalue(expected)
    assert actual.standard_config.fallback_response_route_id == fallback_response_route_id


@pytest.mark.parametrize(
    "new_experiment_mappings,experiment_mappings,fallback_response_route_id,expected", [
        pytest.param(
            [
                {
                    "experiment": "wrong-experiment"
                }
            ],
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
            "route-1",
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
            "route-1",
            InvalidExperimentMappingException
        )
    ])
def test_set_standard_router_ensembler_config_with_invalid_experiment_mappings(
        new_experiment_mappings,
        experiment_mappings,
        fallback_response_route_id,
        expected):
    actual = StandardRouterEnsemblerConfig(
        experiment_mappings=experiment_mappings,
        fallback_response_route_id=fallback_response_route_id
    )
    with pytest.raises(expected):
        actual.experiment_mappings = new_experiment_mappings


@pytest.mark.parametrize(
    "new_experiment_mappings,experiment_mappings,fallback_response_route_id,expected", [
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
            [
                {
                    "experiment": "wrong-experiment",
                    "treatment": "wrong-treatment",
                    "route": "wrong-route"
                }
            ],
            "route-1",
            "generic_standard_router_ensembler_config"
        )
    ])
def test_set_standard_router_ensembler_config_with_valid_experiment_mappings(
        new_experiment_mappings,
        experiment_mappings,
        fallback_response_route_id,
        expected,
        request):
    actual = StandardRouterEnsemblerConfig(
        experiment_mappings=experiment_mappings,
        fallback_response_route_id=fallback_response_route_id
    )
    actual.experiment_mappings = new_experiment_mappings
    assert actual.to_open_api() == request.getfixturevalue(expected)
    assert actual.standard_config.fallback_response_route_id == fallback_response_route_id


@pytest.mark.parametrize(
    "final_response_route_id,nop_config,expected", [
        pytest.param(
            "test-route",
            EnsemblerNopConfig(final_response_route_id="test-route"),
            None
        )
    ])
def test_create_nop_router_ensembler_config(
    final_response_route_id,
    nop_config,
    expected):
    ensembler = NopRouterEnsemblerConfig(final_response_route_id=final_response_route_id)
    assert ensembler.to_open_api() == expected
    assert ensembler.nop_config == nop_config

@pytest.mark.parametrize(
    "router_config,ensembler_config", [
        pytest.param(
            "generic_router_config",
           NopRouterEnsemblerConfig(final_response_route_id="model-b"),
        )
    ])
def test_copy_nop_ensembler_default_route(
    router_config,
    ensembler_config,
    request):
    router = request.getfixturevalue(router_config)
    # Check precondition
    assert router.default_route_id != ensembler_config.final_response_route_id
    
    router.ensembler = ensembler_config
    actual = router.to_open_api()
    router.default_route_id = ensembler_config.final_response_route_id
    expected = router.to_open_api()
    assert actual == expected

@pytest.mark.parametrize(
    "router_config,ensembler_config", [
        pytest.param(
            "generic_router_config",
           StandardRouterEnsemblerConfig(
               experiment_mappings=[],
               fallback_response_route_id="model-b",
           )
        )
    ])
def test_copy_standard_ensembler_default_route(
    router_config,
    ensembler_config,
    request):
    router = request.getfixturevalue(router_config)
    # Check precondition
    assert router.default_route_id != ensembler_config.fallback_response_route_id
    
    router.ensembler = ensembler_config
    actual = router.to_open_api()
    router.default_route_id = ensembler_config.fallback_response_route_id
    expected = router.to_open_api()
    assert actual == expected

@pytest.mark.parametrize(
    "router_config,ensembler_config,expected", [
        pytest.param(
            "generic_router_config",
           NopRouterEnsemblerConfig(final_response_route_id="test-route-not-exists"),
           InvalidRouteException,
        )
    ])
def test_create_nop_router_ensembler_config_with_invalid_route(
    router_config,
    ensembler_config,
    expected,
    request):
    router = request.getfixturevalue(router_config)
    router.ensembler = ensembler_config
    with pytest.raises(expected):
       router.to_open_api()

@pytest.mark.parametrize(
    "cls,config,expected", [
        pytest.param(
            NopRouterEnsemblerConfig,
            "nop_router_ensembler_config",
            {"type":"nop", "final_response_route_id": "test"}
        ),
        pytest.param(
            StandardRouterEnsemblerConfig,
            "standard_router_ensembler_config",
            {
                "type":"standard",
                "experiment_mappings": [
                    {
                        "experiment": "experiment-1",
                        "treatment": "treatment-1",
                        "route": "route-1"
                    },
                    {
                        "experiment": "experiment-2",
                        "treatment": "treatment-2",
                        "route": "route-2"
                    }
                ],
                "fallback_response_route_id": "route-1"
            }
        ),
        pytest.param(
            DockerRouterEnsemblerConfig,
            "generic_ensembler_docker_config",
            {
                "type": "docker",
                "image": "test.io/just-a-test/turing-ensembler:0.0.0-build.0",
                "resource_request": ResourceRequest(
                    min_replica=1,
                    max_replica=3,
                    cpu_request='100m',
                    memory_request='512Mi'
                ),
                "endpoint": "http://localhost:5000/ensembler_endpoint",
                "timeout": "500ms",
                "port": 5120,
                "env": [EnvVar(name="env_name", value="env_val")],
                "service_account": "secret-name-for-google-service-account"
            }
        ),
        pytest.param(
            PyfuncRouterEnsemblerConfig,
            "generic_ensembler_pyfunc_config",
            {
                "type":"pyfunc",
                "project_id": 77,
                "ensembler_id": 11,
                "resource_request": ResourceRequest(
                    min_replica=1,
                    max_replica=3,
                    cpu_request='100m',
                    memory_request='512Mi'
                ),
                "timeout": "500ms",
                "env": [EnvVar(name="env_name", value="env_val")]
            }
        )
    ])
def test_create_ensembler_config_from_config(
    cls, config, expected, request
):
    config_data = request.getfixturevalue(config)
    assert cls.from_config(config_data).to_dict() == expected

def test_set_nop_ensembler_config_with_default_route(request):
    router = request.getfixturevalue("generic_router_config")
    router.ensembler = None
    router.default_route_id = "model-b"
    assert router.ensembler.final_response_route_id == "model-b"

def test_set_standard_ensembler_config_with_default_route(request):
    router = request.getfixturevalue("generic_router_config")
    router.ensembler = StandardRouterEnsemblerConfig(
        experiment_mappings=[],
        fallback_response_route_id=""
    )
    router.default_route_id = "model-b"
    assert router.ensembler.fallback_response_route_id == "model-b"
