from mlflow import pyfunc


class PyFuncEnsembler:
    """
    PyFunc ensembler used for real-time outputs
    """

    def __init__(self, name: str, artifact_dir: str):
        self.name = name
        self.artifact_dir = artifact_dir
        self.ready = False
        self._ensembler = None

    def load(self):
        self._ensembler = pyfunc.load_model(self.artifact_dir)
        self.ready = True

    def predict(self, inputs: dict) -> dict:
        ensembler_inputs = PyFuncEnsembler.preprocess_input(inputs)
        return self._ensembler.predict(ensembler_inputs)

    @staticmethod
    def preprocess_input(inputs):
        return inputs
