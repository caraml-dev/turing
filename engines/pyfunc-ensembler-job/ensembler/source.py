import os
import logging
from typing import List, TypeVar, Generic
from pyspark.sql import DataFrame, SparkSession
from turing.ensembler import PyFunc
from .dataset import Dataset, BigQueryDataset, jinja
import turing.generated.models as openapi

T = TypeVar("T", bound="Dataset")


class Source(Generic[T]):
    def __init__(self, dataset: T, join_on_columns: List[str]):
        self._dataset = dataset
        self._join_columns = join_on_columns
        self._logger = logging.getLogger("ensembler.Source")

    @property
    def dataset(self) -> T:
        return self._dataset

    @property
    def join_columns(self):
        return self._join_columns

    def load(self, spark: SparkSession) -> DataFrame:
        return self.dataset.load(spark)

    def join(self, **predictions: "PredictionSource") -> "Source[T]":
        raise NotImplementedError

    @classmethod
    def from_config(cls, config: openapi.EnsemblingJobSource) -> "Source":
        dataset = Dataset.from_config(config.dataset)

        if isinstance(dataset, BigQueryDataset):
            return BigQuerySource(dataset, config.join_on)
        return Source(dataset, config.join_on)


class BigQuerySource(Source["BigQueryDataset"]):
    with open(
        os.path.join(os.path.dirname(__file__), "sql", "bq_join.sql.jinja2"), "r"
    ) as _t:
        _SQL_TEMPLATE = _t.read()

    def __init__(self, dataset: "BigQueryDataset", join_on_columns: List[str]):
        super().__init__(dataset, join_on_columns)

    def join(self, **predictions: "PredictionSource") -> "Source[BigQueryDataset]":
        template, bind_params = jinja.prepare_query(
            self._SQL_TEMPLATE,
            {
                "features_query": self.dataset.query,
                "join_columns": self.join_columns,
                "predictions": predictions,
                "prefix": PyFunc.PREDICTION_COLUMN_PREFIX,
            },
        )

        query = template % bind_params
        self._logger.debug(f"Query to fetch data and predictions:\n" f"{query}\n")

        return BigQuerySource(
            BigQueryDataset(query, self.dataset.options), self.join_columns
        )


class PredictionSource(Source[T]):
    def __init__(self, dataset: T, join_on_columns, prediction_columns: List[str]):
        super().__init__(dataset, join_on_columns)
        self._prediction_columns = prediction_columns

    @property
    def prediction_columns(self):
        return self._prediction_columns

    @classmethod
    def from_config(
        cls, config: openapi.EnsemblingJobPredictionSource
    ) -> "PredictionSource":
        dataset = Dataset.from_config(config.dataset)
        return PredictionSource(dataset, config.join_on, config.columns)
