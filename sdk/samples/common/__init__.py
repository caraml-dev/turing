from typing import Any, Optional
import turing
import pandas


class MyEnsembler(turing.ensembler.PyFunc):
    """
    Simple ensembler, that returns predictions from the `model_odd`
    if `customer_id` is odd, or predictions from `model_even` otherwise
    """

    def initialize(self, artifacts: dict):
        pass

    def ensemble(
        self,
        input: pandas.Series,
        predictions: pandas.Series,
        treatment_config: Optional[dict],
        **kwargs: Optional[dict]
    ) -> Any:
        customer_id = input["customer_id"]
        if (customer_id % 2) == 0:
            return predictions["model_even"]
        else:
            return predictions["model_odd"]
