import pytest
import turing.router.config.common.schemas

from turing.router.config.common.schemas import DockerImageSchema, EnvVarNameSchema, TimeoutSchema


@pytest.mark.parametrize(
    "value,expected", [
        pytest.param(
            "echo:1:1:2",
            turing.router.config.common.schemas.InvalidImageException
        ),
        pytest.param(
            "echo:1-0-2!",
            turing.router.config.common.schemas.InvalidImageException
        )
    ]
)
def test_invalid_docker_image_value(value, expected):
    with pytest.raises(expected):
        DockerImageSchema.verify_regex(value)


@pytest.mark.parametrize(
    "value,expected", [
        pytest.param(
            "not-a-good-name",
            turing.router.config.common.schemas.InvalidEnvironmentVariableNameException
        ),
        pytest.param(
            "notagoodnameeither!",
            turing.router.config.common.schemas.InvalidEnvironmentVariableNameException
        )
    ]
)
def test_invalid_env_var_name_value(value, expected):
    with pytest.raises(expected):
        EnvVarNameSchema.verify_regex(value)


@pytest.mark.parametrize(
    "value,expected", [
        pytest.param(
            "500gs",
            turing.router.config.common.schemas.InvalidTimeoutException
        ),
        pytest.param(
            "5s5",
            turing.router.config.common.schemas.InvalidTimeoutException
        )
    ]
)
def test_invalid_timeout_value(value, expected):
    with pytest.raises(expected):
        TimeoutSchema.verify_regex(value)
