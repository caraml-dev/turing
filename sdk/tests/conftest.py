import pytest


@pytest.fixture
def turing_api() -> str:
    return "http://turing.local.svc:8080"


@pytest.fixture
def use_google_oauth() -> bool:
    return False
