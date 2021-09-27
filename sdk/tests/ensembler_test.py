import json
import os.path
import random
import pandas
import pytest
import re
import tests
import turing.ensembler
from urllib3_mock import Responses
import turing.generated.models

responses = Responses('requests.packages.urllib3')


@pytest.fixture(scope="module", name="responses")
def _responses():
    return responses


def test_predict():
    default_value = random.random()
    ensembler = tests.MyTestEnsembler(default_value)

    model_input = pandas.DataFrame(data={
        "treatment": ["model_a", "model_b", "unknown"],
        f"{turing.ensembler.PyFunc.PREDICTION_COLUMN_PREFIX}model_a": [0.01, 0.2, None],
        f"{turing.ensembler.PyFunc.PREDICTION_COLUMN_PREFIX}model_b": [0.03, 0.6, 0.4]
    })

    expected = pandas.Series(data=[0.01, 0.6, default_value])
    result = ensembler.predict(context=None, model_input=model_input)

    from pandas._testing import assert_series_equal
    assert_series_equal(expected, result)


@responses.activate
@pytest.mark.parametrize("num_ensemblers", [6])
def test_list_ensemblers(turing_api, active_project, generic_ensemblers, use_google_oauth):
    with pytest.raises(Exception, match=re.escape("Active project isn't set, use set_project(...) to set it")):
        turing.PyFuncEnsembler.list()

    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    page = turing.generated.models.EnsemblersPaginatedResults(
        results=generic_ensemblers,
        paging=turing.generated.models.PaginationPaging(total=1, page=1, pages=1)
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/ensemblers?type={turing.PyFuncEnsembler.TYPE.value}",
        body=json.dumps(page, default=tests.json_serializer),
        match_querystring=True,
        status=200,
        content_type="application/json"
    )

    actual = turing.PyFuncEnsembler.list()
    assert all([isinstance(p, turing.PyFuncEnsembler) for p in actual])

    for actual, expected in zip(actual, generic_ensemblers):
        assert actual == turing.PyFuncEnsembler.from_open_api(expected)


@responses.activate
@pytest.mark.parametrize("ensembler_name", ['ensembler_1'])
@pytest.mark.usefixtures("mock_mlflow", "mock_gcs")
def test_create_ensembler(
        turing_api,
        active_project,
        pyfunc_ensembler,
        use_google_oauth):
    responses.add(
        method="POST",
        url=f"/v1/projects/{active_project.id}/ensemblers",
        body=json.dumps(pyfunc_ensembler, default=tests.json_serializer),
        status=201,
        content_type="application/json"
    )

    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    actual = turing.PyFuncEnsembler.create(
        name=pyfunc_ensembler.name,
        ensembler_instance=tests.MyTestEnsembler(0.01),
        conda_env={
            'channels': ['defaults'],
            'dependencies': [
                'python=3.7.0'
            ]
        }
    )

    assert actual == turing.PyFuncEnsembler.from_open_api(pyfunc_ensembler)


@responses.activate
@pytest.mark.parametrize(
    ('num_ensemblers', 'ensembler_name'),
    [(3, "updated")]
)
@pytest.mark.usefixtures("mock_mlflow", "mock_gcs")
def test_update_ensembler(
        turing_api,
        active_project,
        generic_ensemblers,
        pyfunc_ensembler,
        use_google_oauth):
    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(active_project.name)

    page = turing.generated.models.EnsemblersPaginatedResults(
        results=generic_ensemblers,
        paging=turing.generated.models.PaginationPaging(total=1, page=1, pages=1)
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{active_project.id}/ensemblers",
        body=json.dumps(page, default=tests.json_serializer),
        status=201,
        content_type="application/json"
    )

    actual, *rest = turing.PyFuncEnsembler.list()

    responses.add(
        method="PUT",
        url=f"/v1/projects/{active_project.id}/ensemblers/{actual.id}",
        body=json.dumps(pyfunc_ensembler, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    actual.update(
        name=pyfunc_ensembler.name,
        ensembler_instance=tests.MyTestEnsembler(0.06),
        conda_env={
            'channels': ['defaults'],
            'dependencies': [
                'python>=3.8.0'
            ]
        },
        code_dir=[os.path.join(os.path.dirname(os.path.realpath(__file__)), "..", "samples/quickstart")]
    )
    assert actual == turing.PyFuncEnsembler.from_open_api(pyfunc_ensembler)
