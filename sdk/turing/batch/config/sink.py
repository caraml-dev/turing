import abc
from typing import Iterable, MutableMapping, Optional
import turing.generated.models
from turing.generated.model_utils import OpenApiModel


class SaveMode:
    ERRORIFEXISTS = turing.generated.models.SaveMode("ERRORIFEXISTS")
    OVERWRITE = turing.generated.models.SaveMode("OVERWRITE")
    APPEND = turing.generated.models.SaveMode("APPEND")
    IGNORE = turing.generated.models.SaveMode("IGNORE")


class EnsemblingJobSink:
    def __init__(self, type: str, save_mode: SaveMode = None):
        self._type = type
        self._save_mode = save_mode
        self._columns = []

    def save_mode(self, save_mode: SaveMode) -> 'EnsemblingJobSink':
        self._save_mode = save_mode
        return self

    def select(self, columns: Iterable[str]) -> 'EnsemblingJobSink':
        self._columns = columns
        return self

    @abc.abstractmethod
    def to_open_api(self):
        pass


class BigQuerySink(EnsemblingJobSink):
    _TYPE_ = "BQ"

    def __init__(
            self,
            table: str,
            staging_bucket: str,
            options: MutableMapping[str, str] = None):
        super(BigQuerySink, self).__init__(type=BigQuerySink._TYPE_)
        self._table = table
        self._staging_bucket = staging_bucket
        self._options = options

    @property
    def table(self) -> str:
        return self._table

    @property
    def staging_bucket(self) -> str:
        return self._staging_bucket

    @property
    def options(self) -> Optional[MutableMapping[str, str]]:
        return self._options

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.BigQuerySink(
            type=self._type,
            save_mode=self._save_mode,
            columns=self._columns,
            bq_config=turing.generated.models.BigQuerySinkConfig(
                table=self.table,
                staging_bucket=self.staging_bucket,
                options=self.options
            )
        )


__all__ = [
    "SaveMode", "BigQuerySink",
]
