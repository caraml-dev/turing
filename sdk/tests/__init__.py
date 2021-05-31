from datetime import date, datetime
from turing import generated as client


def json_serializer(o):
    if isinstance(o, (date, datetime)):
        return o.isoformat()
    if isinstance(o, (client.model_utils.ModelNormal, client.model_utils.ModelComposed)):
        return o.to_dict()
