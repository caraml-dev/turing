import pytest
import turing.batch.config.source


@pytest.mark.parametrize(
    "table,query,features,options,expected", [
        pytest.param(
            "project.table.dataset_1",
            None,
            ["feature_1", "feature_2"],
            None,
            {
                "table": "project.table.dataset_1",
                "features": ["feature_1", "feature_2"],
            },
            id="Initialize BQ dataset table and list of features"
        ),
        pytest.param(
            None,
            "SELECT * FROM `project.dataset.table`",
            None,
            {
                "viewsEnabled": "true",
                "materializationDataset": "my_dataset"
            },
            {
                "query": "SELECT * FROM `project.dataset.table`",
                "options": {
                    "viewsEnabled": "true",
                    "materializationDataset": "my_dataset"
                },
            },
            id="Initialize BQ dataset from query"
        )
    ]
)
def test_bq_dataset(table, query, features, options, expected):
    dataset = turing.batch.config.source.BigQueryDataset(
        table, features, query, options
    )

    assert dataset.to_dict() == expected
