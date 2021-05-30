import json
import pytest


@pytest.fixture
def mock_mlflow(responses, project, ensembler, experiment_id, run_id, artifact_uri):
    import urllib.parse
    experiment_name = f"{project.name}/ensemblers/{ensembler.name}"
    responses.add(
        method="GET",
        url=f"/api/2.0/mlflow/experiments/get-by-name?experiment_name={urllib.parse.quote(experiment_name)}",
        body=json.dumps({
            'experiment': {
                'id': experiment_id,
                'name': experiment_name
            }
        }),
        match_querystring=True,
        status=200,
        content_type="application/json"
    )

    responses.add(
        method="POST",
        url="/api/2.0/mlflow/runs/create",
        body=json.dumps({
            'run': {
                'info': {
                    'run_id': run_id,
                    'experiment_id': experiment_id,
                    'status': "RUNNING",
                    'artifact_uri': artifact_uri,
                    'lifecycle_stage': "active"
                },
                'data': {}
            }
        }),
        status=200,
        content_type="application/json"
    )

    responses.add(
        method="GET",
        url=f"/api/2.0/mlflow/runs/get?run_uuid={run_id}&run_id={run_id}",
        match_querystring=True,
        body=json.dumps({
            'run': {
                'info': {
                    'run_id': run_id,
                    'experiment_id': experiment_id,
                    'status': "RUNNING",
                    'artifact_uri': artifact_uri,
                    'lifecycle_stage': "active"
                },
                'data': {}
            }
        }),
        status=200,
        content_type="application/json",
    )

    responses.add(
        method="POST",
        url=f"/api/2.0/mlflow/runs/log-model",
        body="{}",
        status=200,
        content_type="application/json",
    )

    responses.add(
        method="POST",
        url="/api/2.0/mlflow/runs/update",
        body="{}",
        status=200,
        content_type="application/json"
    )
