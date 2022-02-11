import os
import orjson
import pytest
import pandas as pd

from tornado.testing import AsyncHTTPTestCase
from tornado.httpclient import HTTPError
from pyfunc_ensembler_runner.ensembler_runner import PyFuncEnsemblerRunner
from pyfunc_ensembler_runner.server import PyFuncEnsemblerServer


data_dir = os.path.join(os.path.dirname(__file__), "./testdata")
with open(os.path.join(data_dir, "request_short.json")) as f:
    dummy_short_request = f.read()

with open(os.path.join(data_dir, "request_long.json")) as f:
    dummy_long_request = f.read()

with open(os.path.join(data_dir, "request_invalid.json")) as f:
    dummy_invalid_request = f.read()


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
                    '__treatment_config__configuration.name': {0: 'configuration_0'},
                    '__treatment_config__error': {0: ''}
                }
            )
        )
    ])
def test_preprocess_input(inputs, expected):
    actual = PyFuncEnsemblerRunner._preprocess_input(orjson.loads(inputs))
    pd.testing.assert_frame_equal(actual, expected, check_dtype=False)


@pytest.mark.parametrize(
    "inputs,expected", [
        pytest.param(
            dummy_long_request,
            [
                296.15732,
                0
            ]
        )
    ])
def test_ensembler_prediction(simple_ensembler_uri, inputs, expected):
    ensembler = PyFuncEnsemblerRunner(simple_ensembler_uri)
    ensembler.load()
    actual = ensembler.predict(orjson.loads(inputs))
    assert actual == expected


def test_create_ensembler_server(simple_ensembler_uri):
    ensembler = PyFuncEnsemblerRunner(simple_ensembler_uri)
    app = PyFuncEnsemblerServer(ensembler)

    assert app.ensembler == ensembler


# e2e test for real-time ensembler web server and handler
@pytest.mark.usefixtures("simple_ensembler_uri")
class TestEnsemblerService(AsyncHTTPTestCase):
    def get_app(self):
        ensembler = PyFuncEnsemblerRunner(self.ensembler)
        ensembler.load()
        return PyFuncEnsemblerServer(ensembler).create_application()

    def test_valid_request(self):
        response = self.fetch('/ensemble', method="POST", body=dummy_long_request)
        self.assertEqual(response.code, 200)
        self.assertEqual(orjson.loads(response.body), [296.15732, 0])

    def test_invalid_request(self):
        response = self.fetch('/ensemble', method="POST", body=dummy_invalid_request)
        self.assertEqual(response.code, 400)
        self.assertEqual(type(response.error), HTTPError)

    @pytest.fixture(autouse=True)
    def _get_ensembler(self, simple_ensembler_uri):
        self.ensembler = simple_ensembler_uri

