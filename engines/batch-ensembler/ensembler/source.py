import os
from abc import abstractmethod
from typing import List, TypeVar, Generic, SupportsAbs
from pyspark.sql import DataFrame, SparkSession
import ensembler.api.proto.v1.batch_ensembling_job_pb2 as pb2
from ensembler.components.experimentation import PREDICTION_COLUMN_PREFIX
from ensembler.dataset import DataSet, BigQueryDataSet, jinja

T = TypeVar('T', bound=SupportsAbs['DataSet'])


class Source(Generic[T]):
    def __init__(self, dataset: T, join_on_columns: List[str]):
        self._dataset = dataset
        self._join_columns = join_on_columns

    def dataset(self) -> T:
        return self._dataset

    def join_columns(self):
        return self._join_columns

    def load(self, spark: SparkSession) -> DataFrame:
        return self.dataset().load(spark)

    @abstractmethod
    def join(self, **predictions: 'PredictionSource') -> 'Source':
        pass

    @classmethod
    def from_config(cls, config: pb2.Source) -> 'Source':
        dataset = DataSet.from_config(config.dataset)

        if isinstance(dataset, BigQueryDataSet):
            return BigQuerySource(dataset, config.join_on)
        return Source[type(dataset)](dataset, config.join_on)


class BigQuerySource(Source['BigQueryDataSet']):
    with open(os.path.join(os.path.dirname(__file__), 'sql', 'bq_join.sql.jinja2'), 'r') as template:
        _SQL_TEMPLATE = template.read()

    def __init__(self,
                 dataset: 'BigQueryDataSet',
                 join_on_columns: List[str]):
        super().__init__(dataset, join_on_columns)

    def join(self, **predictions: 'PredictionSource') -> 'Source[BigQueryDataSet]':
        template, bind_params = jinja.prepare_query(
            self._SQL_TEMPLATE,
            {
                'features_query': self.dataset().query,
                'join_columns': self.join_columns(),
                'predictions': predictions,
                'prefix': PREDICTION_COLUMN_PREFIX
            }
        )

        query = template % bind_params
        print(query)

        return BigQuerySource(
            BigQueryDataSet(query, self.dataset().options),
            self.join_columns()
        )


class PredictionSource(Source[T]):

    def __init__(self, dataset: T, join_on_columns, prediction_columns: List[str]):
        super(PredictionSource, self).__init__(dataset, join_on_columns)
        self._prediction_columns = prediction_columns

    def prediction_columns(self):
        return self._prediction_columns

    def join(self, **predictions: 'PredictionSource') -> 'Source[T]':
        return super(PredictionSource, self).join(**predictions)

    @classmethod
    def from_config(cls, config: pb2.PredictionSource) -> 'PredictionSource[T]':
        dataset = DataSet.from_config(config.dataset)
        return PredictionSource(dataset, config.join_on, config.columns)

