from typing import Dict
import turing.generated.models

from .source import EnsemblingJobSource, EnsemblingJobPredictionSource
from .sink import EnsemblingJobSink

ResourceRequest = turing.generated.models.EnsemblingResources
ResultConfig = turing.generated.models.EnsemblingJobEnsemblerSpecResult


class EnsemblingJobConfig:
    """
    Configuration of the batch ensembling job
    """

    def __init__(self,
                 source: EnsemblingJobSource,
                 predictions: Dict[str, EnsemblingJobPredictionSource],
                 result_config: ResultConfig,
                 sink: EnsemblingJobSink,
                 service_account: str,
                 resource_request: ResourceRequest = None,
                 env_vars: Dict[str, str] = None):
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
    def source(self) -> 'EnsemblingJobSource':
        return self._source

    @property
    def predictions(self) -> Dict[str, EnsemblingJobPredictionSource]:
        return self._predictions

    @property
    def sink(self) -> 'EnsemblingJobSink':
        return self._sink

    def job_spec(self) -> turing.generated.models.EnsemblingJobSpec:
        return turing.generated.models.EnsemblingJobSpec(
            source=self.source.to_open_api(),
            predictions={name: source.to_open_api() for name, source in self.predictions.items()},
            ensembler=turing.generated.models.EnsemblingJobEnsemblerSpec(
                result=self._result_config
            ),
            sink=self.sink.to_open_api()
        )

    def infra_spec(self) -> turing.generated.models.EnsemblerInfraConfig:
        return turing.generated.models.EnsemblerInfraConfig(
            service_account_name=self._service_account,
            resources=self._resource_request
        )
