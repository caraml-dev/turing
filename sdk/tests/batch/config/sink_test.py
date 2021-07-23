import pytest
import turing.batch.config.sink
import turing.generated.models


@pytest.mark.parametrize(
    "table,staging_bucket,options,save_mode,columns,expected", [
        pytest.param(
            "project.dataset.results_table",
            "my-staging_bucket",
            {
                "partitionField": "target_date"
            },
            turing.batch.config.sink.SaveMode.APPEND,
            ["target_date", "user_id", "prediction_score"],
            turing.generated.models.BigQuerySink(
                type=turing.batch.config.sink.BigQuerySink.TYPE,
                save_mode=turing.generated.models.SaveMode("APPEND"),
                columns=["target_date", "user_id", "prediction_score"],
                bq_config=turing.generated.models.BigQuerySinkConfig(
                    table="project.dataset.results_table",
                    staging_bucket="my-staging_bucket",
                    options={
                        "partitionField": "target_date"
                    }
                )
            ),
            id="Initialize BQ sink"
        )
    ]
)
def test_create_bq_sink(table, staging_bucket, options, save_mode, columns, expected):
    sink = turing.batch.config.sink.BigQuerySink(table, staging_bucket, options)\
        .select(columns)\
        .save_mode(save_mode)

    assert sink.to_open_api() == expected
