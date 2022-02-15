import pytest

from typing import Any, Dict
from turing.ensembler import PyFunc


class TestEnsembler(PyFunc):

    def initialize(self, artifacts: dict):
        pass

    def ensemble(
            self,
            features: Dict,
            predictions: Dict,
            treatment_config: Dict) -> Any:
        route_name_to_id = TestEnsembler.get_route_name_to_id_mapping(predictions)
        if treatment_config['configuration']['name'] == "choose_the_control":
            return predictions[route_name_to_id['control']]['data']['predictions']
        else:
            return predictions[0]['data']['predictions']

    @staticmethod
    def get_route_name_to_id_mapping(predictions):
        """
        Helper function to look through the predictions returned from the various routes and to map their names to
        their id numbers (the order in which they are found in the payload.
        """
        route_name_to_id = {}
        for i, pred in enumerate(predictions):
            route_name_to_id[pred['route']] = i
        return route_name_to_id


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
