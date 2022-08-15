import textwrap
import pytest
from ensembler.dataset import Dataset, BigQueryDataset
from tests.utils.openapi_utils import from_yaml
import turing.generated.models as openapi
import turing.batch.config as sdk


@pytest.fixture(scope="session")
def dataset_config():
    return from_yaml(
        """\
    type: BQ
    bq_config:
      table: "project.dataset.table"
      features:
        - customer_id
        - target_date
        - predictions
    """,
        openapi.Dataset,
    )


def test_load_from_config(dataset_config):
    dataset = Dataset.from_config(config=dataset_config)

    assert dataset.type == sdk.source.BigQueryDataset.TYPE
    assert isinstance(dataset, BigQueryDataset)
    assert dataset.query == textwrap.dedent(
        """\
    SELECT
        customer_id,
        target_date,
        predictions
    FROM `project.dataset.table`"""
    )
    assert len(dataset.options) == 0
