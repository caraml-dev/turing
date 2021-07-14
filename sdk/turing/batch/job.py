from enum import Enum
from typing import List, Optional
import turing.generated.models
from .config import EnsemblingJobConfig
from turing._base_types import ApiObject, ApiObjectSpec


class EnsemblingJobStatus(Enum):
    """
    Status of ensembling job

    Possible statuses:
    JobPending --▶ JobFailedSubmission
        |
        |
        |
        |
    JobBuildingImage --▶ JobFailedBuildImage
        |
        |
        |
        |
        ▼
    JobRunning --▶ JobFailed
        |
        |
        |--▶ JobTerminating --▶ JobTerminated
        |
        |
        ▼
    JobCompleted
    """

    PENDING = "pending"
    BUILDING = "building"
    RUNNING = "running"
    TERMINATING = "terminating"
    TERMINATED = "terminated"
    COMPLETED = "completed"
    FAILED = "failed"
    FAILED_SUBMISSION = "failed_submission"
    FAILED_BUILDING = "failed_building"


@ApiObjectSpec(turing.generated.models.EnsemblingJob)
class EnsemblingJob(ApiObject):
    """
    API entity that represents ensembling batch job
    """

    _VERSION = "v1"
    _KIND = "BatchEnsemblingJob"

    def __init__(
            self,
            name: str,
            ensembler_id: int,
            status: EnsemblingJobStatus = None,
            project_id: int = None,
            error: str = None,
            **kwargs):

        super(EnsemblingJob, self).__init__(**kwargs)
        self._name = name
        self._project_id = project_id
        self._ensembler_id = ensembler_id
        self._status = EnsemblingJobStatus(status)
        self._error = error

    @property
    def name(self) -> str:
        return self._name

    @property
    def project_id(self) -> int:
        return self._project_id

    @property
    def ensembler_id(self) -> int:
        return self._ensembler_id

    @property
    def status(self) -> 'EnsemblingJobStatus':
        return self._status

    @property
    def error(self) -> str:
        return self._error

    def refresh(self):
        """
        Fetches latest updates of this ensembling job
        """
        self.__dict__.update(
            EnsemblingJob.from_open_api(
                turing.active_session.get_ensembling_job(job_id=self.id)
            ).__dict__
        )

    def terminate(self):
        """
        Terminates this ensembling job
        """
        turing.active_session.terminate_ensembling_job(job_id=self.id)
        self.refresh()

    @classmethod
    def get_by_id(cls, job_id: int) -> 'EnsemblingJob':
        """
        Fetch ensembling job by its ID
        """
        return EnsemblingJob.from_open_api(
            turing.active_session.get_ensembling_job(job_id=job_id))

    @classmethod
    def submit(
            cls,
            ensembler_id: int,
            config: EnsemblingJobConfig) -> 'EnsemblingJob':
        """
        Submit ensembling job with a given configuration for execution

        :param ensembler_id: Id of the ensembler, that should be used for ensembling
        :param config: configuration of ensembling job
        :return: instance of ensembling job
        """
        job_config = turing.generated.models.EnsemblerConfig(
            version=EnsemblingJob._VERSION,
            kind=turing.generated.models.EnsemblerConfigKind(EnsemblingJob._KIND),
            spec=config.job_spec()
        )

        job = turing.generated.models.EnsemblingJob(
            ensembler_id=ensembler_id,
            infra_config=config.infra_spec(),
            job_config=job_config,
        )

        return EnsemblingJob.from_open_api(
            turing.active_session.submit_ensembling_job(job=job))

    @classmethod
    def list(
            cls,
            status: List[EnsemblingJobStatus] = None,
            page: Optional[int] = None,
            page_size: Optional[int] = None) -> List['EnsemblingJob']:
        """
        List ensembling jobs in the active project

        :param status: (optional) filter jobs by one or more statuses
        :param page:  (optional) pagination parameters – page number
        :param page_size: (optional) pagination parameters - page size

        :return: list of ensembling jobs
        """
        mapped_statuses = None
        if status:
            mapped_statuses = [turing.generated.models.EnsemblerJobStatus(s.value) for s in status]

        response = turing.active_session.list_ensembling_jobs(
            status=mapped_statuses,
            page=page,
            page_size=page_size
        )
        return [EnsemblingJob.from_open_api(item) for item in response.results]
