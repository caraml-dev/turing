import json
import tests
import turing
import pytest
from urllib3_mock import Responses

responses = Responses('requests.packages.urllib3')


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
