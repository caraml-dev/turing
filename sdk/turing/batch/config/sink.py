import abc
from enum import Enum
from typing import Iterable, MutableMapping, Optional
import turing.generated.models
from turing.generated.model_utils import OpenApiModel


class SaveMode(Enum):
    """
    Configuration that specifies the mode of saving results
    See: https://spark.apache.org/docs/latest/api/java/index.html?org/apache/spark/sql/SaveMode.html
    """
    ERRORIFEXISTS = 0
    OVERWRITE = 1
    APPEND = 2
    IGNORE = 3

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.SaveMode(self.name)


class EnsemblingJobSink:
    """
    Abstract sink configuration of the ensembling job
    """

    def __init__(self, save_mode: SaveMode = None):
        self._save_mode = save_mode
        self._columns = []

    def save_mode(self, save_mode: SaveMode) -> 'EnsemblingJobSink':
        """
        Configure `save_mode` of the sink

        :param save_mode:
        :return: instance of the sink with configured `save_mode`
        """
        self._save_mode = save_mode
        return self

    def select(self, columns: Iterable[str]) -> 'EnsemblingJobSink':
        """
        Configure columns, that would be written into the ensembling results destination

        :param columns: list of columns
        :return: instance of the sink with configured `columns`
        """
        self._columns = columns
        return self

    @abc.abstractmethod
    def to_open_api(self) -> OpenApiModel:
        """
        Converts EnsemblingJobSink into a corresponding openapi schema model

        :return: instance of OpenApiModel
        """
        pass


class BigQuerySink(EnsemblingJobSink):
    """
    BigQuery Sink configuration
    """

    TYPE = "BQ"

    def __init__(
            self,
            table: str,
            staging_bucket: str,
            options: MutableMapping[str, str] = None):
        """
        :param table: fully-qualified name of the BQ table, where results will be written to
        :param staging_bucket: temporary GCS bucket for staging write into BQ table
        :param options: additional sink option to configure the prediction job
        """
        super(BigQuerySink, self).__init__()
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
            save_mode=self._save_mode.to_open_api(),
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
