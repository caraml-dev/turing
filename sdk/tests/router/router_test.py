import os
import json
import tests
import turing
import pytest
import turing.generated.models
from urllib3_mock import Responses
from turing.router.config.router_version import RouterStatus


responses = Responses('requests.packages.urllib3')
data_dir = os.path.join(os.path.dirname(__file__), "../testdata/api_responses")

with open(os.path.join(data_dir, "create_router_0000.json")) as f:
    create_router_0000 = f.read()


@pytest.fixture(scope="module", name="responses")
def _responses():
    return responses


@responses.activate
@pytest.mark.parametrize('num_routers', [2])
def test_list_routers(turing_api, active_project, generic_routers, use_google_oauth):
    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/routers",
        body=json.dumps(generic_routers, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)
    actual = turing.Router.list()

    assert len(actual) == len(generic_routers)
    assert all([isinstance(r, turing.Router) for r in actual])

    for actual, expected in zip(actual, generic_routers):
        assert actual.id == expected.id
        assert actual.name == expected.name
        assert actual.endpoint == expected.endpoint
        assert actual.environment_name == expected.environment_name
        assert actual.monitoring_url == expected.monitoring_url
        assert actual.project_id == expected.project_id
        assert actual.status.value == expected.status.value
        assert actual.created_at == expected.created_at
        assert actual.updated_at == expected.updated_at


@responses.activate
@pytest.mark.parametrize(
    "actual,expected", [
        pytest.param(
            "generic_router_config",
            create_router_0000
        )
    ]
)
def test_create_router(turing_api, active_project, actual, expected, use_google_oauth, request):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    router_config = request.getfixturevalue(actual)

    responses.add(
        method="POST",
        url=f"/v1/projects/{active_project.id}/routers",
        body=expected,
        status=200,
        content_type="application/json"
    )

    actual_response = turing.Router.create(router_config)
    actual_config = actual_response.config

    assert actual_config.environment_name == router_config.environment_name
    assert actual_config.name == router_config.name
    assert actual_config.rules == router_config.rules
    assert actual_config.default_route_id == router_config.default_route_id
    assert actual_config.experiment_engine.to_open_api() == router_config.experiment_engine.to_open_api()
    assert actual_config.resource_request.to_open_api() == router_config.resource_request.to_open_api()
    assert actual_config.timeout == router_config.timeout
    assert actual_config.log_config.to_open_api() == router_config.log_config.to_open_api()

    assert actual_config.enricher.image == router_config.enricher.image
    assert actual_config.enricher.resource_request.to_open_api() == router_config.enricher.resource_request.to_open_api()
    assert actual_config.enricher.endpoint == router_config.enricher.endpoint
    assert actual_config.enricher.timeout == router_config.enricher.timeout
    assert actual_config.enricher.port == router_config.enricher.port

    assert actual_config.ensembler.type == router_config.ensembler.type


@responses.activate
@pytest.mark.parametrize(
    "actual,expected", [
        pytest.param(
            1,
            turing.generated.models.InlineResponse200(id=1)
        )
    ]
)
def test_delete_router(turing_api, active_project, actual, expected, use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    responses.add(
        method="DELETE",
        url=f"/v1/projects/{active_project.id}/routers/{actual}",
        body=json.dumps(expected, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    response = turing.Router.delete(1)
    assert actual == response


@responses.activate
def test_get_router(turing_api, active_project, generic_router, use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    actual_id = 1

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/routers/{actual_id}",
        body=json.dumps(generic_router, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    response = turing.Router.get(actual_id)
    assert actual_id == response.id


@responses.activate
@pytest.mark.parametrize(
    "actual,expected", [
        pytest.param(
            "generic_router_config",
            create_router_0000
        )
    ]
)
def test_update_router(turing_api, active_project, actual, expected, use_google_oauth, request):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    base_router = turing.Router(
        id=1,
        name="router-1",
        project_id=active_project.id,
        environment_name = "id-dev",
        monitoring_url="http://localhost:5000/endpoint_1",
        status=turing.router.config.router_version.RouterStatus.DEPLOYED,
    )

    router_config = request.getfixturevalue(actual)

    responses.add(
        method="PUT",
        url=f"/v1/projects/{active_project.id}/routers/{base_router.id}",
        body=expected,
        status=200,
        content_type="application/json"
    )

    actual_response = base_router.update(router_config)
    actual_config = actual_response.config

    assert actual_config.environment_name == router_config.environment_name
    assert actual_config.name == router_config.name
    assert actual_config.rules == router_config.rules
    assert actual_config.default_route_id == router_config.default_route_id
    assert actual_config.experiment_engine.to_open_api() == router_config.experiment_engine.to_open_api()
    assert actual_config.resource_request.to_open_api() == router_config.resource_request.to_open_api()
    assert actual_config.timeout == router_config.timeout
    assert actual_config.log_config.to_open_api() == router_config.log_config.to_open_api()

    assert actual_config.enricher.image == router_config.enricher.image
    assert actual_config.enricher.resource_request.to_open_api() == router_config.enricher.resource_request.to_open_api()
    assert actual_config.enricher.endpoint == router_config.enricher.endpoint
    assert actual_config.enricher.timeout == router_config.enricher.timeout
    assert actual_config.enricher.port == router_config.enricher.port

    assert actual_config.ensembler.type == router_config.ensembler.type


@responses.activate
def test_deploy_router(turing_api, active_project, generic_router, generic_router_version, use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    base_router = turing.Router.from_open_api(generic_router)

    expected = turing.generated.models.InlineResponse202(
        router_id=1,
        version=1
    )

    responses.add(
        method="POST",
        url=f"/v1/projects/{active_project.id}/routers/{base_router.id}/deploy",
        body=json.dumps(expected, default=tests.json_serializer),
        status=202,
        content_type="application/json"
    )

    response = base_router.deploy()
    assert base_router.id == response['router_id']
    assert generic_router.config.version == response['version']


@responses.activate
def test_undeploy_router(turing_api, active_project, generic_router, generic_router_version, use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    base_router = turing.Router.from_open_api(generic_router)

    expected = turing.generated.models.InlineResponse2001(
        router_id=1,
    )

    responses.add(
        method="POST",
        url=f"/v1/projects/{active_project.id}/routers/{base_router.id}/undeploy",
        body=json.dumps(expected, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    response = base_router.undeploy()
    assert base_router.id == response['router_id']


@responses.activate
def test_list_versions(turing_api, active_project, generic_router, generic_router_version, use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    base_router = turing.Router.from_open_api(generic_router)

    expected_versions = [generic_router_version for _ in range(3)]

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/routers/{base_router.id}/versions",
        body=json.dumps(expected_versions, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    actual_versions = base_router.list_versions()

    assert len(actual_versions) == len(expected_versions)

    for actual, expected in zip(actual_versions, expected_versions):
        assert actual.id == generic_router_version.id
        assert actual.monitoring_url == generic_router_version.monitoring_url
        assert actual.status.value == generic_router_version.status.value
        assert actual.created_at == generic_router_version.created_at
        assert actual.updated_at == generic_router_version.updated_at


@responses.activate
def test_get_version(turing_api, active_project, generic_router, generic_router_version, use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    base_router = turing.Router.from_open_api(generic_router)

    actual_version = 1

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/routers/{base_router.id}/versions/{actual_version}",
        body=json.dumps(generic_router_version, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    actual_response = base_router.get_version(actual_version)

    assert actual_response.id == generic_router_version.id
    assert actual_response.monitoring_url == generic_router_version.monitoring_url
    assert actual_response.status.value == generic_router_version.status.value
    assert actual_response.created_at == generic_router_version.created_at
    assert actual_response.updated_at == generic_router_version.updated_at


@responses.activate
def test_get_version_config(turing_api, active_project, generic_router, generic_router_version, use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    base_router = turing.Router.from_open_api(generic_router)

    actual_version = 1

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/routers/{base_router.id}/versions/{actual_version}",
        body=json.dumps(generic_router_version, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    actual_response = base_router.get_version(actual_version).get_config()

    assert actual_response.environment_name == base_router.config.environment_name
    assert actual_response.name == base_router.config.name
    assert actual_response.rules == base_router.config.rules
    assert actual_response.default_route_id == base_router.config.default_route_id
    assert actual_response.experiment_engine.to_open_api() == base_router.config.experiment_engine.to_open_api()
    assert actual_response.resource_request.to_open_api() == base_router.config.resource_request.to_open_api()
    assert actual_response.timeout == base_router.config.timeout
    assert actual_response.log_config.to_open_api() == base_router.config.log_config.to_open_api()


@responses.activate
def test_delete_version(turing_api, active_project, generic_router, use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    base_router = turing.Router.from_open_api(generic_router)

    expected_router_id = 1
    expected_version = 1
    expected = turing.generated.models.InlineResponse202(
        router_id=expected_router_id,
        version=expected_version
    )

    responses.add(
        method="DELETE",
        url=f"/v1/projects/{active_project.id}/routers/{base_router.id}/versions/{expected_version}",
        body=json.dumps(expected, default=tests.json_serializer),
        status=202,
        content_type="application/json"
    )

    response = base_router.delete_version(1)
    assert base_router.id == response['router_id']
    assert generic_router.config.version == response['version']


@responses.activate
def test_deploy_version(turing_api, active_project, generic_router, use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    base_router = turing.Router.from_open_api(generic_router)

    expected_router_id = 1
    expected_version = 1
    expected = turing.generated.models.InlineResponse202(
        router_id=expected_router_id,
        version=expected_version
    )

    responses.add(
        method="POST",
        url=f"/v1/projects/{active_project.id}/routers/{base_router.id}/versions/{expected_version}/deploy",
        body=json.dumps(expected, default=tests.json_serializer),
        status=202,
        content_type="application/json"
    )

    response = base_router.deploy_version(1)
    assert base_router.id == response['router_id']
    assert generic_router.config.version == response['version']


@responses.activate
def test_get_events_list(turing_api, active_project, generic_router, generic_events, use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    base_router = turing.Router.from_open_api(generic_router)

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/routers/{base_router.id}/events",
        body=json.dumps(generic_events, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    response = base_router.get_events()
    expected_events = generic_events.get('events')

    assert len(response) == len(expected_events)

    for actual, expected in zip(response, expected_events):
        assert actual.id == expected.id
        assert actual.version == expected.version
        assert actual.event_type == expected.event_type
        assert actual.stage == expected.stage
        assert actual.message == expected.message
        assert actual.created_at == expected.created_at
        assert actual.updated_at == expected.updated_at


@responses.activate
@pytest.mark.parametrize(
    "status,max_tries,duration,expected", [
        pytest.param(
            RouterStatus.DEPLOYED,
            1,
            1,
            TimeoutError
        )
    ]
)
def test_wait_for_status(
        turing_api,
        active_project,
        generic_router,
        status,
        max_tries,
        duration,
        expected,
        use_google_oauth
):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    base_router = turing.Router.from_open_api(generic_router)
    generic_router.status = turing.generated.models.RouterStatus('pending')

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/routers/{base_router.id}",
        body=json.dumps(generic_router, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    with pytest.raises(expected):
        base_router.wait_for_status(status, max_tries=max_tries, duration=duration)


@responses.activate
@pytest.mark.parametrize(
    "version,status,max_tries,duration,expected", [
        pytest.param(
            1,
            RouterStatus.DEPLOYED,
            1,
            1,
            TimeoutError
        )
    ]
)
def test_wait_for_version_status(
        turing_api,
        active_project,
        generic_router,
        generic_router_version,
        version,
        status,
        max_tries,
        duration,
        expected,
        use_google_oauth
):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    base_router = turing.Router.from_open_api(generic_router)
    generic_router_version.status = turing.generated.models.RouterVersionStatus('pending')

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/routers/{base_router.id}/versions/{version}",
        body=json.dumps(generic_router_version, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    with pytest.raises(expected):
        base_router.wait_for_version_status(status, version=version, max_tries=max_tries, duration=duration)
