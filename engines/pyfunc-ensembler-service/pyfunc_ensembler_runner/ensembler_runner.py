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

    def predict(self, inputs: Dict[str, Any]) -> List[Any]:
        logging.info(f"Input request payload: {inputs}")
        output = self._ensembler.predict(inputs)
        logging.info(f"Output response: {output}")
        return output
