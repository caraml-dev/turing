from enum import Enum
from typing import Dict, Optional
import turing.generated.models
from turing.generated.model_utils import OpenApiModel
from turing.generated.model.env_var import EnvVar
from .source import EnsemblingJobSource, EnsemblingJobPredictionSource
from .sink import EnsemblingJobSink

ResourceRequest = turing.generated.models.EnsemblingResources


class ResultType(Enum):
    DOUBLE = 0
    FLOAT = 1
    INTEGER = 2
    LONG = 3
    STRING = 4
    ARRAY = 10

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.EnsemblingJobResultType(self.name)


class ResultConfig:
    def __init__(
        self, type: ResultType, column_name: str, item_type: Optional[ResultType] = None
    ):
        self._type, self._column_name, self._item_type = type, column_name, item_type

    def to_open_api(self) -> OpenApiModel:
        kwargs = {"type": self._type.to_open_api(), "column_name": self._column_name}
        if self._item_type:
            kwargs["item_type"] = self._item_type.to_open_api()

        return turing.generated.models.EnsemblingJobEnsemblerSpecResult(**kwargs)


class EnsemblingJobConfig:
    """
    Configuration of the batch ensembling job
    """

    def __init__(
        self,
        source: EnsemblingJobSource,
        predictions: Dict[str, EnsemblingJobPredictionSource],
        result_config: ResultConfig,
        sink: EnsemblingJobSink,
        service_account: str,
        resource_request: ResourceRequest = None,
        env_vars: Dict[str, str] = {},
    ):
        """
        Create new instance of batch ensembling job configuration

        :param source: source configuration
        :param predictions: dictionary with configuration of model predictions
        :param result_config: configuration of ensembling results
        :param sink: sink configuration
        :param service_account:  secret name containing the service account for executing the ensembling job
        :param resource_request: optional resource request for starting the ensembling job.
            If not given the system default will be used.
        :param env_vars: optional environment variables in the form of a key value pair in a list.
        """
        self._source = source
        self._predictions = predictions
        self._result_config = result_config
        self._sink = sink
        self._service_account = service_account
        self._resource_request = resource_request
        self._env_vars = env_vars

    @property
    def source(self) -> "EnsemblingJobSource":
        return self._source

    @property
    def predictions(self) -> Dict[str, EnsemblingJobPredictionSource]:
        return self._predictions

    @property
    def sink(self) -> "EnsemblingJobSink":
        return self._sink

    @property
    def result_config(self) -> "ResultConfig":
        return self._result_config

    @property
    def service_account(self) -> str:
        return self._service_account

    @property
    def env_vars(self) -> Dict[str, str]:
        return self._env_vars

    @property
    def resource_request(self) -> Optional["ResourceRequest"]:
        return self._resource_request

    def job_spec(self) -> turing.generated.models.EnsemblingJobSpec:
        return turing.generated.models.EnsemblingJobSpec(
            source=self.source.to_open_api(),
            predictions={
                name: source.to_open_api() for name, source in self.predictions.items()
            },
            ensembler=turing.generated.models.EnsemblingJobEnsemblerSpec(
                result=self.result_config.to_open_api()
            ),
            sink=self.sink.to_open_api(),
        )

    def infra_spec(self) -> turing.generated.models.EnsemblerInfraConfig:
        if self.env_vars is None:
            env_vars = []
        else:
            env_vars = [
                EnvVar(name=name, value=value) for name, value in self.env_vars.items()
            ]
        return turing.generated.models.EnsemblerInfraConfig(
            service_account_name=self.service_account,
            resources=self.resource_request,
            env=env_vars,
        )
