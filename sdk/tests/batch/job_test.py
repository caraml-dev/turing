import json
import os
from unittest.mock import MagicMock, patch

import pytest
from tests import utc_date
import tests
import turing
import turing.batch
import turing.batch.config

data_dir = os.path.join(os.path.dirname(__file__), "../testdata/api_responses")

with open(os.path.join(data_dir, "list_jobs_0000.json")) as f:
    list_jobs_0000 = f.read()

with open(os.path.join(data_dir, "submit_job_0000.json")) as f:
    submit_job_0000 = f.read()

with open(os.path.join(data_dir, "get_job_0000.json")) as f:
    get_job_0000 = f.read()

@pytest.mark.parametrize(
    "api_response, expected",
    [
        pytest.param(
            '{"results": [], "paging": {"total": 0, "page": 1, "pages": 1}}',
            [],
            id="Empty list response",
        ),
        pytest.param(
            list_jobs_0000,
            [
                turing.batch.EnsemblingJob(
                    id=11,
                    name="my-ensembler-updated: 2021-07-06T12:28:32+03:00",
                    ensembler_id=2,
                    status=turing.batch.EnsemblingJobStatus.PENDING,
                    project_id=1,
                    error="",
                    created_at=utc_date("2021-07-06T12:28:32.850365Z"),
                    updated_at=utc_date("2021-07-06T13:28:56.252642Z"),
                ),
                turing.batch.EnsemblingJob(
                    id=17,
                    name="my-ensembler: 2021-07-06T23:44:30+03:00",
                    ensembler_id=3,
                    status=turing.batch.EnsemblingJobStatus.FAILED_BUILDING,
                    project_id=1,
                    error="failed building OCI image",
                    created_at=utc_date("2021-07-06T23:44:30.675673Z"),
                    updated_at=utc_date("2021-07-07T07:36:33.604794Z"),
                ),
            ],
            id="Non empty list",
        ),
    ],
)
def test_list_jobs(
    turing_api, project, api_response, expected, use_google_oauth, active_project_magic_mock
):
    with patch("urllib3.PoolManager.request") as mock_request:
        turing.set_url(turing_api, use_google_oauth)
        
        mock_request.return_value = active_project_magic_mock
        turing.set_project(project.name)
        
        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/jobs?"
        f"status={turing.batch.EnsemblingJobStatus.PENDING.value}&"
        f"status={turing.batch.EnsemblingJobStatus.RUNNING.value}"
        mock_response.data = api_response.encode('utf-8')
        mock_response.getheader.return_value = 'application/json'
        
        mock_request.return_value = mock_response

        actual = turing.batch.EnsemblingJob.list(
            status=[
                turing.batch.EnsemblingJobStatus.PENDING,
                turing.batch.EnsemblingJobStatus.RUNNING,
            ]
        )

        assert len(actual) == len(expected)

        for actual, expected in zip(actual, expected):
            assert actual == expected
            
@pytest.mark.parametrize(
    "api_response, expected",
    [
        pytest.param(
            submit_job_0000,
            turing.batch.EnsemblingJob(
                id=1,
                name="pyfunc-ensembler: 2021-07-06T00:00:00+03:00",
                ensembler_id=2,
                status=turing.batch.EnsemblingJobStatus.PENDING,
                project_id=1,
                error="",
                created_at=utc_date("2021-07-06T12:28:32.850365Z"),
                updated_at=utc_date("2021-07-06T13:28:56.252642Z"),
            ),
        )
    ],
)
def test_submit_job(
    turing_api,
    project,
    ensembling_job_config,
    api_response,
    expected,
    use_google_oauth,
    active_project_magic_mock
):
    with patch("urllib3.PoolManager.request") as mock_request:
        turing.set_url(turing_api, use_google_oauth)

        mock_request.return_value = active_project_magic_mock
        turing.set_project(project.name)
        
        mock_response = MagicMock()
        mock_response.method = "POST"
        mock_response.status = 201
        mock_response.path = f"/v1/projects/{project.id}/jobs"
        mock_response.data = api_response.encode('utf-8')
        mock_response.getheader.return_value = 'application/json'

        mock_request.return_value = mock_response

        actual = turing.batch.job.EnsemblingJob.submit(
            ensembler_id=2,
            config=ensembling_job_config,
        )
        assert actual == expected

@pytest.mark.parametrize(
    "api_response_get, expected, api_response_refresh, updated",
    [
        pytest.param(
            submit_job_0000,
            turing.batch.EnsemblingJob(
                id=1,
                name="pyfunc-ensembler: 2021-07-06T00:00:00+03:00",
                ensembler_id=2,
                status=turing.batch.EnsemblingJobStatus.PENDING,
                project_id=1,
                error="",
                created_at=utc_date("2021-07-06T12:28:32.850365Z"),
                updated_at=utc_date("2021-07-06T13:28:56.252642Z"),
            ),
            get_job_0000,
            turing.batch.EnsemblingJob(
                id=1,
                name="pyfunc-ensembler: 2021-07-06T00:00:00+03:00",
                ensembler_id=2,
                status=turing.batch.EnsemblingJobStatus.FAILED_BUILDING,
                project_id=1,
                error="timeout has occurred",
                created_at=utc_date("2021-07-06T12:28:32.850365Z"),
                updated_at=utc_date("2021-07-07T00:00:00.252642Z"),
            ),
        )
    ],
)
def test_fetch_job(
    turing_api,
    project,
    api_response_get,
    expected,
    api_response_refresh,
    updated,
    use_google_oauth,
    active_project_magic_mock
):
    with patch("urllib3.PoolManager.request") as mock_request:
        turing.set_url(turing_api, use_google_oauth)

        mock_request.return_value = active_project_magic_mock
        turing.set_project(project.name)

        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/jobs/{expected.id}"
        mock_response.data = api_response_get.encode('utf-8')
        mock_response.getheader.return_value = 'application/json'
        
        mock_request.return_value = mock_response

        job = turing.batch.EnsemblingJob.get_by_id(expected.id)

        assert job == expected
        
        mock_response = MagicMock()
        mock_response.method = "GET"
        mock_response.status = 200
        mock_response.path = f"/v1/projects/{project.id}/jobs/{expected.id}"
        mock_response.data = api_response_refresh.encode('utf-8')
        mock_response.getheader.return_value = 'application/json'
        
        mock_request.return_value = mock_response

        job.refresh()

        assert job == updated
        
@pytest.mark.parametrize(
    "job, api_response_delete, api_response_get, expected",
    [
        pytest.param(
            turing.batch.EnsemblingJob(
                id=1,
                name="ensembling-job",
                ensembler_id=1,
                status=turing.batch.EnsemblingJobStatus.RUNNING,
                project_id=1,
                error="",
                created_at=utc_date("2021-07-06T12:28:32.850365Z"),
                updated_at=utc_date("2021-07-06T13:28:56.252642Z"),
            ),
            '{"id": 1}',
            get_job_0000,
            turing.batch.EnsemblingJob(
                id=1,
                name="pyfunc-ensembler: 2021-07-06T00:00:00+03:00",
                ensembler_id=2,
                status=turing.batch.EnsemblingJobStatus.FAILED_BUILDING,
                project_id=1,
                error="timeout has occurred",
                created_at=utc_date("2021-07-06T12:28:32.850365Z"),
                updated_at=utc_date("2021-07-07T00:00:00.252642Z"),
            ),
        )
    ],
)
def test_terminate_job(
    turing_api,
    project,
    job,
    api_response_delete,
    api_response_get,
    expected,
    use_google_oauth,
    active_project_magic_mock
):
    with patch("urllib3.PoolManager.request") as mock_request:
        turing.set_url(turing_api, use_google_oauth)

        mock_request.return_value = active_project_magic_mock
        turing.set_project(project.name)

        mock_response_1 = MagicMock()
        mock_response_1.method = "DELETE"
        mock_response_1.status = 201
        mock_response_1.path = f"/v1/projects/{project.id}/jobs/{job.id}"
        mock_response_1.data = api_response_delete.encode('utf-8')
        mock_response_1.getheader.return_value = 'application/json'
        
        mock_response_2 = MagicMock()
        mock_response_2.method = "get"
        mock_response_2.status = 200
        mock_response_2.path = f"/v1/projects/{project.id}/jobs/{job.id}"
        mock_response_2.data = api_response_get.encode('utf-8')
        mock_response_2.getheader.return_value = 'application/json'
        
        mock_request.side_effect = [mock_response_1, mock_response_2]

        assert job != expected

        job.terminate()

        assert job == expected