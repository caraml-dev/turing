import pytest
import turing.generated.models as openapi
import turing.batch.config as sdk
from ensembler.sink import Sink, BigQuerySink
from tests.utils.openapi_utils import from_yaml


@pytest.fixture(scope="session")
def sink_config():
    return from_yaml(
        """\
    type: BQ
    save_mode: OVERWRITE
    columns:
      - customer_id
      - ensembling_result
    bq_config:
      table: "project.dataset.table"
      staging_bucket: "bucket_name"
      options:
        partitionField: target_date
    """,
        openapi.EnsemblingJobSink,
    )


def test_load_from_config(sink_config):
    sink = Sink.from_config(sink_config)

    assert sink.type == sdk.sink.BigQuerySink.TYPE
    assert isinstance(sink, BigQuerySink)
    assert sink.save_mode == sdk.sink.SaveMode.OVERWRITE
    assert sink.columns == ["customer_id", "ensembling_result"]
    assert sink.options == {
        "table": "project.dataset.table",
        "temporaryGcsBucket": "bucket_name",
        "partitionField": "target_date",
    }
