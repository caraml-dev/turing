import tornado.web
import orjson
import json

from http import HTTPStatus


class EnsemblerHandler(tornado.web.RequestHandler):
    def initialize(self, ensembler):
        self.ensembler = ensembler
        if not self.ensembler.ready:
            self.ensembler.load()

    def post(self):
        request = EnsemblerHandler.validate_request(self.request)
        response = self.ensembler.predict(request)

        response_json = orjson.dumps(response)
        self.write(response_json)
        self.set_header("Content-Type", "application/json; charset=UTF-8")

    @staticmethod
    def validate_request(request):
        try:
            body = orjson.loads(request.body)
        except json.decoder.JSONDecodeError as e:
            raise tornado.web.HTTPError(
                status_code=HTTPStatus.BAD_REQUEST,
                reason="Unrecognized request format: %s" % e
            )
        return body
