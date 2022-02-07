import pandas as pd

from turing.ensembler import PyFunc
from mlflow import pyfunc


class PyFuncEnsemblerRunner:
    """
    PyFunc ensembler runner used for real-time outputs
    """

    def __init__(self, artifact_dir: str):
        self.artifact_dir = artifact_dir
        self.ready = False
        self._ensembler = None

    def load(self):
        self._ensembler = pyfunc.load_model(self.artifact_dir)
        self.ready = True

    def predict(self, inputs: dict) -> dict:
        ensembler_inputs = PyFuncEnsemblerRunner.preprocess_input(inputs)
        output = self._ensembler.predict(ensembler_inputs).iloc[0].to_list()
        return output

    @staticmethod
    def preprocess_input(inputs: dict):
        features = pd.Series(PyFuncEnsemblerRunner._get_features_from_inputs(inputs))
        predictions = pd.Series(PyFuncEnsemblerRunner._get_predictions_from_inputs(inputs))
        treatment_config = pd.Series(PyFuncEnsemblerRunner._get_treatment_config_from_inputs(inputs))
        preprocessed_input = pd.concat([features, predictions, treatment_config]).to_frame().transpose()
        return preprocessed_input

    @staticmethod
    def _get_features_from_inputs(inputs: dict) -> dict:
        features = PyFuncEnsemblerRunner._flatten_json(inputs['request'])
        return features

    @staticmethod
    def _get_predictions_from_inputs(inputs: dict) -> dict:
        raw_predictions = PyFuncEnsemblerRunner._flatten_json(inputs['response']['route_responses'])
        predictions = PyFuncEnsemblerRunner._create_dict_with_headers(raw_predictions,
                                                                      PyFunc.PREDICTION_COLUMN_PREFIX)
        return predictions

    @staticmethod
    def _get_treatment_config_from_inputs(inputs: dict) -> dict:
        raw_predictions = PyFuncEnsemblerRunner._flatten_json(inputs['response']['experiment'])

        treatment_config = PyFuncEnsemblerRunner._create_dict_with_headers(raw_predictions,
                                                                           PyFunc.TREATMENT_CONFIG_COLUMN_PREFIX)
        return treatment_config

    @staticmethod
    def _flatten_json(y):
        """
        Helper function to normalise a nested dictionary into a dictionary of depth 1 with names following the
        convention: key_1.key_2.key_3..., with a period acting as a delimiter between nested keys

        Items in lists have their names rendered using their index numbers within the lists they are found in.
        """
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
