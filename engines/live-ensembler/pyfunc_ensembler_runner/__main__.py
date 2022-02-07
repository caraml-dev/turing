import argparse
import logging
import traceback

import tornado.ioloop

from pyfunc_ensembler_runner.server import PyFuncEnsemblerServer
from pyfunc_ensembler_runner import PyFuncEnsemblerRunner


parser = argparse.ArgumentParser()
parser.add_argument('--mlflow_ensembler_uri', required=True, help='An MLflow URI pointing to the saved ensembler')

args, _ = parser.parse_known_args()


if __name__ == "__main__":
    ensembler = PyFuncEnsemblerRunner(args.mlflow_ensembler_uri)

    try:
        ensembler.load()
    except Exception as e:
        logging.error("Unable to initialise PyFuncEnsemblerRunner from the MLflow URI provided.")
        logging.error(traceback.format_exc())
        exit(1)

    app = PyFuncEnsemblerServer(ensembler).create_application()
    app.listen(8080)
    tornado.ioloop.IOLoop.current().start()
