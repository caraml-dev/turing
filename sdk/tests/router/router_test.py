import os
import json
from unittest.mock import patch, MagicMock
import tests
import turing
import pytest
import turing.generated.models
from turing.router.config.router_version import RouterStatus

data_dir = os.path.join(os.path.dirname(__file__), "../testdata/api_responses")

with open(os.path.join(data_dir, "create_router_0000.json")) as f:
    create_router_0000 = f.read()

@pytest.mark.parametrize("num_routers", [2])
def test_list_routers(turing_api, project, generic_routers, use_google_oauth, active_project_magic_mock):
    with patch("urllib3.PoolManager.request") as mock_request:
        turing.set_url(turing_api, use_google_oauth)
        
        mock_request.return_value = active_project_magic_mock
        turing.set_project(project.name)
        
        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/routers"
        mock_response.data = json.dumps(generic_routers, default=tests.json_serializer).encode('utf-8')
        mock_response.getheader.return_value = 'application/json'
        
        mock_request.return_value = mock_response
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
