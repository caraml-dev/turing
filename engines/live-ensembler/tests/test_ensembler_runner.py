import os
import json
import pytest
import pandas as pd
from pyfunc_ensembler_runner.ensembler_runner import PyFuncEnsemblerRunner


data_dir = os.path.join(os.path.dirname(__file__), "./testdata")
with open(os.path.join(data_dir, "request_short.json")) as f:
    dummy_short_request = json.load(f)


@pytest.mark.parametrize(
    "inputs,expected", [
        pytest.param(
            dummy_short_request,
            pd.DataFrame(
                {
                    'feature_0': {0: 0.12},
                    'feature_1': {0: 'a good feature'},
                    '__predictions__0.route': {0: 'control'},
                    '__predictions__0.data.predictions': {0: 213},
                    '__predictions__0.is_default': {0: True},
                    '__predictions__1.route': {0: 'route_1'},
                    '__predictions__1.data.predictions': {0: 'a bad prediction'},
                    '__predictions__1.is_default': {0: False},
                    '__treatment_config__configuration.name': {0: 'configuration_1'},
                    '__treatment_config__error': {0: ''}
                }
            )
        )
    ])
def test_preprocess_input(inputs, expected):
    actual = PyFuncEnsemblerRunner._preprocess_input(inputs)
    pd.testing.assert_frame_equal(actual, expected, check_dtype=False)
