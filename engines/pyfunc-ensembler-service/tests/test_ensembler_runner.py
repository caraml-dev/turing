import os
import orjson
import pytest

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
    "ensembler_uri,inputs,headers,expected",
    [
        pytest.param("simple_ensembler_uri", dummy_long_request, {}, [296.15732, 0]),
        pytest.param(
            "simple_ensembler_uri",
            dummy_short_request,
            {"Key": "Value"},
            {"headers": {"Key": "Value"}},
        ),
        pytest.param("legacy_ensembler_uri", dummy_long_request, {}, [296.15732, 0]),
        pytest.param(
            "legacy_ensembler_uri", dummy_short_request, {"Key": "Value"}, [0, 0]
        ),
    ],
)
def test_ensembler_prediction(ensembler_uri, inputs, headers, expected, request):
    ensembler = PyFuncEnsemblerRunner(request.getfixturevalue(ensembler_uri))
    ensembler.load()
    actual = ensembler.predict(orjson.loads(inputs), headers)
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
        response = self.fetch("/ensemble", method="POST", body=dummy_long_request)
        self.assertEqual(response.code, 200)
        self.assertEqual(orjson.loads(response.body), [296.15732, 0])

    def test_invalid_request(self):
        response = self.fetch("/ensemble", method="POST", body=dummy_invalid_request)
        self.assertEqual(response.code, 400)
        self.assertEqual(type(response.error), HTTPError)

    @pytest.fixture(autouse=True)
    def _get_ensembler(self, simple_ensembler_uri):
        self.ensembler = simple_ensembler_uri
