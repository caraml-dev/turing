from datetime import date, datetime
from turing import generated as client
import turing.ensembler


def json_serializer(o):
    if isinstance(o, (date, datetime)):
        return o.isoformat()
    if isinstance(o, (client.model_utils.ModelNormal, client.model_utils.ModelComposed)):
        return o.to_dict()


class MyTestEnsembler(turing.ensembler.PyFunc):
    import pandas
    from typing import Any, Optional

    def __init__(self, default: float):
        self._default = default

    def initialize(self, artifacts: dict):
        pass

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
