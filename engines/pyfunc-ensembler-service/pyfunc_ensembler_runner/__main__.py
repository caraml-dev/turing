import argparse
import logging
import traceback
import os

import tornado.ioloop

from pyfunc_ensembler_runner.server import PyFuncEnsemblerServer
from pyfunc_ensembler_runner import PyFuncEnsemblerRunner


parser = argparse.ArgumentParser()
parser.add_argument(
    "--mlflow_ensembler_dir",
    required=True,
    help="A dir pointing to the saved Mlflow Pyfunc ensembler",
)
parser.add_argument(
    "--dry_run",
    default=False,
    action="store_true",
    required=False,
    help="Dry run pyfunc ensembler by loading the specified ensembler "
    "in --mlflow_ensembler_dir without starting webserver",
)

args, _ = parser.parse_known_args()
log_level = os.getenv("LOG_LEVEL", logging.WARNING)

if __name__ == "__main__":
    logging.basicConfig(level=log_level)
    logging.info(
        "Called with arguments:\n%s\n",
        "\n".join([f"{k}: {v}" for k, v in vars(args).items()]),
    )

    ensembler = PyFuncEnsemblerRunner(args.mlflow_ensembler_dir)

    try:
        ensembler.load()
    except Exception as e:
        logging.error(
            "Unable to initialise PyFuncEnsemblerRunner from the MLflow directory provided."
        )
        logging.error(traceback.format_exc())
        exit(1)

    if args.dry_run:
        logging.info("Dry run success")
        exit(0)

    app = PyFuncEnsemblerServer(ensembler).create_application()
    logging.info("Ensembler ready to serve requests!")
    app.listen(8083)
    tornado.ioloop.IOLoop.current().start()
