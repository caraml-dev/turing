import pytest
import turing.batch.config.source
import turing.generated.models


@pytest.mark.parametrize(
    "table,features,query,options,expected", [
        pytest.param(
            "project.table.dataset_1",
            ["feature_1", "feature_2"],
            None,
            None,
            turing.generated.models.BigQueryDataset(
                type="BQ",
                bq_config=turing.generated.models.BigQueryDatasetConfig(
                    table="project.table.dataset_1",
                    features=["feature_1", "feature_2"],
                    query=None,
                    options=None
                )
            ),
            id="Initialize BQ dataset table and list of features"
        ),
        pytest.param(
            None,
            None,
            "SELECT * FROM `project.dataset.table`",
            {
                "viewsEnabled": "true",
                "materializationDataset": "my_dataset"
            },
            turing.generated.models.BigQueryDataset(
                type="BQ",
                bq_config=turing.generated.models.BigQueryDatasetConfig(
                    query="SELECT * FROM `project.dataset.table`",
                    options={
                       "viewsEnabled": "true",
                       "materializationDataset": "my_dataset"
                    },
                    table=None,
                    features=None
                )
            ),
            id="Initialize BQ dataset from query"
        )
    ]
)
def test_bq_dataset(table, query, features, options, expected):
    dataset = turing.batch.config.source.BigQueryDataset(
        table, features, query, options
    )

    assert dataset.to_open_api() == expected


@pytest.mark.parametrize(
    "dataset,join_on,expected_fn", [
        pytest.param(
            turing.batch.config.BigQueryDataset(
                table="project.table.dataset_1",
                features=["feature_1", "feature_2", "feature_3"],
            ),
            ["feature_2"],
            lambda dataset, join_on: turing.generated.models.EnsemblingJobSource(
                dataset=dataset.to_open_api(),
                join_on=join_on
            ),
            id="Initialize source from BQ dataset"
        )
    ]
)
def test_bq_source(dataset, join_on, expected_fn):
    expected = expected_fn(dataset, join_on)
    assert dataset.join_on(columns=join_on).to_open_api() == expected


@pytest.mark.parametrize(
    "source,prediction_columns,expected_fn", [
        pytest.param(
            turing.batch.config.BigQueryDataset(
                query="SELECT feature_2, feature_3, score FROM `project.dataset.table`",
                options={
                    "viewsEnabled": "true",
                    "materializationDataset": "my_dataset"
                },
            ).join_on(columns=["feature_2", "feature_3"]),
            ["score"],
            lambda source, prediction_columns: turing.generated.models.EnsemblingJobPredictionSource(
                dataset=source.dataset.to_open_api(),
                join_on=source.join_on,
                columns=prediction_columns
            ),
            id="Initialize prediction source from BQ dataset"
        )
    ]
)
def test_bq_prediction_source(source, prediction_columns, expected_fn):
    expected = expected_fn(source, prediction_columns)
    assert source.select(columns=prediction_columns).to_open_api() == expected
