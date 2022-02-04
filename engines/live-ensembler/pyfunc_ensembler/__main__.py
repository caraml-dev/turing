import argparse

from server import PyFuncEnsemblerServer
from pyfunc_ensembler import PyFuncEnsembler


parser = argparse.ArgumentParser()
parser.add_argument('--mlflow_ensembler_uri', required=True, help='An MLflow URI pointing to the saved ensembler')

args, _ = parser.parse_known_args()


if __name__ == "__main__":
    ensembler = PyFuncEnsembler(args.mlflow_ensembler_uri)

    try:
        ensembler.load()
    except Exception as e:
        print("fml")

    PyFuncEnsemblerServer(ensembler)
