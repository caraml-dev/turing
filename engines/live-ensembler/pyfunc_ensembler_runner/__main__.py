import argparse
import tornado.ioloop

from pyfunc_ensembler_runner.server import PyFuncEnsemblerServer
from pyfunc_ensembler_runner import PyFuncEnsembler


parser = argparse.ArgumentParser()
parser.add_argument('--mlflow_ensembler_uri', required=True, help='An MLflow URI pointing to the saved ensembler')

args, _ = parser.parse_known_args()


if __name__ == "__main__":
    ensembler = PyFuncEnsembler(args.mlflow_ensembler_uri)

    try:
        ensembler.load()
    except Exception as e:
        print("fml")

    app = PyFuncEnsemblerServer(ensembler).create_application()
    app.listen(8080)
    tornado.ioloop.IOLoop.current().start()
