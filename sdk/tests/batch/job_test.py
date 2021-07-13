import os
import pytest
from tests import utc_date
import turing
import turing.batch
import turing.batch.config
from urllib3_mock import Responses

responses = Responses('requests.packages.urllib3')
data_dir = os.path.join(os.path.dirname(__file__), "../testdata/api_responses")

with open(os.path.join(data_dir, "list_jobs_0000.json")) as f:
    list_jobs_0000 = f.read()

with open(os.path.join(data_dir, "submit_job_0000.json")) as f:
    submit_job_0000 = f.read()

with open(os.path.join(data_dir, "get_job_0000.json")) as f:
    get_job_0000 = f.read()


@pytest.fixture(scope="module", name="responses")
def _responses():
    return responses


@responses.activate
@pytest.mark.parametrize(
    "api_response, expected", [
        pytest.param(
            '{"results": [], "paging": {"total": 0, "page": 1, "pages": 1}}',
            [],
            id="Empty list response"
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
                    updated_at=utc_date("2021-07-06T13:28:56.252642Z")
                ),
                turing.batch.EnsemblingJob(
                    id=17,
                    name="my-ensembler: 2021-07-06T23:44:30+03:00",
                    ensembler_id=3,
                    status=turing.batch.EnsemblingJobStatus.FAILED_BUILDING,
                    project_id=1,
                    error="failed building OCI image",
                    created_at=utc_date("2021-07-06T23:44:30.675673Z"),
                    updated_at=utc_date("2021-07-07T07:36:33.604794Z")
                )
            ],
            id="Non empty list"
        )
    ]
)
def test_list_jobs(turing_api, active_project, api_response, expected, use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/jobs?"
            f"status={turing.batch.EnsemblingJobStatus.PENDING.value}&"
            f"status={turing.batch.EnsemblingJobStatus.RUNNING.value}",
        body=api_response,
        match_querystring=True,
        status=200,
        content_type="application/json"
    )

    actual = turing.batch.EnsemblingJob.list(status=[
        turing.batch.EnsemblingJobStatus.PENDING,
        turing.batch.EnsemblingJobStatus.RUNNING
    ])

    assert len(actual) == len(expected)

    for actual, expected in zip(actual, expected):
        assert actual == expected


@responses.activate
@pytest.mark.parametrize(
    "api_response, expected", [
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
                updated_at=utc_date("2021-07-06T13:28:56.252642Z")
            )
        )
    ]
)
def test_submit_job(
        turing_api,
        active_project,
        ensembling_job_config,
        api_response,
        expected,
        use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    responses.add(
        method="POST",
        url=f"/v1/projects/{active_project.id}/jobs",
        body=api_response,
        status=201,
        content_type="application/json"
    )

    actual = turing.batch.job.EnsemblingJob.submit(
        ensembler_id=2,
        config=ensembling_job_config,
    )
    assert actual == expected


@responses.activate
@pytest.mark.parametrize(
    "api_response_get, expected, api_response_refresh, updated", [
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
                updated_at=utc_date("2021-07-06T13:28:56.252642Z")
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
                updated_at=utc_date("2021-07-07T00:00:00.252642Z")
            )
        )
    ]
)
def test_fetch_job(
        turing_api,
        active_project,
        api_response_get,
        expected,
        api_response_refresh,
        updated,
        use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/jobs/{expected.id}",
        body=api_response_get,
        status=200,
        content_type="application/json"
    )

    job = turing.batch.EnsemblingJob.get_by_id(expected.id)

    assert job == expected

    responses.reset()
    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/jobs/{expected.id}",
        body=api_response_refresh,
        status=200,
        content_type="application/json"
    )

    job.refresh()

    assert job == updated


@responses.activate
@pytest.mark.parametrize(
    "job, api_response_delete, api_response_get, expected", [
        pytest.param(
            turing.batch.EnsemblingJob(
                id=1,
                name="ensembling-job",
                ensembler_id=1,
                status=turing.batch.EnsemblingJobStatus.RUNNING,
                project_id=1,
                error="",
                created_at=utc_date("2021-07-06T12:28:32.850365Z"),
                updated_at=utc_date("2021-07-06T13:28:56.252642Z")
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
                updated_at=utc_date("2021-07-07T00:00:00.252642Z")
            )
        )
    ]
)
def test_terminate_job(
        turing_api,
        active_project,
        job,
        api_response_delete,
        api_response_get,
        expected,
        use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    responses.add(
        method="DELETE",
        url=f"/v1/projects/{active_project.id}/jobs/{job.id}",
        body=api_response_delete,
        status=201,
        content_type="application/json"
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/jobs/{job.id}",
        body=api_response_get,
        status=200,
        content_type="application/json"
    )

    assert job != expected

    job.terminate()

    assert job == expected
