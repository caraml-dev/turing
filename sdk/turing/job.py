from typing import List, Optional
import turing.generated.models
import turing.batch.config
from enum import Enum

from turing._base_types import ApiObject, ApiObjectSpec


class EnsemblingJobStatus(Enum):
    PENDING = "pending"
    RUNNING = "running"
    TERMINATING = "terminating"
    TERMINATED = "terminated"
    COMPLETED = "completed"
    FAILED = "failed"
    FAILED_SUBMISSION = "failed_submission"
    FAILED_BUILDING = "failed_building"


@ApiObjectSpec(turing.generated.models.EnsemblingJob)
class EnsemblingJob(ApiObject):
    _VERSION = "v1"
    _KIND = "BatchEnsemblingJob"

    def __init__(
            self,
            name: str,
            ensembler_id: int,
            status: turing.generated.models.EnsemblerJobStatus = None,
            project_id: int = None,
            **kwargs):
        super(EnsemblingJob, self).__init__(**kwargs)
        self._name = name
        self._project_id = project_id
        self._ensembler_id = ensembler_id
        self._status = status

    @property
    def name(self) -> str:
        return self._name

    @name.setter
    def name(self, name):
        self._name = name

    @property
    def project_id(self) -> int:
        return self._project_id

    @project_id.setter
    def project_id(self, project_id: int):
        self._project_id = project_id

    @property
    def ensembler_id(self) -> int:
        return self._ensembler_id

    @ensembler_id.setter
    def ensembler_id(self, ensembler_id: int):
        self._ensembler_id = ensembler_id

    @property
    def status(self):
        return self._status

    @status.setter
    def status(self, status: turing.generated.models.EnsemblerJobStatus):
        self._status = status

    @classmethod
    def submit(
            cls,
            ensembler: turing.PyFuncEnsembler,
            config: turing.batch.config.EnsemblingJobConfig) -> 'EnsemblingJob':
        from turing.session import active_session

        job_config = turing.generated.models.EnsemblerConfig(
            version=EnsemblingJob._VERSION,
            kind=turing.generated.models.EnsemblerConfigKind(EnsemblingJob._KIND),
            spec=config.job_spec
        )

        job = turing.generated.models.EnsemblingJob(
            ensembler_id=ensembler.id,
            job_config=job_config,

        )

        return EnsemblingJob.from_open_api(
            active_session.submit_ensembling_job(job=job)
        )

    @classmethod
    def list(
            cls,
            status: List[EnsemblingJobStatus] = None,
            page: Optional[int] = None,
            page_size: Optional[int] = None) -> List['EnsemblingJob']:
        from turing.session import active_session

        mapped_statuses = None
        if status:
            mapped_statuses = [turing.generated.models.EnsemblerJobStatus(s.value) for s in status]

        response = active_session.list_ensembling_jobs(
            status=mapped_statuses,
            page=page,
            page_size=page_size
        )
        return [EnsemblingJob.from_open_api(item) for item in response.results]
