from typing import Optional, List, Any, Dict

from turing.generated.models import Project, EnsemblersPaginatedResults
from turing.client import TuringClient

_turing_client: Optional[TuringClient] = None
_active_project: Optional[Project]


def set_url(url: str, use_google_oauth: bool = True):
    """
    Set Turing URL

    :param url: Turing URL
    :param use_google_oauth: whether use google auth or not
    """
    global _turing_client
    _turing_client = TuringClient(url, use_google_oauth)


def _require_client(f):
    def wrap(*args, **kwargs):
        if not _turing_client:
            raise Exception("URL is not set, use set_url(...) to set it")
        return f(*args, **kwargs)

    return wrap


def __require_active_project(f):
    def wrap(*args, **kwargs):
        if not _active_project:
            raise Exception("Active project isn't set, use set_project(...) to set it")
        return f(*args, **kwargs)

    return wrap


@_require_client
def set_project(project_name: str):
    """
    Set active project

    :param project_name: project name
    """

    p = _turing_client.get_project(project_name)
    global _active_project
    _active_project = p


@_require_client
@__require_active_project
def active_project() -> Optional[Project]:
    """
    Get current active project

    :return: active project
    """
    return _active_project


@_require_client
@__require_active_project
def list_ensemblers(
        page: Optional[int] = None,
        page_size: Optional[int] = None) -> EnsemblersPaginatedResults:
    return _turing_client.list_ensemblers(_active_project.id, page, page_size)


@_require_client
@__require_active_project
def log_new_ensembler():
    return _turing_client.create_ensembler(_active_project)
