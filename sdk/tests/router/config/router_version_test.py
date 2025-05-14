import json
from unittest.mock import patch, MagicMock
import tests
import turing

from turing.router.config.router_version import RouterVersion

def test_create_version(
    turing_api,
    project,
    generic_router_config,
    generic_router_version,
    use_google_oauth,
    active_project_magic_mock
):
    with patch("urllib3.PoolManager.request") as mock_request:
        turing.set_url(turing_api, use_google_oauth)
        
        mock_request.return_value = active_project_magic_mock
        turing.set_project(project.name)

        test_router_id = 1
        
        mock_response = MagicMock()
        mock_response.method = "POST"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/routers/{test_router_id}/versions"
        mock_response.data = json.dumps(generic_router_version, default=tests.json_serializer).encode('utf-8')
        mock_response.getheader.return_value = 'application/json'
        
        mock_request.return_value = mock_response

        actual_response = RouterVersion.create(generic_router_config, test_router_id)

        assert actual_response.id == generic_router_version.id
        assert actual_response.monitoring_url == generic_router_version.monitoring_url
        assert actual_response.status.value == generic_router_version.status.value
        assert actual_response.created_at == generic_router_version.created_at
        assert actual_response.updated_at == generic_router_version.updated_at

        assert len(actual_response.rules) == len(generic_router_version.rules)
        for i in range((len(actual_response.rules))):
            assert actual_response.rules[i].to_open_api() == generic_router_version.rules[i]

        assert actual_response.default_route_id == generic_router_version.default_route_id

        # Assert the response against an experiment engine object that has its passkey removed
        expected_experiment_engine = generic_router_version.experiment_engine
        if (
            expected_experiment_engine.config is not None
            and expected_experiment_engine.config.get("client") is not None
            and expected_experiment_engine.config["client"].get("passkey") != ""
        ):
            expected_experiment_engine.config["client"]["passkey"] = None
        assert actual_response.experiment_engine.to_open_api() == expected_experiment_engine
        assert (
            actual_response.resource_request.to_open_api()
            == generic_router_version.resource_request
        )
        assert actual_response.timeout == generic_router_version.timeout

        assert (
            actual_response.log_config.to_open_api().result_logger_type
            == generic_router_version.log_config.result_logger_type
        )

        assert actual_response.enricher.image == generic_router_version.enricher.image
        assert (
            actual_response.enricher.resource_request.to_open_api()
            == generic_router_version.enricher.resource_request
        )
        assert actual_response.enricher.endpoint == generic_router_version.enricher.endpoint
        assert actual_response.enricher.timeout == generic_router_version.enricher.timeout
        assert actual_response.enricher.port == generic_router_version.enricher.port

        assert actual_response.ensembler.type == generic_router_version.ensembler.type
