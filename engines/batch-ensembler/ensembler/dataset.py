import os
from abc import ABC, abstractmethod
from typing import MutableMapping
from pyspark.sql import DataFrame, SparkSession
from jinjasql import JinjaSql
import turing.generated.models as openapi
import turing.batch.config as sdk

__all__ = ['DataSet', 'BigQueryDataSet', 'jinja']

jinja = JinjaSql(param_style='pyformat')
jinja.env.filters['zip'] = zip


class DataSet(ABC):

    @abstractmethod
    def load(self, spark: SparkSession) -> DataFrame:
        pass

    @abstractmethod
    def type(self) -> str:
        pass

    @classmethod
    def from_config(cls, config: openapi.Dataset) -> 'DataSet':
        if config.type == sdk.source.BigQueryDataset.TYPE:
            return BigQueryDataSet.from_config(config.bq_config)
        raise ValueError(f'Unknown dataset type: {config.type} is not implemented')


class BigQueryDataSet(DataSet):
    with open(os.path.join(os.path.dirname(__file__), 'sql', 'bq_select.sql.jinja2'), 'r') as _f:
        _SQL_TEMPLATE = _f.read()

    _READ_FORMAT = 'bigquery'
    _OPTION_QUERY = 'query'

    def __init__(self, query: str, options: MutableMapping[str, str]):
        self.query = query
        self.options = options

    def type(self):
        return sdk.source.BigQueryDataset.TYPE

    def load(self, spark: SparkSession) -> DataFrame:
        return spark.read \
            .format(BigQueryDataSet._READ_FORMAT) \
            .options(**self.options) \
            .option(BigQueryDataSet._OPTION_QUERY, self.query) \
            .load()

    @classmethod
    def from_config(cls, config: openapi.BigQueryDatasetConfig) -> 'BigQueryDataSet':
        if config.get('query', ""):
            query = config.query
        elif config.table:
            template, bind_params = jinja.prepare_query(
                cls._SQL_TEMPLATE,
                {'table': config.table, 'columns': config.features}
            )
            query = template % bind_params
        else:
            raise ValueError(
                'Dataset initialization failed: '
                'either "query" or "table" should be provided'
            )
        return BigQueryDataSet(query, config.get('options', {}))
