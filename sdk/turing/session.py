import os
import mlflow
from typing import List, Optional
from turing.ensembler import EnsemblerType
from turing.generated import ApiClient, Configuration
from turing.generated.apis import EnsemblerApi, EnsemblingJobApi, ProjectApi
from turing.generated.models import \
    Project, \
    Ensembler, \
    EnsemblingJob, \
    EnsemblerJobStatus, \
    EnsemblersPaginatedResults, \
    EnsemblingJobPaginatedResults, \
    IdObject


def require_active_project(f):
    def wrap(*args, **kwargs):
        if not args[0].active_project:
            raise Exception("Active project isn't set, use set_project(...) to set it")
        return f(*args, **kwargs)
    return wrap


class TuringSession:
    """
    Session object for interacting with Turing back-end
    """

    OAUTH_SCOPES = ['https://www.googleapis.com/auth/userinfo.email']

    def __init__(self, host: str, project_name: str = None, use_google_oauth: bool = True):
        """
        Create new session

        :param host: URL of Turing API
        :param project_name: name of the project, this session should stick to
        :param use_google_oauth: should be True if Turing API is protected with Google OAuth
        """
        config = Configuration(host=os.path.join(host, 'v1'))
        self._api_client = ApiClient(config)

        if use_google_oauth:
            import google.auth
            from google.auth.transport.urllib3 import urllib3, AuthorizedHttp

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
        """
        Set this session's active projects
        """
        self.active_project = self.get_project_by_name(project_name)

    def list_projects(self, name: Optional[str] = None) -> List[Project]:
        """
        List all projects, that the current user has access to

        :param name: filter projects by name
        :return: list of projects
        """
        kwargs = {}
        if name:
            kwargs["name"] = name
        return ProjectApi(self._api_client).projects_get(**kwargs)

    def get_project_by_name(self, project_name: str) -> Project:
        """
        Get MLP project by its name

        :param project_name: name of the project
        :raise Exception if the project with given name doesn't exist
        :return: Project details
        """
        p_list = self.list_projects(name=project_name)

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
        """
        List ensemblers
        """
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
        """
        Create a new ensembler in the Turing back-end
        """
        return EnsemblerApi(self._api_client).create_ensembler(
            project_id=self.active_project.id,
            ensembler=ensembler)

    @require_active_project
    def get_ensembler(self, ensembler_id: int) -> Ensembler:
        """
        Fetch ensembler details by its ID
        """
        return EnsemblerApi(self._api_client).get_ensembler_details(
            project_id=self.active_project.id,
            ensembler_id=ensembler_id,
        )

    @require_active_project
    def update_ensembler(self, ensembler: Ensembler) -> Ensembler:
        """
        Update existing ensembler
        """
        return EnsemblerApi(self._api_client).update_ensembler(
            project_id=ensembler.project_id,
            ensembler_id=ensembler.id,
            ensembler=ensembler)

    @require_active_project
    def list_ensembling_jobs(self,
                             status: List[EnsemblerJobStatus] = None,
                             page: Optional[int] = None,
                             page_size: Optional[int] = None) -> EnsemblingJobPaginatedResults:
        """
        List ensembling jobs
        """
        kwargs = {}

        if status:
            kwargs["status"] = status
        if page:
            kwargs["page"] = page
        if page_size:
            kwargs["page_size"] = page_size

        return EnsemblingJobApi(self._api_client).list_ensembling_jobs(
            project_id=self.active_project.id,
            **kwargs
        )

    @require_active_project
    def get_ensembling_job(self, job_id: int) -> EnsemblingJob:
        """
        Fetch ensembling job by its ID
        """
        return EnsemblingJobApi(self._api_client).get_ensembling_job(
            project_id=self.active_project.id,
            job_id=job_id
        )

    @require_active_project
    def terminate_ensembling_job(self, job_id: int) -> IdObject:
        return EnsemblingJobApi(self._api_client).terminate_ensembling_job(
            project_id=self.active_project.id,
            job_id=job_id
        )

    @require_active_project
    def submit_ensembling_job(self, job: EnsemblingJob) -> EnsemblingJob:
        return EnsemblingJobApi(self._api_client).create_ensembling_job(
            project_id=self.active_project.id,
            ensembling_job=job
        )
