import logging

from typing import Dict, List, Any
from mlflow import pyfunc


class PyFuncEnsemblerRunner:
    """
    PyFunc ensembler runner used for real-time outputs
    """

    def __init__(self, artifact_dir: str):
        self.artifact_dir = artifact_dir
        self._ensembler = None

    def load(self):
        self._ensembler = pyfunc.load_model(self.artifact_dir)

    def predict(self, body: Dict[str, Any], headers: Dict[str, str]) -> List[Any]:
        logging.info(f"Input request payload: {body}")
        output = self._ensembler.predict({"headers": dict(headers), "body": body})
        logging.info(f"Output response: {output}")
        return output
