import json
import os.path
import random
from unittest.mock import patch, MagicMock

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
