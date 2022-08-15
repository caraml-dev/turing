import tornado.web

from pyfunc_ensembler_runner.handler import EnsemblerHandler
from pyfunc_ensembler_runner.ensembler_runner import PyFuncEnsemblerRunner


class PyFuncEnsemblerServer:
    def __init__(self, ensembler: PyFuncEnsemblerRunner):
        self.ensembler = ensembler

    def create_application(self):
        return tornado.web.Application(
            [(r"/ensemble", EnsemblerHandler, dict(ensembler=self.ensembler))]
        )
