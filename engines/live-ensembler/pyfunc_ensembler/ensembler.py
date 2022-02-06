import pandas as pd
from mlflow import pyfunc


class PyFuncEnsembler:
    """
    PyFunc ensembler used for real-time outputs
    """
    FEATURE_COLUMN_PREFIX = '__features__'
    PREDICTION_COLUMN_PREFIX = '__predictions__'
    TREATMENT_CONFIG_COLUMN_PREFIX = '__treatment_config__'

    def __init__(self, artifact_dir: str):
        self.artifact_dir = artifact_dir
        self.ready = False
        self._ensembler = None

    def load(self):
        self._ensembler = pyfunc.load_model(self.artifact_dir)
        self.ready = True

    def predict(self, inputs: dict) -> dict:
        ensembler_inputs = PyFuncEnsembler.preprocess_input(inputs)
        a = self._ensembler.predict(ensembler_inputs)
        return a

    @staticmethod
    def preprocess_input(inputs: dict):
        """
        Dummy preprocessing method
        """
        features = pd.Series(PyFuncEnsembler._get_features_from_inputs(inputs))
        predictions = pd.Series(PyFuncEnsembler._get_predictions_from_inputs(inputs))
        treatment_config = pd.Series(PyFuncEnsembler._get_treatment_config_from_inputs(inputs))
        preprocessed_input = pd.concat([features, predictions, treatment_config]).to_frame().transpose()
        return preprocessed_input

    @staticmethod
    def _get_features_from_inputs(inputs: dict) -> dict:
        raw_predictions = PyFuncEnsembler._flatten_json(inputs['request'])
        features = PyFuncEnsembler._create_dict_with_headers(raw_predictions,
                                                             PyFuncEnsembler.FEATURE_COLUMN_PREFIX)
        return features

    @staticmethod
    def _get_predictions_from_inputs(inputs: dict) -> dict:
        raw_predictions = PyFuncEnsembler._flatten_json(inputs['response']['route_responses'])
        predictions = PyFuncEnsembler._create_dict_with_headers(raw_predictions,
                                                                PyFuncEnsembler.PREDICTION_COLUMN_PREFIX)
        return predictions

    @staticmethod
    def _get_treatment_config_from_inputs(inputs: dict) -> dict:
        raw_predictions = PyFuncEnsembler._flatten_json(inputs['response']['experiment'])

        treatment_config = PyFuncEnsembler._create_dict_with_headers(raw_predictions,
                                                                     PyFuncEnsembler.TREATMENT_CONFIG_COLUMN_PREFIX)
        return treatment_config

    @staticmethod
    def _flatten_json(y):
        out = {}

        def flatten(x, name=''):
            if type(x) is dict:
                for a in x:
                    flatten(x[a], name + a + '.')
            elif type(x) is list:
                i = 0
                for a in x:
                    flatten(a, name + str(i) + '.')
                    i += 1
            else:
                out[name[:-1]] = x

        flatten(y)
        return out

    @staticmethod
    def _create_dict_with_headers(input_dict: dict, header: str):
        new_dict = {}
        for key, value in input_dict.items():
            new_dict[header + key] = value
        return new_dict
