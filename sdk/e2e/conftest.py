import os
import turing


def pytest_sessionstart():
    turing.set_url(os.getenv("API_BASE_PATH"), use_google_oauth=False)
    turing.set_project(os.getenv("PROJECT_NAME"))
