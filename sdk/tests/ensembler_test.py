import json
import os.path
import random
from unittest.mock import patch, MagicMock
import google.auth.environment_vars

import pandas
import pytest

import tests
import turing.ensembler
import turing.generated.models

data_dir = os.path.join(os.path.dirname(__file__), "./testdata/api_responses")

with open(os.path.join(data_dir, "list_jobs_0000.json")) as f:
    list_jobs_0000 = f.read()

def test_predict():
    default_value = random.random()
    ensembler = tests.MyTestEnsembler(default_value)

    model_input = pandas.DataFrame(
        data={
            "treatment": ["model_a", "model_b", "unknown"],
            f"{turing.ensembler.PyFunc.PREDICTION_COLUMN_PREFIX}model_a": [
                0.01,
                0.2,
                None,
            ],
            f"{turing.ensembler.PyFunc.PREDICTION_COLUMN_PREFIX}model_b": [
                0.03,
                0.6,
                0.4,
            ],
        }
    )

    expected = pandas.Series(data=[0.01, 0.6, default_value])
    result = ensembler.predict(context=None, model_input=model_input)

    from pandas._testing import assert_series_equal

    assert_series_equal(expected, result)

@pytest.mark.parametrize("num_ensemblers", [6])
def test_list_ensemblers(
    turing_api, project, generic_ensemblers, use_google_oauth, active_project_magic_mock
):
    with patch("urllib3.PoolManager.request") as mock_request:
        turing.set_url(turing_api, use_google_oauth)
        
        mock_request.return_value = active_project_magic_mock
        turing.set_project(project.name)

        page = turing.generated.models.EnsemblersPaginatedResults(
            results=generic_ensemblers,
            paging=turing.generated.models.PaginationPaging(total=1, page=1, pages=1),
        )
        
        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/ensemblers?type={turing.PyFuncEnsembler.TYPE.value}"
        mock_response.data = json.dumps(page, default=tests.json_serializer).encode('utf-8')
        mock_response.getheader.return_value = 'application/json'
        
        mock_request.return_value = mock_response

        actual = turing.PyFuncEnsembler.list()
        assert all([isinstance(p, turing.PyFuncEnsembler) for p in actual])

        for actual, expected in zip(actual, generic_ensemblers):
            assert actual == turing.PyFuncEnsembler.from_open_api(expected)

@patch("google.cloud.storage.Client")
@patch("requests.Session.request")
@patch("urllib3.PoolManager.request")
@pytest.mark.parametrize("ensembler_name", ["ensembler_1"])
def test_create_ensembler(
    turing_mock_request, mlflow_mock_request, gcs_mock_request,
    turing_api, project, pyfunc_ensembler, use_google_oauth, active_project_magic_mock, ensembler_mlflow_magic_mock_sequence, ensembler_gcs_magic_mock_sequence
):
    turing.set_url(turing_api, use_google_oauth)
    
    turing_mock_request.return_value = active_project_magic_mock
    turing.set_project(project.name)
    
    mlflow_mock_request.side_effect = ensembler_mlflow_magic_mock_sequence
    
    os.environ[google.auth.environment_vars.PROJECT] = "test-project"
    
    gcs_mock_request.side_effect = ensembler_gcs_magic_mock_sequence
    
    turing_resp = MagicMock()
    turing_resp.method = "POST"
    turing_resp.status = 201
    turing_resp.path = f"/v1/projects/{project.id}/ensemblers"
    turing_resp.data = json.dumps(pyfunc_ensembler, default=tests.json_serializer).encode('utf-8')
    turing_resp.getheader.return_value = 'application/json'
    
    turing_mock_request.return_value = turing_resp

    actual = turing.PyFuncEnsembler.create(
        name=pyfunc_ensembler.name,
        ensembler_instance=tests.MyTestEnsembler(0.01),
        conda_env={
            "channels": ["defaults"],
            "dependencies": ["python=3.9.0", {"pip": ["test-lib==0.0.1"]}],
        },
    )
    
    assert actual == turing.PyFuncEnsembler.from_open_api(pyfunc_ensembler)
    
@pytest.mark.parametrize(("num_ensemblers", "ensembler_name"), [(3, "updated")])
@patch("google.cloud.storage.Client")
@patch("requests.Session.request")
@patch("urllib3.PoolManager.request")
def test_update_ensembler(
    turing_mock_request, mlflow_mock_request, gcs_mock_request,
    turing_api, project, generic_ensemblers, pyfunc_ensembler, use_google_oauth, active_project_magic_mock, ensembler_mlflow_magic_mock_sequence, ensembler_gcs_magic_mock_sequence
):
    turing.set_url(turing_api, use_google_oauth)
    
    turing_mock_request.return_value = active_project_magic_mock
    turing.set_project(project.name)

    page = turing.generated.models.EnsemblersPaginatedResults(
        results=generic_ensemblers,
        paging=turing.generated.models.PaginationPaging(total=1, page=1, pages=1),
    )

    emptyJob = turing.generated.models.EnsemblingJobPaginatedResults(
        results=[],
        paging=turing.generated.models.PaginationPaging(total=0, page=1, pages=1),
    )
    
    mlflow_mock_request.side_effect = ensembler_mlflow_magic_mock_sequence
    
    os.environ[google.auth.environment_vars.PROJECT] = "test-project"
    
    gcs_mock_request.side_effect = ensembler_gcs_magic_mock_sequence

    turing_resp = MagicMock()
    turing_resp.method = "GET"
    turing_resp.status = 200
    turing_resp.path = f"/v1/projects/{project.id}/ensemblers"
    turing_resp.data = json.dumps(page, default=tests.json_serializer).encode('utf-8')
    turing_resp.getheader.return_value = 'application/json'
    
    turing_mock_request.return_value = turing_resp

    actual, *rest = turing.PyFuncEnsembler.list()

    turing_resp_1 = MagicMock()
    turing_resp_1.method = "GET"
    turing_resp_1.status = 200
    turing_resp_1.path = f"/v1/projects/{project.id}/router-versions"
    turing_resp_1.data = json.dumps([], default=tests.json_serializer).encode('utf-8')
    turing_resp_1.getheader.return_value = 'application/json'

    turing_resp_2 = MagicMock()
    turing_resp_2.method = "GET"
    turing_resp_2.status = 200
    turing_resp_2.path = f"/v1/projects/{project.id}/jobs"
    turing_resp_2.data = json.dumps(emptyJob, default=tests.json_serializer).encode('utf-8')
    turing_resp_2.getheader.return_value = 'application/json'
    
    turing_resp_3 = MagicMock()
    turing_resp_3.method = "PUT"
    turing_resp_3.status = 200
    turing_resp_3.path = f"/v1/projects/{project.id}/ensemblers/{actual.id}"
    turing_resp_3.data = json.dumps(pyfunc_ensembler, default=tests.json_serializer).encode('utf-8')
    turing_resp_3.getheader.return_value = 'application/json'
    
    turing_mock_request.side_effect = [turing_resp_1, turing_resp_2, turing_resp_3]

    actual.update(
        name=pyfunc_ensembler.name,
        ensembler_instance=tests.MyTestEnsembler(0.06),
        conda_env={
            "channels": ["defaults"],
            "dependencies": ["python>=3.8.0", {"pip": ["test-lib==0.0.1"]}],
        },
        code_dir=[
            os.path.join(
                os.path.dirname(os.path.realpath(__file__)), "..", "samples/quickstart"
            )
        ],
    )
    
    assert actual == turing.PyFuncEnsembler.from_open_api(pyfunc_ensembler)
    
@pytest.mark.parametrize(("num_ensemblers", "ensembler_name"), [(3, "updated")])
def test_update_ensembler_existing_router_version(
    turing_api,
    project,
    generic_ensemblers,
    pyfunc_ensembler,
    use_google_oauth,
    generic_router_version,
    active_project_magic_mock
):
    with patch("urllib3.PoolManager.request") as mock_request:
        turing.set_url(turing_api, use_google_oauth)

        mock_request.return_value = active_project_magic_mock
        turing.set_project(project.name)

        page = turing.generated.models.EnsemblersPaginatedResults(
            results=generic_ensemblers,
            paging=turing.generated.models.PaginationPaging(total=1, page=1, pages=1),
        )
        
        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/ensemblers"
        mock_response.data = json.dumps(page, default=tests.json_serializer).encode('utf-8')
        mock_response.getheader.return_value = 'application/json'
        
        mock_request.return_value = mock_response

        actual, *rest = turing.PyFuncEnsembler.list()
        
        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/router-versions"
        mock_response.data = json.dumps([generic_router_version], default=tests.json_serializer).encode('utf-8')
        mock_response.getheader.return_value = 'application/json'
        
        mock_request.return_value = mock_response

        with pytest.raises(ValueError) as error:
            actual.update(
                name=pyfunc_ensembler.name,
                ensembler_instance=tests.MyTestEnsembler(0.06),
                conda_env={
                    "channels": ["defaults"],
                    "dependencies": ["python>=3.8.0", {"pip": ["test-lib==0.0.1"]}],
                },
                code_dir=[
                    os.path.join(
                        os.path.dirname(os.path.realpath(__file__)),
                        "..",
                        "samples/quickstart",
                    )
                ],
            )
        expected_error_message = "There is pending router version using this ensembler. Please wait for the router version to be deployed or undeploy it, before updating the ensembler."
        actual_error_message = str(error.value)
        assert expected_error_message == actual_error_message
        
@pytest.mark.parametrize(("num_ensemblers", "ensembler_name"), [(3, "updated")])
def test_update_ensembler_existing_job(
    turing_api, project, generic_ensemblers, pyfunc_ensembler, use_google_oauth, active_project_magic_mock
):
    with patch("urllib3.PoolManager.request") as mock_request:
        turing.set_url(turing_api, use_google_oauth)

        mock_request.return_value = active_project_magic_mock
        turing.set_project(project.name)

        page = turing.generated.models.EnsemblersPaginatedResults(
            results=generic_ensemblers,
            paging=turing.generated.models.PaginationPaging(total=1, page=1, pages=1),
        )
        
        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/ensemblers"
        mock_response.data = json.dumps(page, default=tests.json_serializer).encode('utf-8')
        mock_response.getheader.return_value = 'application/json'
        
        mock_request.return_value = mock_response

        actual, *rest = turing.PyFuncEnsembler.list()
        
        mock_response_1 = MagicMock()
        mock_response_1.method = "GET"
        mock_response_1.status = 200
        mock_response_1.path = f"/v1/projects/{project.id}/router-versions"
        mock_response_1.data = json.dumps([], default=tests.json_serializer).encode('utf-8')
        mock_response_1.getheader.return_value = 'application/json'

        mock_response_2 = MagicMock()
        mock_response_2.method = "GET"
        mock_response_2.status = 200
        mock_response_2.path = f"/v1/projects/{project.id}/jobs"
        mock_response_2.data = list_jobs_0000.encode('utf-8')
        mock_response_2.getheader.return_value = 'application/json'

        mock_request.side_effect = [mock_response_1, mock_response_2]
        
        with pytest.raises(ValueError) as error:
            actual.update(
                name=pyfunc_ensembler.name,
                ensembler_instance=tests.MyTestEnsembler(0.06),
                conda_env={
                    "channels": ["defaults"],
                    "dependencies": ["python>=3.8.0", {"pip": ["test-lib==0.0.1"]}],
                },
                code_dir=[
                    os.path.join(
                        os.path.dirname(os.path.realpath(__file__)),
                        "..",
                        "samples/quickstart",
                    )
                ],
            )
        expected_error_message = "There is pending ensembling job using this ensembler. Please wait for the ensembling job to be completed or terminate it, before updating the ensembler."
        actual_error_message = str(error.value)
        assert expected_error_message == actual_error_message
        
@patch("google.cloud.storage.Client")
@patch("requests.Session.request")
@patch("urllib3.PoolManager.request")
@pytest.mark.parametrize(
    "actual,expected", [pytest.param(1, turing.generated.models.IdObject(id=1))]
)
@pytest.mark.parametrize("ensembler_name", ["ensembler_1"])
def test_delete_ensembler(
    turing_mock_request, mlflow_mock_request, gcs_mock_request,
    turing_api, project, use_google_oauth, actual, expected, active_project_magic_mock, ensembler_mlflow_magic_mock_sequence, ensembler_gcs_magic_mock_sequence
):
    turing.set_url(turing_api, use_google_oauth)

    turing_mock_request.return_value = active_project_magic_mock
    turing.set_project(project.name)
    
    mlflow_mock_request.side_effect = ensembler_mlflow_magic_mock_sequence
    gcs_mock_request.side_effect = ensembler_gcs_magic_mock_sequence
    
    mock_response = MagicMock()
    mock_response.method = "DELETE"
    mock_response.status = 200
    mock_response.path = f"/v1/projects/{project.id}/ensemblers/{actual}"
    mock_response.data = json.dumps(expected, default=tests.json_serializer).encode('utf-8')
    mock_response.getheader.return_value = 'application/json'
    
    turing_mock_request.return_value = mock_response

    response = turing.PyFuncEnsembler.delete(1)

    assert actual == response