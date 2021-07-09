import json
import os
import pytest
import tests
import turing
import turing.batch
import turing.batch.config
from datetime import datetime
from dateutil.tz import tzutc
from urllib3_mock import Responses

responses = Responses('requests.packages.urllib3')
data_dir = os.path.join(os.path.dirname(__file__), "../testdata/api_responses")

with open(os.path.join(data_dir, "list_jobs_0000.json")) as f:
    list_jobs_0000 = f.read()

with open(os.path.join(data_dir, "submit_job_0000.json")) as f:
    submit_job_0000 = f.read()


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
                    created_at=datetime.strptime(
                        "2021-07-06T12:28:32.850365Z", "%Y-%m-%dT%H:%M:%S.%fZ"
                    ).replace(tzinfo=tzutc()),
                    updated_at=datetime.strptime(
                        "2021-07-06T13:28:56.252642Z", "%Y-%m-%dT%H:%M:%S.%fZ"
                    ).replace(tzinfo=tzutc())
                ),
                turing.batch.EnsemblingJob(
                    id=17,
                    name="my-ensembler: 2021-07-06T23:44:30+03:00",
                    ensembler_id=3,
                    status=turing.batch.EnsemblingJobStatus.FAILED_BUILDING,
                    project_id=1,
                    error="failed building OCI image",
                    created_at=datetime.strptime(
                        "2021-07-06T23:44:30.675673Z", "%Y-%m-%dT%H:%M:%S.%fZ"
                    ).replace(tzinfo=tzutc()),
                    updated_at=datetime.strptime(
                        "2021-07-07T07:36:33.604794Z", "%Y-%m-%dT%H:%M:%S.%fZ"
                    ).replace(tzinfo=tzutc())
                )
            ],
            id="Non empty list"
        )
    ]
)
def test_list_jobs(turing_api, project, use_google_oauth, api_response, expected):
    responses.add(
        method="GET",
        url=f"/v1/projects?name={project.name}",
        body=json.dumps([project], default=tests.json_serializer),
        match_querystring=True,
        status=200,
        content_type="application/json"
    )

    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(project.name)

    responses.add(
        method="GET",
        url=f"/v1/projects/{project.id}/jobs?"
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
        assert actual.id == expected.id
        assert actual.name == expected.name
        assert actual.project_id == expected.project_id
        assert actual.ensembler_id == expected.ensembler_id
        assert actual.status == expected.status
        assert actual.error == expected.error
        assert actual.created_at == expected.created_at
        assert actual.updated_at == expected.updated_at


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
                created_at=datetime.strptime(
                    "2021-07-06T12:28:32.850365Z", "%Y-%m-%dT%H:%M:%S.%fZ"
                ).replace(tzinfo=tzutc()),
                updated_at=datetime.strptime(
                    "2021-07-06T13:28:56.252642Z", "%Y-%m-%dT%H:%M:%S.%fZ"
                ).replace(tzinfo=tzutc())
            )
        )
    ]
)
def test_submit_job(turing_api, project, use_google_oauth, ensembling_job_config, api_response, expected):
    responses.add(
        method="GET",
        url=f"/v1/projects?name={project.name}",
        body=json.dumps([project], default=tests.json_serializer),
        match_querystring=True,
        status=200,
        content_type="application/json"
    )

    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(project.name)

    responses.add(
        method="POST",
        url=f"/v1/projects/{project.id}/jobs",
        body=api_response,
        status=201,
        content_type="application/json"
    )

    actual = turing.batch.job.EnsemblingJob.submit(
        ensembler_id=2,
        config=ensembling_job_config,
    )

    assert actual.id == expected.id
    assert actual.name == expected.name
    assert actual.project_id == expected.project_id
    assert actual.ensembler_id == expected.ensembler_id
    assert actual.status == expected.status
    assert actual.error == expected.error
    assert actual.created_at == expected.created_at
    assert actual.updated_at == expected.updated_at
