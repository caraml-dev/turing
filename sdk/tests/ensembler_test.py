import random
from typing import Optional, Any

import pandas

import turing.ensembler


class TestEnsembler(turing.ensembler.PyFunc):
    def __init__(self, default: float):
        self._default = default

    def ensemble(
            self,
            features: pandas.Series,
            predictions: pandas.Series,
            treatment_config: Optional[dict]
    ) -> Any:
        if features["treatment"] in predictions:
            return predictions[features["treatment"]]
        else:
            return self._default


def test_predict():
    default_value = random.random()
    ensembler = TestEnsembler(default_value)

    model_input = pandas.DataFrame(data={
        "treatment": ["model_a", "model_b", "unknown"],
        f"{turing.ensembler.PyFunc.PREDICTION_COLUMN_PREFIX}model_a": [0.01, 0.2, None],
        f"{turing.ensembler.PyFunc.PREDICTION_COLUMN_PREFIX}model_b": [0.03, 0.6, 0.4]
    })

    expected = pandas.Series(data=[0.01, 0.6, default_value])
    result = ensembler.predict(context=None, model_input=model_input)

    from pandas._testing import assert_series_equal
    assert_series_equal(expected, result)
