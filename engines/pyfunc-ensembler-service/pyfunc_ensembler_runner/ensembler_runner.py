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
        self._is_legacy_ensembler = False

    def load(self):
        self._ensembler = pyfunc.load_model(self.artifact_dir)
        # If the predict_legacy method is found on the ensembler, it is created from a newer SDK version
        # that supports both versions of the predict API.
        try:
            self._is_legacy_ensembler = not hasattr(self._ensembler._model_impl.python_model, "VERSION")
            print(dir(self._ensembler._model_impl.python_model))
        except:
            pass

    def predict(self, body: Dict[str, Any], headers: Dict[str, str]) -> List[Any]:
        logging.info(f"Input request payload: {body}")
        if self._is_legacy_ensembler:
            output = self._ensembler.predict(body)
        else:
            # Wrap request headers and body into a Dict
            output = self._ensembler.predict({"headers": headers, "body": body})
        logging.info(f"Output response: {output}")
        return output
