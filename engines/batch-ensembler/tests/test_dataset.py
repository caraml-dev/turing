import textwrap
import pytest
from ensembler.api.proto.v1 import batch_ensembling_job_pb2 as pb2
from ensembler.dataset import DataSet, BigQueryDataSet
from tests.utils.proto_utils import from_yaml


@pytest.fixture()
def config():
    return from_yaml("""\
    type: BQ
    bqConfig:
      table: "project.dataset.table"
      features:
        - customer_id
        - target_date
        - predictions
    """, pb2.Dataset())


def test_load_from_config(config):
    dataset = DataSet.from_config(config=config)

    assert dataset.type() == pb2.Dataset.BQ
    assert isinstance(dataset, BigQueryDataSet)
    assert dataset.query == textwrap.dedent("""\
    SELECT
        customer_id,
        target_date,
        predictions
    FROM `project.dataset.table`""")
    assert len(dataset.options) == 0
