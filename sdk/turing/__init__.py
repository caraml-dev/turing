from turing.ensembler import Ensembler, EnsemblerType, PyFuncEnsembler
from turing.project import Project
from turing.router.router import Router
from turing.session import TuringSession
from turing.version import VERSION as __version__

active_session: TuringSession = TuringSession(
    host="http://localhost:8080",
    use_google_oauth=False
)


def set_url(url: str, use_google_oauth: bool = True):
    """
    Set Turing API URL

    :param url: Turing API URL
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


__all__ = ["set_url", "set_project", "active_session"]
