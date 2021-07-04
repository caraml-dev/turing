import abc
from typing import Iterable, MutableMapping


class EnsemblingJobSource:

    def __init__(self, dataset, join_on):
        self._dataset = dataset
        self._join_on = join_on

    @property
    def dataset(self):
        return self._dataset

    @property
    def join_on(self) -> Iterable[str]:
        return self._join_on

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


class Dataset(abc.ABC):

    def join_on(self, columns: Iterable[str]) -> 'EnsemblingJobSource':
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

    def join_on(self, columns: Iterable[str]) -> 'EnsemblingJobSource':
        import turing.generated.models
        dataset = turing.generated.models.BigQueryDataset(
            type=BigQueryDataset._TYPE,
            bq_config=turing.generated.models.BigQueryDatasetConfig(
                table=self._table,
                query=self._query,
                features=self._features,
                options=self._options
            )
        )

        return EnsemblingJobSource(
            dataset=dataset,
            join_on=columns
        )
