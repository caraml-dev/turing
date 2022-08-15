from typing import Dict, MutableMapping, Tuple
from pyspark.sql import SparkSession
import yaml
from .source import Source, PredictionSource
from .ensembler import Ensembler
from .sink import Sink
from turing.generated.models import EnsemblerConfig, EnsemblingJobMeta
from turing.generated.model_utils import validate_and_convert_types
from turing.generated.api_client import Configuration


class BatchEnsemblingJob:
    def __init__(
        self,
        metadata: EnsemblingJobMeta,
        source: "Source",
        predictions: Dict[str, "PredictionSource"],
        ensembler: "Ensembler",
        sink: "Sink",
    ):
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
        combined_df = self.source.join(**self.predictions).load(spark)
        result_df = self.ensembler.ensemble(combined_df, spark)
        self.sink.save(result_df)

    @classmethod
    def from_yaml(cls, spec_path: str) -> Tuple["BatchEnsemblingJob", str]:
        with open(spec_path, "r") as file:
            job_spec_raw = file.read()

        parsed_data = yaml.safe_load(job_spec_raw)
        job_spec = validate_and_convert_types(
            input_value=parsed_data,
            required_types_mixed=(EnsemblerConfig,),
            path_to_item=[],
            spec_property_naming=True,
            _check_type=True,
            configuration=Configuration(),
        )
        return BatchEnsemblingJob.from_config(job_spec), job_spec_raw

    @classmethod
    def from_config(cls, config: EnsemblerConfig) -> "BatchEnsemblingJob":
        metadata = config.metadata
        source = Source.from_config(config.spec.source)
        predictions: Dict[str, "PredictionSource"] = {
            k: PredictionSource.from_config(v)
            for k, v in config.spec.predictions.items()
        }
        ensembler = Ensembler.from_config(config.spec.ensembler)
        sink = Sink.from_config(config.spec.sink)

        return BatchEnsemblingJob(
            metadata=metadata,
            source=source,
            predictions=predictions,
            ensembler=ensembler,
            sink=sink,
        )
