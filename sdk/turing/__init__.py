from typing import Optional

import mlflow.tracking

from turing.generated.models import Project, Ensembler, PyFuncEnsembler, EnsemblersPaginatedResults
from turing.client import TuringClient
from turing.session import TuringSession

_turing_client: Optional[TuringClient] = None
_active_project: Optional[Project]
active_session: Optional[TuringSession] = None


def set_url(url: str, use_google_oauth: bool = True):
    """
    Set Turing URL

    :param url: Turing URL
    :param use_google_oauth: whether use google auth or not
    """
    global _turing_client
    _turing_client = TuringClient(url, use_google_oauth)

    global active_session
    active_session = TuringSession(host=url, use_google_oauth=use_google_oauth)


def require_client(f):
    def wrap(*args, **kwargs):
        if not _turing_client:
            raise Exception("URL is not set, use set_url(...) to set it")
        return f(*args, **kwargs)

    return wrap


def require_active_project(f):
    def wrap(*args, **kwargs):
        if not _active_project:
            raise Exception("Active project isn't set, use set_project(...) to set it")
        return f(*args, **kwargs)

    return wrap


@require_client
def set_project(project_name: str):
    """
    Set active project

    :param project_name: project name
    """

    p = _turing_client.get_project_by_name(project_name)
    global _active_project
    _active_project = p
    mlflow.tracking.set_tracking_uri(_active_project.mlflow_tracking_url)

    global active_session
    active_session.set_project(project_name)


@require_client
@require_active_project
def active_project() -> Optional[Project]:
    """
    Get current active project

    :return: active project
    """
    return _active_project


@require_client
@require_active_project
def list_ensemblers(
        page: Optional[int] = None,
        page_size: Optional[int] = None) -> EnsemblersPaginatedResults:
    return _turing_client.list_ensemblers(_active_project.id, page, page_size)


@require_client
@require_active_project
def get_ensembler(ensembler_id: int) -> Ensembler:
    return _turing_client.get_ensembler_by_id(_active_project.id, ensembler_id)
