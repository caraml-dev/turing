import os
from typing import Optional
from turing.generated.api_client import ApiClient, Configuration
from turing.generated.apis import ProjectApi, EnsemblerApi
from turing.generated.models import Ensembler, Project, EnsemblersPaginatedResults
import google.auth
from google.auth.transport.urllib3 import urllib3, AuthorizedHttp

OAUTH_SCOPES = ['https://www.googleapis.com/auth/userinfo.email']


class TuringClient:
    def __init__(self, turing_url: str, use_google_oauth: bool = True):
        config = Configuration(host=os.path.join(turing_url, 'v1'))
        self._api_client = ApiClient(config)

        if use_google_oauth:
            credentials, project = google.auth.default(scopes=OAUTH_SCOPES)
            authorized_http = AuthorizedHttp(credentials, urllib3.PoolManager())
            self._api_client.rest_client.pool_manager = authorized_http

        self._project_api = ProjectApi(self._api_client)
        self._ensemblers_api = EnsemblerApi(self._api_client)

    def get_project_by_name(self, project_name: str) -> Project:
        p_list = self._project_api.projects_get(name=project_name)

        filtered = [p for p in p_list if p.name == project_name][:1]
        if not filtered:
            raise Exception(
                f"{project_name} does not exist or you don't have access to the project. Please create new "
                f"project using MLP console or ask the project's administrator to be able to access "
                f"existing project.")

        return filtered[0]

    def list_ensemblers(
            self,
            project_id: int,
            page: Optional[int] = None,
            page_size: Optional[int] = None
    ) -> EnsemblersPaginatedResults:
        kwargs = {
            'project_id': project_id
        }
        if page:
            kwargs['page'] = page
        if page_size:
            kwargs['page_size'] = page_size
        return self._ensemblers_api.list_ensemblers(**kwargs)

    def get_ensembler_by_id(self, project_id: int, ensembler_id: int) -> Ensembler:
        return self._ensemblers_api.get_ensembler_details(
            project_id=project_id,
            ensembler_id=ensembler_id
        )

    def create_ensembler(
            self,
            project_id: int,
            ensembler: Ensembler) -> Ensembler:
        return self._ensemblers_api.create_ensembler(
            project_id=project_id,
            ensembler=ensembler
        )
