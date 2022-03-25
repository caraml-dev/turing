import pytest

from typing import Any, Dict
from turing.ensembler import PyFunc


class TestEnsembler(PyFunc):

    def initialize(self, artifacts: dict):
        pass

    def ensemble(
            self,
            input: Dict,
            predictions: Dict,
            treatment_config: Dict) -> Any:
        if treatment_config['configuration']['name'] == "choose_the_control":
            return predictions['control']['data']['predictions']
        else:
            return [0, 0]


@pytest.fixture
def simple_ensembler_uri():
    import os
    import mlflow
    from mlflow.pyfunc import log_model
    log_model(
        artifact_path='ensembler',
        python_model=TestEnsembler(),
        code_path=[os.path.join(os.path.dirname(__file__), '../pyfunc_ensembler_runner')])

    ensembler_path = os.path.join(mlflow.get_artifact_uri(), 'ensembler')

    return ensembler_path
