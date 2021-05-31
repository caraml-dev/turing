import json
import tests
import turing
import pytest
from urllib3_mock import Responses

responses = Responses('requests.packages.urllib3')


@responses.activate
@pytest.mark.parametrize('num_projects', [10])
def test_list_projects(turing_api, projects, use_google_oauth):
    responses.add(
        method="GET",
        url=f"/v1/projects",
        body=json.dumps(projects, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    turing.set_url(turing_api, use_google_oauth)
    actual = turing.Project.list()

    assert len(actual) == len(projects)
    assert all([isinstance(p, turing.Project) for p in actual])

    for actual, expected in zip(actual, projects):
        assert actual.id == expected.id
        assert actual.name == expected.name
        assert actual.mlflow_tracking_url == expected.mlflow_tracking_url
        assert actual.created_at == expected.created_at
        assert actual.updated_at == expected.updated_at

    turing.Project.list(name="project_1")
    assert responses.calls[1].request.url == "/v1/projects?name=project_1"
