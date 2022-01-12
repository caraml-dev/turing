import pytest
from turing.router.config.common.schemas import *


@pytest.mark.parametrize(
    "value,expected", [
        pytest.param(
            "echo:1:1:2",
            InvalidImageException
        ),
        pytest.param(
            "echo:1-0-2!",
            InvalidImageException
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
            InvalidEnvironmentVariableNameException
        ),
        pytest.param(
            "notagoodnameeither!",
            InvalidEnvironmentVariableNameException
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
            InvalidTimeoutException
        ),
        pytest.param(
            "5s5",
            InvalidTimeoutException
        )
    ]
)
def test_invalid_timeout_value(value, expected):
    with pytest.raises(expected):
        TimeoutSchema.verify_regex(value)


@pytest.mark.parametrize(
    "value,expected", [
        pytest.param(
            "432k",
            InvalidCPURequestException
        ),
        pytest.param(
            "2322222",
            InvalidCPURequestException
        )
    ]
)
def test_invalid_cpu_request_value(value, expected):
    with pytest.raises(expected):
        CpuRequestSchema.verify_regex(value)


@pytest.mark.parametrize(
    "value,expected", [
        pytest.param(
            "50Ri",
            InvalidMemoryRequestException
        ),
        pytest.param(
            "100Kb",
            InvalidMemoryRequestException
        )
    ]
)
def test_invalid_memory_request_value(value, expected):
    with pytest.raises(expected):
        MemoryRequestSchema.verify_regex(value)


@pytest.mark.parametrize(
    "value,expected", [
        pytest.param(
            "this-is-not-a-table",
            InvalidBigQueryTableException
        ),
        pytest.param(
            "this.table.is.not_valid",
            InvalidBigQueryTableException
        )
    ]
)
def test_invalid_bigquery_table_value(value, expected):
    with pytest.raises(expected):
        BigQueryTableSchema.verify_regex(value)


@pytest.mark.parametrize(
    "value,expected", [
        pytest.param(
            "1.2.3.4:5.6.7.8,9.0.1.2:3.4.5.6",
            InvalidKafkaBrokersException
        ),
        pytest.param(
            "1.2.3.4.2.3.4.5",
            InvalidKafkaBrokersException
        )
    ]
)
def test_invalid_kafka_brokers_value(value, expected):
    with pytest.raises(expected):
        KafkaBrokersSchema.verify_regex(value)


@pytest.mark.parametrize(
    "value,expected", [
        pytest.param(
            "!@#$%^&*()",
            InvalidKafkaTopicException
        ),
        pytest.param(
            "is_this_a_topic?",
            InvalidKafkaTopicException
        )
    ]
)
def test_invalid_kafka_topic_value(value, expected):
    with pytest.raises(expected):
        KafkaTopicSchema.verify_regex(value)
