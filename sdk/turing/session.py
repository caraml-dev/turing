import os
import mlflow
from typing import Optional
from turing.ensembler import EnsemblerType
from turing.generated.apis import EnsemblerApi, ProjectApi
from turing.generated.models import Project, Ensembler, EnsemblersPaginatedResults


def require_active_project(f):
    def wrap(*args, **kwargs):
        if not args[0].active_project:
            raise Exception("Active project isn't set, use set_project(...) to set it")
        return f(*args, **kwargs)
    return wrap


class TuringSession:
    OAUTH_SCOPES = ['https://www.googleapis.com/auth/userinfo.email']

    def __init__(self, host: str, project_name: str = None, use_google_oauth: bool = True):
        from turing.generated.api_client import ApiClient, Configuration
        import google.auth
        from google.auth.transport.urllib3 import urllib3, AuthorizedHttp

        config = Configuration(host=os.path.join(host, 'v1'))
        self._api_client = ApiClient(config)

        if use_google_oauth:
            credentials, project = google.auth.default(scopes=TuringSession.OAUTH_SCOPES)
            authorized_http = AuthorizedHttp(credentials, urllib3.PoolManager())
            self._api_client.rest_client.pool_manager = authorized_http

        self._project = None

        if project_name:
            self.set_project(project_name)

    @property
    def active_project(self) -> Optional[Project]:
        return self._project

    @active_project.setter
    def active_project(self, project):
        mlflow.tracking.set_tracking_uri(project.mlflow_tracking_url)
        self._project = project

    def set_project(self, project_name: str):
        self.active_project = self.get_project_by_name(project_name)

    def get_project_by_name(self, project_name: str) -> Project:
        projects_api = ProjectApi(self._api_client)
        p_list = projects_api.projects_get(name=project_name)

        filtered = [p for p in p_list if p.name == project_name][:1]
        if not filtered:
            raise Exception(
                f"{project_name} does not exist or you don't have access to the project. Please create new "
                f"project using MLP console or ask the project's administrator to be able to access "
                f"existing project.")

        return filtered[0]

    @require_active_project
    def list_ensemblers(
            self,
            ensembler_type: Optional[EnsemblerType] = None,
            page: Optional[int] = None,
            page_size: Optional[int] = None) -> EnsemblersPaginatedResults:
        kwargs = {}

        if ensembler_type:
            kwargs["type"] = ensembler_type.value
        if page:
            kwargs["page"] = page
        if page_size:
            kwargs["page_size"] = page_size

        return EnsemblerApi(self._api_client).list_ensemblers(
            project_id=self.active_project.id,
            **kwargs
        )

    @require_active_project
    def create_ensembler(self, ensembler: Ensembler) -> Ensembler:
        return EnsemblerApi(self._api_client).create_ensembler(
            project_id=self.active_project.id,
            ensembler=ensembler)

    @require_active_project
    def update_ensembler(self, ensembler: Ensembler) -> Ensembler:
        return EnsemblerApi(self._api_client).update_ensembler(
            project_id=ensembler.project_id,
            ensembler_id=ensembler.id,
            ensembler=ensembler)
