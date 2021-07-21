from abc import ABC, abstractmethod
from typing import List
from pyspark.sql import DataFrame
import turing.batch.config as sdk
import turing.generated.models as openapi


class Sink(ABC):
    def __init__(self, save_mode: sdk.SaveMode, columns: List[str] = None):
        self._save_mode = sdk.SaveMode.Name(save_mode).lower()
        self._columns = columns

    def save(self, df: DataFrame):
        if self._columns:
            df = df.selectExpr(*self._columns)
        self._save(df)

    @abstractmethod
    def _save(self, df: DataFrame):
        pass

    @classmethod
    def from_config(cls, config: openapi.EnsemblingJobSink):
        if config.type == sdk.BigQuerySink.TYPE:
            return BigQuerySink(config.save_mode, config.columns, config.bq_config)
        raise ValueError(f'Sink not implemented: {config.type}')


class ConsoleSink(Sink):
    def __init__(self, columns: List[str] = None):
        super().__init__(save_mode=None, columns=columns)

    def _save(self, df: DataFrame):
        df.show()


class BigQuerySink(Sink):
    _WRITE_FORMAT = 'bigquery'
    _OPTION_NAME_TABLE = 'table'
    _OPTION_NAME_STAGING_BUCKET = 'temporaryGcsBucket'

    def __init__(
            self,
            save_mode: sdk.SaveMode,
            columns: List[str],
            config: openapi.BigQuerySinkConfig):
        super().__init__(save_mode=save_mode, columns=columns)

        self._options = {
            **config.options,
            self._OPTION_NAME_STAGING_BUCKET: config.staging_bucket,
            self._OPTION_NAME_TABLE: config.table,
        }

    def _save(self, df: DataFrame):
        df.write \
            .mode(self._save_mode) \
            .format(self._WRITE_FORMAT) \
            .options(**self._options) \
            .save()
