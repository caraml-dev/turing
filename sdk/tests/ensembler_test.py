import json
import random
from typing import Optional, Any
import pandas
import pytest
import re
import tests
import turing.ensembler
from urllib3_mock import Responses

responses = Responses('requests.packages.urllib3')

class TestEnsembler(turing.ensembler.PyFunc):
    def __init__(self, default: float):
        self._default = default

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
def test_list_ensemblers(turing_api, project, use_google_oauth):
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

    expected = [tests.ensembler_1]
    page = client.models.EnsemblersPaginatedResults(
        results=expected,
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

    for actual, expected in zip(actual, expected):
        assert actual.id == expected.id
        assert actual.name == expected.name
        assert actual.project_id == project.id
        assert actual.created_at == expected.created_at
        assert actual.updated_at == expected.updated_at