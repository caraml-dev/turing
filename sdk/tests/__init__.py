from datetime import date, datetime
from turing import generated as client


def json_serializer(o):
    if isinstance(o, (date, datetime)):
        return o.isoformat()
    if isinstance(o, (client.model_utils.ModelNormal, client.model_utils.ModelComposed)):
        return o.to_dict()


project_1 = client.models.Project(
    id=1,
    name="project_1",
    mlflow_tracking_url="http://localhost:5000",
    created_at=datetime(2021, 5, 25, 0, 0, 0),
    updated_at=datetime(2021, 5, 25, 0, 0, 5)
)
project_2 = client.models.Project(
    id=2,
    name="project_2",
    mlflow_tracking_url="http://localhost:5000",
    created_at=datetime(2021, 5, 25, 2, 0, 0),
    updated_at=datetime(2021, 5, 25, 2, 0, 5)
)

ensembler_1 = client.models.GenericEnsembler(
    id=1,
    project_id=project_1.id,
    type="pyfunc",
    name="test_ensembler_1",
    created_at=datetime(2021, 5, 25, 0, 0, 0),
    updated_at=datetime(2021, 5, 25, 0, 0, 0)
)
