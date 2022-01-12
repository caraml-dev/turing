import os
import json
import tests
import turing
import pytest
from urllib3_mock import Responses

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

    actual_response = turing.Router.create(router_config.to_open_api())
    actual_config = actual_response.config

    assert actual_config.environment_name == router_config.environment_name

    assert actual_config.name == router_config.name

    assert actual_config.rules == router_config.rules

    assert actual_config.default_route_id == router_config.default_route_id

    assert actual_config.experiment_engine.type == router_config.experiment_engine.type
    assert actual_config.experiment_engine.config == router_config.experiment_engine.config

    assert actual_config.timeout == router_config.timeout

    assert actual_config.log_config.result_logger_type == router_config.log_config.result_logger_type
    assert actual_config.log_config.bigquery_config == router_config.log_config.bigquery_config
    assert actual_config.log_config.kafka_config == router_config.log_config.kafka_config

    assert actual_config.enricher.image == router_config.enricher.image
    assert actual_config.enricher.endpoint == router_config.enricher.endpoint
    assert actual_config.enricher.timeout == router_config.enricher.timeout
    assert actual_config.enricher.port == router_config.enricher.port

    assert actual_config.ensembler.type == router_config.ensembler.type
