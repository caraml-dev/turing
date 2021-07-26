import json
from datetime import datetime, timedelta
import pytest
import random
import tests
from turing import ensembler
import turing.generated.models
import turing.batch.config
import turing.batch.config.source
import turing.batch.config.sink
import uuid
from tests.fixtures.mlflow import mock_mlflow
from tests.fixtures.gcs import mock_gcs


@pytest.fixture
def turing_api() -> str:
    return "http://turing.local.svc:8080"


@pytest.fixture
def use_google_oauth() -> bool:
    return False


@pytest.fixture
def project():
    return turing.generated.models.Project(
        id=10,
        name=f"project_name",
        mlflow_tracking_url="http://localhost:5000",
        created_at=datetime.now(),
        updated_at=datetime.now()
    )


@pytest.fixture
def active_project(responses, project):
    responses.add(
        method="GET",
        url=f"/v1/projects?name={project.name}",
        body=json.dumps([project], default=tests.json_serializer),
        match_querystring=True,
        status=200,
        content_type="application/json"
    )
    return project


@pytest.fixture
def projects(num_projects):
    return [
        turing.generated.models.Project(
            id=i,
            name=f"project_{i}",
            mlflow_tracking_url="http://localhost:5000",
            created_at=datetime.now() + timedelta(seconds=i + 10),
            updated_at=datetime.now() + timedelta(seconds=i + 10)
        ) for i in range(1, num_projects + 1)]


@pytest.fixture
def generic_ensemblers(project, num_ensemblers):
    return [
        turing.generated.models.GenericEnsembler(
            id=i,
            project_id=project.id,
            type=ensembler.PyFuncEnsembler.TYPE,
            name=f"test_ensembler_{i}",
            created_at=datetime.now() + timedelta(seconds=i + 10),
            updated_at=datetime.now() + timedelta(seconds=i + 10)
        ) for i in range(1, num_ensemblers + 1)]


@pytest.fixture
def bucket_name():
    return "bucket-name"


@pytest.fixture
def experiment_name(project, ensembler_name):
    return f"{project.name}/ensemblers/{ensembler_name}"


@pytest.fixture
def experiment_id():
    return random.randint(1, 100)


@pytest.fixture
def run_id():
    return uuid.uuid4().hex


@pytest.fixture
def artifact_uri(bucket_name, experiment_id, run_id):
    return f"gs://{bucket_name}/mlflow/{experiment_id}/{run_id}"


@pytest.fixture
def pyfunc_ensembler(project, ensembler_name, experiment_id, run_id, artifact_uri):
    return turing.generated.models.PyFuncEnsembler(
        id=1,
        project_id=project.id,
        type=ensembler.PyFuncEnsembler.TYPE,
        name=ensembler_name,
        mlflow_experiment_id=experiment_id,
        mlflow_run_id=run_id,
        artifact_uri=artifact_uri,
        created_at=datetime.now(),
        updated_at=datetime.now()
    )


@pytest.fixture
def ensembling_job_config():
    source = turing.batch.config.source.BigQueryDataset(
        table="project.dataset.features",
        features=["feature_1", "feature_2", "features_3"]
    ).join_on(columns=["feature_1"])

    predictions = {
        'model_odd':
            turing.batch.config.source.BigQueryDataset(
                table="project.dataset.scores_model_odd",
                features=["feature_1", "prediction_score"]
            ).join_on(columns=["feature_1"]).select(columns=["prediction_score"]),

        'model_even':
            turing.batch.config.source.BigQueryDataset(
                query="""
                    SELECT feature_1, prediction_score
                    FROM `project.dataset.scores_model_even`
                    WHERE target_date = DATE("2021-03-15", "Asia/Jakarta")
                """,
                options={
                    "viewsEnabled": "true",
                    "materializationDataset": "my_dataset"
                }
            ).join_on(columns=["feature_1"]).select(columns=["prediction_score"])
    }

    result_config = turing.batch.config.ResultConfig(
        type=turing.batch.config.ResultType.INTEGER,
        column_name="prediction_result"
    )

    sink = turing.batch.config.sink.BigQuerySink(
        table="project.dataset.ensembling_results",
        staging_bucket="staging_bucket"
    ).save_mode(turing.batch.config.sink.SaveMode.OVERWRITE) \
        .select(columns=["feature_1", "feature_2", "prediction_result"])

    resource_request = turing.batch.config.ResourceRequest(
        driver_cpu_request="1",
        driver_memory_request="1G",
        executor_replica=5,
        executor_cpu_request="500Mi",
        executor_memory_request="800M"
    )

    return turing.batch.config.EnsemblingJobConfig(
        source=source,
        predictions=predictions,
        result_config=result_config,
        sink=sink,
        service_account="service-account",
        resource_request=resource_request
    )
