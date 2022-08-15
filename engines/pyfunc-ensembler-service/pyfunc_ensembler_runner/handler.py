import tornado.web
import orjson
import json

from http import HTTPStatus
from typing import Any, Dict
from pyfunc_ensembler_runner.ensembler_runner import PyFuncEnsemblerRunner


class EnsemblerHandler(tornado.web.RequestHandler):
    def initialize(self, ensembler: PyFuncEnsemblerRunner):
        self.ensembler = ensembler

    def post(self):
        body = EnsemblerHandler.validate_request(self.request)
        response = self.ensembler.predict(body, self.request.headers)
        response_json = orjson.dumps(response)
        self.write(response_json)
        self.set_header("Content-Type", "application/json; charset=UTF-8")

    @staticmethod
    def validate_request(request: Any) -> Dict[str, Any]:
        try:
            body = orjson.loads(request.body)
        except json.decoder.JSONDecodeError as e:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason="Unrecognized request format: %s" % e,
            )
        return body
