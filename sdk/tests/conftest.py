from datetime import datetime, timedelta
import pytest
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
def projects(n=3):
    from turing import generated as client
    return [
        client.models.Project(
            id=i,
            name=f"project_{i}",
            mlflow_tracking_url="http://localhost:5000",
            created_at=datetime.now() + timedelta(seconds=i+10),
            updated_at=datetime.now() + timedelta(seconds=i+10)
        ) for i in range(1, n + 1)]
