import json
import os.path
import random
from unittest.mock import patch, MagicMock
from urllib.parse import quote_plus
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
    turing_api, project, pyfunc_ensembler, use_google_oauth, active_project_magic_mock, experiment_name, experiment_id, artifact_uri, run_id, bucket_name
):
    turing.set_url(turing_api, use_google_oauth)
    
    turing_mock_request.return_value = active_project_magic_mock
    turing.set_project(project.name)
             
    mlflow_resp_1 = MagicMock()
    mlflow_resp_1.method = "GET"
    mlflow_resp_1.status_code = 200
    mlflow_resp_1.path = f"/api/2.0/mlflow/experiments/get-by-name?experiment_name={quote_plus(experiment_name)}"
    mlflow_resp_1.text = json.dumps(
            {
                "experiment": {
                    "id": experiment_id,
                    "name": experiment_name,
                    "lifecycle_stage": "active",
                }
            }
        )
    mlflow_resp_1.headers = {'Content-Type': 'application/json'}
    
    mlflow_resp_2 = MagicMock()
    mlflow_resp_2.method = "POST"
    mlflow_resp_2.status_code = 200
    mlflow_resp_2.path = "/api/2.0/mlflow/runs/create"
    mlflow_resp_2.text = json.dumps(
            {
                "run": {
                    "info": {
                        "run_id": run_id,
                        "experiment_id": experiment_id,
                        "status": "RUNNING",
                        "artifact_uri": artifact_uri,
                        "lifecycle_stage": "active",
                    },
                    "data": {},
                }
            }
        )
    mlflow_resp_2.headers = {'Content-Type': 'application/json'}
    
    mlflow_resp_3 = MagicMock()
    mlflow_resp_3.method = "GET"
    mlflow_resp_3.status_code = 200
    mlflow_resp_3.path = f"/api/2.0/mlflow/runs/get?run_uuid={run_id}&run_id={run_id}"
    mlflow_resp_3.text = json.dumps(
            {
                "run": {
                    "info": {
                        "run_id": run_id,
                        "experiment_id": experiment_id,
                        "status": "RUNNING",
                        "artifact_uri": artifact_uri,
                        "lifecycle_stage": "active",
                    },
                    "data": {},
                }
            }
        )
    mlflow_resp_3.headers = {'Content-Type': 'application/json'}
    
    mlflow_resp_4 = MagicMock()
    mlflow_resp_4.method = "POST"
    mlflow_resp_4.status_code = 200
    mlflow_resp_4.path = "/api/2.0/mlflow/runs/log-model"
    mlflow_resp_4.text = json.dumps({})
    mlflow_resp_4.headers = {'Content-Type': 'application/json'}
    
    mlflow_resp_5 = MagicMock()
    mlflow_resp_5.method = "POST"
    mlflow_resp_5.status_code = 200
    mlflow_resp_5.path = "/api/2.0/mlflow/runs/update"
    mlflow_resp_5.text = json.dumps({})
    mlflow_resp_5.headers = {'Content-Type': 'application/json'}
    
    mlflow_resp_6 = MagicMock()
    mlflow_resp_6.method = "POST"
    mlflow_resp_6.status_code = 200
    mlflow_resp_6.path = "/api/2.0/mlflow/runs/update"
    mlflow_resp_6.text = json.dumps({})
    mlflow_resp_6.headers = {'Content-Type': 'application/json'}
    
    mlflow_mock_request.side_effect = [mlflow_resp_1, mlflow_resp_2, mlflow_resp_3, mlflow_resp_4, mlflow_resp_5, mlflow_resp_6]
    
    os.environ[google.auth.environment_vars.PROJECT] = "test-project"
    
    gcs_resp_1 = MagicMock()
    gcs_resp_1.method = "POST"
    gcs_resp_1.status_code = 200
    gcs_resp_1.path = "/token"
    gcs_resp_1.text = json.dumps(
            {
                "access_token": "ya29.ImCpB6BS2mdOMseaUjhVlHqNfAOz168XjuDrK7Sd33glPd7XvtMLIngi1-V52ReytFSUluE-iBV88OlDkjtraggB_qc-LN2JlGtQ3sHZq_MuTxrU0-oK_kpq-1wsvniFFGQ",
                "expires_in": 3600,
                "scope": "openid https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email",
                "token_type": "Bearer",
                "id_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6IjhhNjNmZTcxZTUzMDY3NTI0Y2JiYzZhM2E1ODQ2M2IzODY0YzA3ODciLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiI3NjQwODYwNTE4NTAtNnFyNHA2Z3BpNmhuNTA2cHQ4ZWp1cTgzZGkzNDFodXIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJhdWQiOiI3NjQwODYwNTE4NTAtNnFyNHA2Z3BpNmhuNTA2cHQ4ZWp1cTgzZGkzNDFodXIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJzdWIiOiIxMDM5ODg0MzM2OTY3NzI1NDkzNjAiLCJoZCI6ImdvLWplay5jb20iLCJlbWFpbCI6InByYWRpdGh5YS5wdXJhQGdvLWplay5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiYXRfaGFzaCI6ImdrbXIxY0dPTzNsT0dZUDhtYjNJRnciLCJpYXQiOjE1NzE5Njg3NDUsImV4cCI6MTU3MTk3MjM0NX0.FIY5xvySNVxt1cbw-QXdDfiwollxcqupz1YDJuP14obKRyDwFG9ZcC_j-mTDZF5_dzpYeNMMK-LPTq9QIaM-blSKm2Eh9LeMvQGUk_S-9y_r2jKCmBlrEeHM8DXk3xyKf65LEoBA8cwMPdgb2s8AMIxxN9JJ09fjou20yLDI84Q4BFMriMIBBYLFgBW0wcg2PQ1hy5hrV1PdZj-ZNKNWmouh0lOjLLYmVFZPCzD9ENWo1N52ZLaLODdI2gDcpbyTUbeAh81sacdtJd0pLf-FuBLdfuktvP4MVvdmIhXv98Zb0dFBzRtmiqlQusSjoG5VEaBc6o2gkM5rHR0ozby0Fg",
            }
        )
    
    gcs_resp_2 = MagicMock()
    gcs_resp_2.method = "POST"
    gcs_resp_2.status_code = 200
    gcs_resp_2.path = f"/upload/storage/v1/b/{bucket_name}/o?uploadType=multipart"
    gcs_resp_2.text = json.dumps({})
    
    gcs_mock_request.side_effect = [gcs_resp_1, gcs_resp_2]
    
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