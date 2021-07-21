import textwrap
import pytest
from ensembler.dataset import DataSet, BigQueryDataSet
from tests.utils.openapi_utils import from_yaml
import turing.generated.models as openapi
import turing.batch.config as sdk


@pytest.fixture()
def config():
    return from_yaml("""\
    type: BQ
    bq_config:
      table: "project.dataset.table"
      features:
        - customer_id
        - target_date
        - predictions
    """, openapi.Dataset)


def test_load_from_config(config):
    print(config)
    dataset = DataSet.from_config(config=config)

    assert dataset.type() == sdk.BigQueryDataset.TYPE
    assert isinstance(dataset, BigQueryDataSet)
    assert dataset.query == textwrap.dedent("""\
    SELECT
        customer_id,
        target_date,
        predictions
    FROM `project.dataset.table`""")
    assert len(dataset.options) == 0
