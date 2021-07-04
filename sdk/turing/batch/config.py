from typing import Dict
import turing.generated.models
from turing.batch.source import EnsemblingJobSource, EnsemblingJobPredictionSource
from turing.batch.sink import EnsemblingJobSink

EnsemblingJobResourceRequest = turing.generated.models.EnsemblingResources
EnsemblingJobResultConfig = turing.generated.models.EnsemblingJobEnsemblerSpecResult


class ResultType:
    DOUBLE = turing.generated.models.EnsemblingJobResultType("DOUBLE")
    FLOAT = turing.generated.models.EnsemblingJobResultType("FLOAT")
    INTEGER = turing.generated.models.EnsemblingJobResultType("INTEGER")
    LONG = turing.generated.models.EnsemblingJobResultType("LONG")
    STRING = turing.generated.models.EnsemblingJobResultType("STRING")
    ARRAY = turing.generated.models.EnsemblingJobResultType("ARRAY")


class EnsemblingJobConfig:
    def __init__(self,
                 source: EnsemblingJobSource,
                 predictions: Dict[str, EnsemblingJobPredictionSource],
                 result_config: EnsemblingJobResultConfig,
                 sink: EnsemblingJobSink,
                 service_account: str,
                 resource_request: EnsemblingJobResourceRequest = None,
                 env_vars: Dict[str, str] = None):
        self._source = source
        self._predictions = predictions
        self._result_config = result_config
        self._sink = sink
        self._service_account = service_account
        self._resource_request = resource_request
        self._env_vars = env_vars

    @property
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
