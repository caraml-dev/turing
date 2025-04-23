import json
from unittest.mock import MagicMock, patch
import tests
import turing
import pytest

@pytest.mark.parametrize("num_projects", [10])
def test_list_projects(turing_api, projects, use_google_oauth):
    with patch("urllib3.PoolManager.request") as mock_request:
        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects"
        mock_response.data = json.dumps(projects, default=tests.json_serializer).encode('utf-8')
        mock_response.getheader.return_value = 'application/json'
   
        mock_request.return_value = mock_response

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
        
        args, kwargs = mock_request.call_args_list[1]
        assert args[1] == "http://turing.local.svc:8080/v1/projects"
        assert kwargs.get("fields") == [('name', 'project_1')]
