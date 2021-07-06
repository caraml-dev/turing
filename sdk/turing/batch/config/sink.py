import abc
from typing import Iterable, MutableMapping
import turing.generated.models


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

        self._bq_config = turing.generated.models.BigQuerySinkConfig(
            table=table,
            staging_bucket=staging_bucket,
            options=options
        )

    def to_open_api(self):
        import turing.generated.models
        return turing.generated.models.BigQuerySink(
            type=self._type,
            save_mode=self._save_mode,
            columns=self._columns,
            bq_config=self._bq_config
        )


__all__ = [
    "SaveMode", "BigQuerySink",
]
