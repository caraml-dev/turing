import json
import os.path
import random

import pandas
import pytest
from urllib3_mock import Responses

import tests
import turing.ensembler
import turing.generated.models

responses = Responses("requests.packages.urllib3")
data_dir = os.path.join(os.path.dirname(__file__), "./testdata/api_responses")

with open(os.path.join(data_dir, "list_jobs_0000.json")) as f:
    list_jobs_0000 = f.read()


@pytest.fixture(scope="module", name="responses")
def _responses():
    return responses


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


@responses.activate
@pytest.mark.parametrize("num_ensemblers", [6])
def test_list_ensemblers(
    turing_api, active_project, generic_ensemblers, use_google_oauth
):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    page = turing.generated.models.EnsemblersPaginatedResults(
        results=generic_ensemblers,
        paging=turing.generated.models.PaginationPaging(total=1, page=1, pages=1),
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/ensemblers?type={turing.PyFuncEnsembler.TYPE.value}",
        body=json.dumps(page, default=tests.json_serializer),
        match_querystring=True,
        status=200,
        content_type="application/json",
    )

    actual = turing.PyFuncEnsembler.list()
    assert all([isinstance(p, turing.PyFuncEnsembler) for p in actual])

    for actual, expected in zip(actual, generic_ensemblers):
        assert actual == turing.PyFuncEnsembler.from_open_api(expected)


@responses.activate
@pytest.mark.parametrize("ensembler_name", ["ensembler_1"])
@pytest.mark.usefixtures("mock_mlflow", "mock_gcs")
def test_create_ensembler(
    turing_api, active_project, pyfunc_ensembler, use_google_oauth
):
    responses.add(
        method="POST",
        url=f"/v1/projects/{active_project.id}/ensemblers",
        body=json.dumps(pyfunc_ensembler, default=tests.json_serializer),
        status=201,
        content_type="application/json",
    )

    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    actual = turing.PyFuncEnsembler.create(
        name=pyfunc_ensembler.name,
        ensembler_instance=tests.MyTestEnsembler(0.01),
        conda_env={
            "channels": ["defaults"],
            "dependencies": ["python=3.9.0", {"pip": ["test-lib==0.0.1"]}],
        },
    )

    assert actual == turing.PyFuncEnsembler.from_open_api(pyfunc_ensembler)


@responses.activate
@pytest.mark.parametrize(("num_ensemblers", "ensembler_name"), [(3, "updated")])
@pytest.mark.usefixtures("mock_mlflow", "mock_gcs")
def test_update_ensembler(
    turing_api, active_project, generic_ensemblers, pyfunc_ensembler, use_google_oauth
):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    page = turing.generated.models.EnsemblersPaginatedResults(
        results=generic_ensemblers,
        paging=turing.generated.models.PaginationPaging(total=1, page=1, pages=1),
    )

    emptyJob = turing.generated.models.EnsemblingJobPaginatedResults(
        results=[],
        paging=turing.generated.models.PaginationPaging(total=0, page=1, pages=1),
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/ensemblers",
        body=json.dumps(page, default=tests.json_serializer),
        status=201,
        content_type="application/json",
    )

    actual, *rest = turing.PyFuncEnsembler.list()

    responses.add(
        method="PUT",
        url=f"/v1/projects/{active_project.id}/ensemblers/{actual.id}",
        body=json.dumps(pyfunc_ensembler, default=tests.json_serializer),
        status=200,
        content_type="application/json",
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/router-versions",
        body=json.dumps([], default=tests.json_serializer),
        status=200,
        content_type="application/json",
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/jobs",
        body=json.dumps(emptyJob, default=tests.json_serializer),
        status=200,
        content_type="application/json",
    )

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


@responses.activate
@pytest.mark.parametrize(("num_ensemblers", "ensembler_name"), [(3, "updated")])
def test_update_ensembler_existing_router_version(
    turing_api,
    active_project,
    generic_ensemblers,
    pyfunc_ensembler,
    use_google_oauth,
    generic_router_version,
):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    page = turing.generated.models.EnsemblersPaginatedResults(
        results=generic_ensemblers,
        paging=turing.generated.models.PaginationPaging(total=1, page=1, pages=1),
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/ensemblers",
        body=json.dumps(page, default=tests.json_serializer),
        status=201,
        content_type="application/json",
    )

    actual, *rest = turing.PyFuncEnsembler.list()

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/router-versions",
        body=json.dumps([generic_router_version], default=tests.json_serializer),
        status=200,
        content_type="application/json",
    )

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


@responses.activate
@pytest.mark.parametrize(("num_ensemblers", "ensembler_name"), [(3, "updated")])
def test_update_ensembler_existing_job(
    turing_api, active_project, generic_ensemblers, pyfunc_ensembler, use_google_oauth
):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    page = turing.generated.models.EnsemblersPaginatedResults(
        results=generic_ensemblers,
        paging=turing.generated.models.PaginationPaging(total=1, page=1, pages=1),
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/ensemblers",
        body=json.dumps(page, default=tests.json_serializer),
        status=201,
        content_type="application/json",
    )

    actual, *rest = turing.PyFuncEnsembler.list()

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/router-versions",
        body=json.dumps([], default=tests.json_serializer),
        status=200,
        content_type="application/json",
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/jobs",
        body=list_jobs_0000,
        status=200,
        content_type="application/json",
    )
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


@responses.activate
@pytest.mark.parametrize(
    "actual,expected", [pytest.param(1, turing.generated.models.IdObject(id=1))]
)
@pytest.mark.parametrize("ensembler_name", ["ensembler_1"])
@pytest.mark.usefixtures("mock_mlflow", "mock_gcs")
def test_delete_ensembler(
    turing_api, active_project, use_google_oauth, actual, expected
):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    responses.add(
        method="DELETE",
        url=f"/v1/projects/{active_project.id}/ensemblers/{actual}",
        body=json.dumps(expected, default=tests.json_serializer),
        status=200,
        content_type="application/json",
    )

    response = turing.PyFuncEnsembler.delete(1)

    assert actual == response
