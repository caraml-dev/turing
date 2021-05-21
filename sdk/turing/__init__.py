from typing import Optional
from turing.generated.models import Project
from turing.session import TuringSession
from turing.version import VERSION


active_session: TuringSession = TuringSession(host="http://localhost:8080")


def set_url(url: str, use_google_oauth: bool = True):
    """
    Set Turing URL

    :param url: Turing URL
    :param use_google_oauth: whether use google auth or not
    """

    global active_session
    active_session = TuringSession(host=url, use_google_oauth=use_google_oauth)


def set_project(project_name: str):
    """
    Set active project

    :param project_name: project name
    """
    active_session.set_project(project_name)


def active_project() -> Optional[Project]:
    """
    Get current active project

    :return: active project
    """
    return active_session.active_project


__version__ = VERSION
