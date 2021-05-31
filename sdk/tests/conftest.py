from datetime import datetime, timedelta
import pytest
import random
from turing import generated as client
from turing import ensembler
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
def project(projects):
    return projects[0]


@pytest.fixture
def projects(num_projects):
    return [
        client.models.Project(
            id=i,
            name=f"project_{i}",
            mlflow_tracking_url="http://localhost:5000",
            created_at=datetime.now() + timedelta(seconds=i+10),
            updated_at=datetime.now() + timedelta(seconds=i+10)
        ) for i in range(1, num_projects + 1)]


@pytest.fixture
def generic_ensemblers(project, num_ensemblers):
    return [
        client.models.GenericEnsembler(
            id=i,
            project_id=project.id,
            type=ensembler.EnsemblerType.PYFUNC.value,
            name=f"test_ensembler_{i}",
            created_at=datetime.now() + timedelta(seconds=i+10),
            updated_at=datetime.now() + timedelta(seconds=i+10)
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
    return client.models.PyFuncEnsembler(
        id=1,
        project_id=project.id,
        type=ensembler.EnsemblerType.PYFUNC.value,
        name=ensembler_name,
        mlflow_experiment_id=experiment_id,
        mlflow_run_id=run_id,
        artifact_uri=artifact_uri,
        created_at=datetime.now(),
        updated_at=datetime.now()
    )
