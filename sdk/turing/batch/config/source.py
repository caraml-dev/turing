import abc
from typing import Iterable, MutableMapping, Optional
import turing.generated.models
from turing.generated.model_utils import OpenApiModel
from turing._base_types import DataObject


class EnsemblingJobSource:

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
        return EnsemblingJobPredictionSource(
            self._dataset,
            self._join_on,
            columns
        )


class EnsemblingJobPredictionSource(EnsemblingJobSource):
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


class Dataset(DataObject, abc.ABC):

    def join_on(self, columns: Iterable[str]) -> 'EnsemblingJobSource':
        pass

    def to_open_api(self) -> OpenApiModel:
        pass


class BigQueryDataset(Dataset):
    _TYPE = "BQ"

    def __init__(self,
                 table: str = None,
                 query: str = None,
                 features: Iterable[str] = None,
                 options: MutableMapping[str, str] = None):
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
            type=BigQueryDataset._TYPE,
            bq_config=turing.generated.models.BigQueryDatasetConfig(
                table=self._table,
                query=self._query,
                features=self._features,
                options=self._options
            )
        )

    def join_on(self, columns: Iterable[str]) -> 'EnsemblingJobSource':
        return EnsemblingJobSource(
            dataset=self,
            join_on=columns
        )
