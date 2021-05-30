from datetime import date, datetime
from turing import generated as client


def json_serializer(o):
    if isinstance(o, (date, datetime)):
        return o.isoformat()
    if isinstance(o, (client.model_utils.ModelNormal, client.model_utils.ModelComposed)):
        return o.to_dict()


ensembler_1 = client.models.GenericEnsembler(
    id=1,
    project_id=1,
    type="pyfunc",
    name="test_ensembler_1",
    created_at=datetime(2021, 5, 25, 0, 0, 0),
    updated_at=datetime(2021, 5, 25, 0, 0, 0)
)
