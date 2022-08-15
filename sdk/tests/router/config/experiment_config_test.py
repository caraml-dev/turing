import pytest

import turing.generated.models
from turing.router.config.experiment_config import ExperimentConfig


@pytest.mark.parametrize(
    "type,config,expected",
    [
        pytest.param(
            "nop",
            {
                "project_id": 102.0,
                "variables": [{"test": 1}, {"count": 200, "id": "random_variable"}],
            },
            turing.generated.models.ExperimentConfig(
                type="nop",
                config={
                    "project_id": 102,
                    "variables": [{"test": 1}, {"count": 200, "id": "random_variable"}],
                },
            ),
        )
    ],
)
def test_create_experiment_config(type, config, expected):
    actual = ExperimentConfig(type=type, config=config).to_open_api()
    assert actual == expected
