from abc import ABC, abstractmethod
from typing import List, MutableMapping, Optional
from pyspark.sql import DataFrame
import turing.batch.config as sdk
import turing.generated.models as openapi


class Sink(ABC):
    def __init__(self, save_mode: sdk.sink.SaveMode, columns: Optional[List[str]] = None):
        self._save_mode = save_mode
        self._columns = columns

    @property
    def type(self) -> str:
        return ""

    @property
    def save_mode(self) -> sdk.sink.SaveMode:
        return self._save_mode

    @property
    def columns(self) -> Optional[List[str]]:
        return self._columns

    def save(self, df: DataFrame):
        if self._columns:
            df = df.selectExpr(*self._columns)
        self._save(df)

    @abstractmethod
    def _save(self, df: DataFrame):
        pass

    @classmethod
    def from_config(cls, config: openapi.EnsemblingJobSink):
        if config.type == sdk.sink.BigQuerySink.TYPE:
            return BigQuerySink(
                getattr(sdk.sink.SaveMode, config.save_mode.to_str()),
                config.columns,
                config.bq_config,
            )
        raise ValueError(f"Sink not implemented: {config.type}")


class ConsoleSink(Sink):
    def __init__(self, columns: Optional[List[str]] = None):
        super(ConsoleSink, self).__init__(save_mode=None, columns=columns)

    @property
    def type(self) -> str:
        return "CONSOLE"

    def _save(self, df: DataFrame):
        df.show()


class BigQuerySink(Sink):
    _WRITE_FORMAT = "bigquery"
    _OPTION_NAME_TABLE = "table"
    _OPTION_NAME_STAGING_BUCKET = "temporaryGcsBucket"

    def __init__(
        self,
        save_mode: sdk.sink.SaveMode,
        columns: List[str],
        config: openapi.BigQuerySinkConfig,
    ):
        super(BigQuerySink, self).__init__(save_mode=save_mode, columns=columns)

        self._options = {
            **config.options,
            self._OPTION_NAME_STAGING_BUCKET: config.staging_bucket,
            self._OPTION_NAME_TABLE: config.table,
        }

    @property
    def type(self) -> str:
        return sdk.sink.BigQuerySink.TYPE

    @property
    def options(self) -> MutableMapping[str, str]:
        return self._options

    def _save(self, df: DataFrame):
        df.write.mode(self.save_mode.name.lower()).format(self._WRITE_FORMAT).options(
            **self.options
        ).save()
