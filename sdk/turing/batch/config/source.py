import abc
from typing import Iterable, MutableMapping, Optional
import turing.generated.models
from turing.generated.model_utils import OpenApiModel


class EnsemblingJobSource:
    """
    Configuration of source of the ensembling job
    """

    def __init__(self, dataset: 'Dataset', join_on: Iterable[str]):
        self._dataset = dataset
        self._join_on = join_on

    @property
    def dataset(self) -> 'Dataset':
        return self._dataset

    @property
    def join_on(self) -> Iterable[str]:
        return self._join_on

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.EnsemblingJobSource(
            dataset=self.dataset.to_open_api(),
            join_on=self.join_on
        )

    def select(self, columns: Iterable[str]) -> 'EnsemblingJobPredictionSource':
        """
        Creates an instance of prediction source configuration

        :param columns: list of columns from this source, that contain prediction data
        :return: instance of `EnsemblingJobPredictionSource`
        """
        return EnsemblingJobPredictionSource(
            self.dataset,
            self.join_on,
            columns
        )


class EnsemblingJobPredictionSource(EnsemblingJobSource):
    """
    Configuration of the prediction data for the ensembling job
    """

    def __init__(self, dataset, join_on, columns):
        super(EnsemblingJobPredictionSource, self).__init__(dataset, join_on)
        self._columns = columns

    @property
    def columns(self) -> Iterable[str]:
        return self._columns

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.EnsemblingJobPredictionSource(
            dataset=self.dataset.to_open_api(),
            join_on=self.join_on,
            columns=self.columns
        )


class Dataset(abc.ABC):
    """
    Abstract dataset
    """

    def join_on(self, columns: Iterable[str]) -> 'EnsemblingJobSource':
        """
        Create ensembling job source configuration from this dataset,
        by specifying how this dataset could be joined with the
        datasets containing predictions of individual models

        :param columns: list of columns, that would be used to join this
            dataset with predictions data
        :return: instance of ensembling job source configuration
        """
        pass

    def to_open_api(self) -> OpenApiModel:
        pass


class BigQueryDataset(Dataset):
    """
    BigQuery dataset configuration
    """

    TYPE = "BQ"

    def __init__(self,
                 table: str = None,
                 features: Iterable[str] = None,
                 query: str = None,
                 options: MutableMapping[str, str] = None):
        """
        Create new instance of BigQuery dataset

        :param table: fully-qualified BQ table id e.g. `gcp-project.dataset.table_name`
        :param features: list of columns from the `table` to be selected for this dataset
        :param query: (optional) Alternatively, dataset can be defined by BQ standard SQL query.
             This allows to define dataset from the data, stored in multiple tables
        :param options: (optional) Additional BQ options to configure the dataset
        """
        self._table = table
        self._query = query
        self._features = features
        self._options = options

    @property
    def table(self) -> Optional[str]:
        return self._table

    @property
    def query(self) -> Optional[str]:
        return self._query

    @property
    def features(self) -> Optional[Iterable[str]]:
        return self._features

    @property
    def options(self) -> Optional[MutableMapping[str, str]]:
        return self._options

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.BigQueryDataset(
            type=BigQueryDataset.TYPE,
            bq_config=turing.generated.models.BigQueryDatasetConfig(
                table=self.table,
                query=self.query,
                features=self.features,
                options=self.options
            )
        )

    def join_on(self, columns: Iterable[str]) -> 'EnsemblingJobSource':
        return EnsemblingJobSource(
            dataset=self,
            join_on=columns
        )
