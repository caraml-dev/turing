import tornado.web

from .handler import EnsemblerHandler


class PyFuncEnsemblerServer:
    def __init__(self, ensembler):
        self.ensembler = ensembler

    def create_application(self):
        return tornado.web.Application([
            (r"/ensemble", EnsemblerHandler)
        ])
