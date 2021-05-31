import json
import random
from typing import Optional, Any
import pandas
import pytest
import re
import tests
import turing.ensembler
from urllib3_mock import Responses
from turing import generated as client

responses = Responses('requests.packages.urllib3')


@pytest.fixture(scope="module", name="responses")
def _responses():
    return responses


class TestEnsembler(turing.ensembler.PyFunc):
    def __init__(self, default: float):
        self._default = default

    def initialize(self, artifacts: dict):
        pass

    def ensemble(
            self,
            features: pandas.Series,
            predictions: pandas.Series,
            treatment_config: Optional[dict]
    ) -> Any:
        if features["treatment"] in predictions:
            return predictions[features["treatment"]]
        else:
            return self._default


def test_predict():
    default_value = random.random()
    ensembler = TestEnsembler(default_value)

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
@pytest.mark.parametrize(('num_projects', 'num_ensemblers'), [(1, 6)])
def test_list_ensemblers(turing_api, project, generic_ensemblers, use_google_oauth):
    with pytest.raises(Exception, match=re.escape("Active project isn't set, use set_project(...) to set it")):
        turing.PyFuncEnsembler.list()

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

    from turing import generated as client

    page = client.models.EnsemblersPaginatedResults(
        results=generic_ensemblers,
        paging=client.models.PaginatedResultsPaging(total=1, page=1, pages=1)
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{project.id}/ensemblers?type={turing.EnsemblerType.PYFUNC.value}",
        body=json.dumps(page, default=tests.json_serializer),
        match_querystring=True,
        status=200,
        content_type="application/json"
    )

    actual = turing.PyFuncEnsembler.list()
    assert all([isinstance(p, turing.PyFuncEnsembler) for p in actual])

    for actual, expected in zip(actual, generic_ensemblers):
        assert actual.id == expected.id
        assert actual.name == expected.name
        assert actual.project_id == project.id
        assert actual.created_at == expected.created_at
        assert actual.updated_at == expected.updated_at


@responses.activate
@pytest.mark.parametrize('num_projects', [1])
@pytest.mark.parametrize('ensembler_name', ["ensembler_1"])
@pytest.mark.usefixtures("mock_mlflow", "mock_gcs")
def test_create_ensembler(
        turing_api,
        project,
        pyfunc_ensembler,
        use_google_oauth):
    responses.add(
        method="GET",
        url=f"/v1/projects?name={project.name}",
        body=json.dumps([project], default=tests.json_serializer),
        match_querystring=True,
        status=200,
        content_type="application/json"
    )

    responses.add(
        method="POST",
        url=f"/v1/projects/{project.id}/ensemblers",
        body=json.dumps(pyfunc_ensembler, default=tests.json_serializer),
        status=201,
        content_type="application/json"
    )

    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(project.name)

    actual = turing.PyFuncEnsembler.create(
        name=pyfunc_ensembler.name,
        ensembler_instance=TestEnsembler(0.01),
        conda_env={
            'channels': ['defaults'],
            'dependencies': [
                'python=3.7.0'
            ]
        }
    )

    assert actual.id == pyfunc_ensembler.id
    assert actual.name == pyfunc_ensembler.name
    assert actual.project_id == pyfunc_ensembler.project_id
    assert actual.mlflow_experiment_id == pyfunc_ensembler.mlflow_experiment_id
    assert actual.mlflow_run_id == pyfunc_ensembler.mlflow_run_id
    assert actual.artifact_uri == pyfunc_ensembler.artifact_uri
    assert actual.created_at == pyfunc_ensembler.created_at
    assert actual.updated_at == pyfunc_ensembler.updated_at


@responses.activate
@pytest.mark.parametrize(('num_projects', 'num_ensemblers'), [(1, 3)])
@pytest.mark.parametrize('ensembler_name', ["updated_ensembler"])
@pytest.mark.usefixtures("mock_mlflow", "mock_gcs")
def test_update_ensembler(
        turing_api,
        project,
        generic_ensemblers,
        pyfunc_ensembler,
        use_google_oauth):
    responses.add(
        method="GET",
        url=f"/v1/projects?name={project.name}",
        body=json.dumps([project], default=tests.json_serializer),
        match_querystring=True,
        status=200,
        content_type="application/json"
    )

    page = client.models.EnsemblersPaginatedResults(
        results=generic_ensemblers,
        paging=client.models.PaginatedResultsPaging(total=1, page=1, pages=1)
    )

    responses.add(
        method="GET",
        url=f"/v1/projects/{project.id}/ensemblers",
        body=json.dumps(page, default=tests.json_serializer),
        status=201,
        content_type="application/json"
    )

    turing.set_url(turing_api, use_google_oauth)
    turing.set_project(project.name)

    actual, *rest = turing.PyFuncEnsembler.list()

    responses.add(
        method="PUT",
        url=f"/v1/projects/{project.id}/ensemblers/{actual.id}",
        body=json.dumps(pyfunc_ensembler, default=tests.json_serializer),
        status=200,
        content_type="application/json"
    )

    actual.update(
        name=pyfunc_ensembler.name,
        ensembler_instance=TestEnsembler(0.06),
        conda_env={
            'channels': ['defaults'],
            'dependencies': [
                'python>=3.8.0'
            ]
        },
        code_dir=["../samples/quickstart"]
    )
    assert actual.id == pyfunc_ensembler.id
    assert actual.name == pyfunc_ensembler.name
    assert actual.project_id == pyfunc_ensembler.project_id
    assert actual.mlflow_experiment_id == pyfunc_ensembler.mlflow_experiment_id
    assert actual.mlflow_run_id == pyfunc_ensembler.mlflow_run_id
    assert actual.artifact_uri == pyfunc_ensembler.artifact_uri
    assert actual.created_at == pyfunc_ensembler.created_at
    assert actual.updated_at == pyfunc_ensembler.updated_at
