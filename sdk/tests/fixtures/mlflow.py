import json
import pytest
from urllib.parse import quote_plus


@pytest.fixture
def mock_mlflow(responses, experiment_name, experiment_id, run_id, artifact_uri):
    responses.add(
        method="GET",
        url=f"/api/2.0/mlflow/experiments/get-by-name?experiment_name={quote_plus(experiment_name)}",
        body=json.dumps(
            {
                "experiment": {
                    "id": experiment_id,
                    "name": experiment_name,
                    "lifecycle_stage": "active",
                }
            }
        ),
        match_querystring=True,
        status=200,
        content_type="application/json",
    )

    responses.add(
        method="POST",
        url="/api/2.0/mlflow/runs/create",
        body=json.dumps(
            {
                "run": {
                    "info": {
                        "run_id": run_id,
                        "experiment_id": experiment_id,
                        "status": "RUNNING",
                        "artifact_uri": artifact_uri,
                        "lifecycle_stage": "active",
                    },
                    "data": {},
                }
            }
        ),
        status=200,
        content_type="application/json",
    )

    responses.add(
        method="GET",
        url=f"/api/2.0/mlflow/runs/get?run_uuid={run_id}&run_id={run_id}",
        match_querystring=True,
        body=json.dumps(
            {
                "run": {
                    "info": {
                        "run_id": run_id,
                        "experiment_id": experiment_id,
                        "status": "RUNNING",
                        "artifact_uri": artifact_uri,
                        "lifecycle_stage": "active",
                    },
                    "data": {},
                }
            }
        ),
        status=200,
        content_type="application/json",
    )

    responses.add(
        method="POST",
        url="/api/2.0/mlflow/runs/log-model",
        body="{}",
        status=200,
        content_type="application/json",
    )

    responses.add(
        method="POST",
        url="/api/2.0/mlflow/runs/update",
        body="{}",
        status=200,
        content_type="application/json",
    )
