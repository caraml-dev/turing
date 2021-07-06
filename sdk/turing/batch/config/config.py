from typing import Dict
import turing.generated.models

from .source import EnsemblingJobSource, EnsemblingJobPredictionSource
from .sink import EnsemblingJobSink

ResourceRequest = turing.generated.models.EnsemblingResources
ResultConfig = turing.generated.models.EnsemblingJobEnsemblerSpecResult


class EnsemblingJobConfig:

    def __init__(self,
                 source: EnsemblingJobSource,
                 predictions: Dict[str, EnsemblingJobPredictionSource],
                 result_config: ResultConfig,
                 sink: EnsemblingJobSink,
                 service_account: str,
                 resource_request: ResourceRequest = None,
                 env_vars: Dict[str, str] = None):
        self._source = source
        self._predictions = predictions
        self._result_config = result_config
        self._sink = sink
        self._service_account = service_account
        self._resource_request = resource_request
        self._env_vars = env_vars

    def job_spec(self) -> turing.generated.models.EnsemblingJobSpec:
        source = turing.generated.models.EnsemblingJobSource(
            dataset=self._source.dataset,
            join_on=self._source.join_on
        )

        predictions = {
            name: turing.generated.models.EnsemblingJobPredictionSource(
                dataset=source.dataset,
                join_on=source.join_on,
                columns=source.columns
            )
            for name, source in self._predictions.items()
        }

        ensembler = turing.generated.models.EnsemblingJobEnsemblerSpec(
            result=self._result_config
        )

        sink = self._sink.to_open_api()

        return turing.generated.models.EnsemblingJobSpec(
            source=source,
            predictions=predictions,
            ensembler=ensembler,
            sink=sink
        )

    def infra_spec(self) -> turing.generated.models.EnsemblerInfraConfig:
        return turing.generated.models.EnsemblerInfraConfig(
            service_account_name=self._service_account,
            resources=self._resource_request
        )
