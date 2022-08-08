import pytest

from typing import Any, Dict
from turing.ensembler import PyFunc


class TestEnsembler(PyFunc):
    def initialize(self, artifacts: Dict):
        pass

    def ensemble(
        self, input: Dict, predictions: Dict, treatment_config: Dict, **kwargs
    ) -> Any:
        if treatment_config["configuration"]["name"] == "choose_the_control":
            return predictions["control"]["data"]["predictions"]
        else:
            return kwargs


class LegacyEnsembler(PyFunc):
    def initialize(self, artifacts: Dict):
        pass

    def ensemble(self, input: Dict, predictions: Dict, treatment_config: Dict) -> Any:
        if treatment_config["configuration"]["name"] == "choose_the_control":
            return predictions["control"]["data"]["predictions"]
        else:
            return [0, 0]


def get_ensembler_path(model):
    import os
    import mlflow
    from mlflow.pyfunc import log_model

    log_model(
        artifact_path="ensembler",
        python_model=model,
        code_path=[
            os.path.join(os.path.dirname(__file__), "../pyfunc_ensembler_runner")
        ],
    )

    ensembler_path = os.path.join(mlflow.get_artifact_uri(), "ensembler")

    return ensembler_path


@pytest.fixture
def simple_ensembler_uri():
    return get_ensembler_path(TestEnsembler())


@pytest.fixture
def legacy_ensembler_uri():
    return get_ensembler_path(LegacyEnsembler())
