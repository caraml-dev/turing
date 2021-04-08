from typing import Any, Dict, MutableMapping
import yaml
from pyspark.sql import SparkSession
import ensembler.api.proto.v1.batch_ensembling_job_pb2 as pb2
from ensembler.source import Source, PredictionSource
from ensembler.ensembler import Ensembler
from ensembler.sink import Sink
from google.protobuf import json_format


class BatchEnsemblingJob(object):
    def __init__(
            self,
            metadata: pb2.BatchEnsemblingJobMetadata,
            source: 'Source',
            predictions: Dict[str, 'PredictionSource'],
            ensembler: 'Ensembler',
            sink: 'Sink'):
        self.metadata = metadata
        self.source = source
        self.predictions = predictions
        self.ensembler = ensembler
        self.sink = sink

    def name(self) -> str:
        return self.metadata.name

    def annotations(self) -> MutableMapping[str, str]:
        return self.metadata.annotations

    def run(self, spark: SparkSession):
        combined_df = self.source \
            .join(**self.predictions) \
            .load(spark)
        result_df = self.ensembler.ensemble(combined_df, spark)
        self.sink.save(result_df)

    @classmethod
    def from_yaml(cls, spec_path: str) -> ('BatchEnsemblingJob', Dict[str, Any]):
        with open(spec_path, 'r') as f:
            job_spec_dict = yaml.safe_load(f)

        job_config = json_format.ParseDict(job_spec_dict, pb2.BatchEnsemblingJob())
        return BatchEnsemblingJob.from_config(job_config), job_spec_dict

    @classmethod
    def from_config(cls, config: pb2.BatchEnsemblingJob) -> 'BatchEnsemblingJob':
        metadata = config.metadata
        source = Source.from_config(config.spec.source)
        predictions = {k: PredictionSource.from_config(v) for k, v in config.spec.predictions.items()}
        ensembler = Ensembler.from_config(config.spec.ensembler)
        sink = Sink.from_config(config.spec.sink)

        return BatchEnsemblingJob(
            metadata=metadata,
            source=source,
            predictions=predictions,
            ensembler=ensembler,
            sink=sink
        )
