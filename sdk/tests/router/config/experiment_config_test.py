import pytest

import turing.generated.models
from turing.router.config.experiment_config import ExperimentConfig
from turing.router.config.common.env_var import EnvVar
from turing.router.config.resource_request import ResourceRequest


@pytest.mark.parametrize(
    "type,config,plugin_config,expected", [
        pytest.param(
            "nop",
            {
                "project_id": 102.0,
                "variables": [
                    {
                        "test": 1
                    },
                    {
                        "count": 200,
                        "id": "random_variable"
                    }
                ]
            },
            {
                "image": "asia.test.io/gods-test/turing-ensembler:0.0.0-build.0"
            },
            turing.generated.models.ExperimentConfig(
                type="nop",
                config={
                    "project_id": 102,
                    "variables": [
                        {
                            "test": 1
                        },
                        {
                            "count": 200,
                            "id": "random_variable"
                        }
                    ]
                },
                plugin_config=turing.generated.models.ExperimentConfigPluginConfig(
                    image="asia.test.io/gods-test/turing-ensembler:0.0.0-build.0"
                )
            ),
        )
    ])
def test_create_experiment_config(type, config, plugin_config, expected):
    actual = ExperimentConfig(
        type=type,
        config=config,
        plugin_config=plugin_config
    ).to_open_api()
    assert actual == expected
