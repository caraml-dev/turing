import turing.generated.models
import turing.router.config.log_config
import turing.router.config.common.schemas
import pytest


@pytest.mark.parametrize(
    "result_logger_type,expected", [
        pytest.param(
            "nop",
            turing.generated.models.ResultLoggerType("nop")
        ),
        pytest.param(
            "console",
            turing.generated.models.ResultLoggerType("console")
        ),
        pytest.param(
            "bigquery",
            turing.generated.models.ResultLoggerType("bigquery")
        ),
        pytest.param(
            "kafka",
            turing.generated.models.ResultLoggerType("kafka")
        )
    ])
def test_create_result_logger_type(result_logger_type, expected):
    actual = turing.router.config.log_config.ResultLoggerType(result_logger_type).to_open_api()
    assert actual == expected


@pytest.mark.parametrize(
    "result_logger_type,bigquery_config,kafka_config,expected", [
        pytest.param(
            turing.router.config.log_config.ResultLoggerType.NOP,
            None,
            None,
            turing.generated.models.RouterConfigConfigLogConfig(
                result_logger_type=turing.generated.models.ResultLoggerType('nop'),
            )
        ),
        pytest.param(
            turing.router.config.log_config.ResultLoggerType.BIGQUERY,
            turing.generated.models.BigQueryConfig(
                table="bigqueryproject.bigquerydataset.bigquerytable",
                service_account_secret="my-little-secret"
            ),
            None,
            turing.generated.models.RouterConfigConfigLogConfig(
                result_logger_type=turing.generated.models.ResultLoggerType('bigquery'),
                bigquery_config=turing.generated.models.BigQueryConfig(
                    table="bigqueryproject.bigquerydataset.bigquerytable",
                    service_account_secret="my-little-secret"
                ),
            )
        ),
        pytest.param(
            turing.router.config.log_config.ResultLoggerType.KAFKA,
            None,
            turing.generated.models.KafkaConfig(
                brokers="1.2.3.4:5678,9.0.1.2:3456",
                topic="new_topics",
                serialization_format="json"
            ),
            turing.generated.models.RouterConfigConfigLogConfig(
                result_logger_type=turing.generated.models.ResultLoggerType('kafka'),
                kafka_config=turing.generated.models.KafkaConfig(
                    brokers="1.2.3.4:5678,9.0.1.2:3456",
                    topic="new_topics",
                    serialization_format="json"
                ),
            )
        )
    ])
def test_create_log_config_with_valid_params(
        result_logger_type,
        bigquery_config,
        kafka_config,
        expected
):
    actual = turing.router.config.log_config.LogConfig(
        result_logger_type,
        bigquery_config,
        kafka_config,
    ).to_open_api()
    assert actual == expected


@pytest.mark.parametrize(
    "table,service_account_secret,batch_load,expected", [
        pytest.param(
            "bigqueryproject.bigquerydataset.bigquerytable",
            "my-little-secret",
            None,
            turing.generated.models.RouterConfigConfigLogConfig(
                result_logger_type=turing.generated.models.ResultLoggerType('bigquery'),
                bigquery_config=turing.generated.models.BigQueryConfig(
                    table="bigqueryproject.bigquerydataset.bigquerytable",
                    service_account_secret="my-little-secret",
                    batch_load=None
                ),
            )
        )
    ])
def test_create_bigquery_log_config_with_valid_params(table, service_account_secret, batch_load, expected):
    actual = turing.router.config.log_config.BigQueryLogConfig(
        table=table,
        service_account_secret=service_account_secret,
        batch_load=batch_load
    ).to_open_api()
    assert actual == expected


@pytest.mark.parametrize(
    "table,service_account_secret,batch_load,expected", [
        pytest.param(
            "bigqueryprojectownsbigquerydatasetownsbigquerytable",
            "my-little-secret",
            None,
            turing.router.config.common.schemas.InvalidBigQueryTableException
        )
    ])
def test_create_bigquery_log_config_with_invalid_table(table, service_account_secret, batch_load, expected):
    with pytest.raises(expected):
        turing.router.config.log_config.BigQueryLogConfig(
            table=table,
            service_account_secret=service_account_secret,
            batch_load=batch_load
        ).to_open_api()


@pytest.mark.parametrize(
    "new_table,table,service_account_secret,batch_load,expected", [
        pytest.param(
            "bigqueryproject.bigquerydataset.bigquerytable",
            "bigproject.bigdataset.bigtable",
            "my-little-secret",
            None,
            turing.generated.models.RouterConfigConfigLogConfig(
                result_logger_type=turing.generated.models.ResultLoggerType('bigquery'),
                bigquery_config=turing.generated.models.BigQueryConfig(
                    table="bigqueryproject.bigquerydataset.bigquerytable",
                    service_account_secret="my-little-secret",
                    batch_load=None
                ),
            )
        )
    ])
def test_set_bigquery_log_config_with_valid_table(new_table, table, service_account_secret, batch_load, expected):
    actual = turing.router.config.log_config.BigQueryLogConfig(
        table=table,
        service_account_secret=service_account_secret,
        batch_load=batch_load
    )
    actual.table = new_table
    assert actual.to_open_api() == expected


@pytest.mark.parametrize(
    "new_table,table,service_account_secret,batch_load,expected", [
        pytest.param(
            "bigqueryprojectownsbigquerydatasetownsbigquerytable",
            "bigqueryproject.bigquerydataset.bigquerytable",
            "my-little-secret",
            None,
            turing.router.config.common.schemas.InvalidBigQueryTableException
        )
    ])
def test_set_bigquery_log_config_with_invalid_table(new_table, table, service_account_secret, batch_load, expected):
    actual = turing.router.config.log_config.BigQueryLogConfig(
        table=table,
        service_account_secret=service_account_secret,
        batch_load=batch_load
    )
    with pytest.raises(expected):
        actual.table = new_table


@pytest.mark.parametrize(
    "brokers,topic,serialization_format,expected", [
        pytest.param(
            "1.2.3.4:5678,9.0.1.2:3456",
            "new_topics",
            turing.router.config.log_config.KafkaConfigSerializationFormat.JSON,
            turing.generated.models.RouterConfigConfigLogConfig(
                result_logger_type=turing.generated.models.ResultLoggerType('kafka'),
                kafka_config=turing.generated.models.KafkaConfig(
                    brokers="1.2.3.4:5678,9.0.1.2:3456",
                    topic="new_topics",
                    serialization_format="json"
                ),
            )
        ),
        pytest.param(
            "1.2.3.4:5678,9.0.1.2:3456",
            "new_topics",
            turing.router.config.log_config.KafkaConfigSerializationFormat.PROTOBUF,
            turing.generated.models.RouterConfigConfigLogConfig(
                result_logger_type=turing.generated.models.ResultLoggerType('kafka'),
                kafka_config=turing.generated.models.KafkaConfig(
                    brokers="1.2.3.4:5678,9.0.1.2:3456",
                    topic="new_topics",
                    serialization_format="protobuf"
                ),
            )
        )
    ])
def test_create_kafka_log_config_with_valid_params(brokers, topic, serialization_format, expected):
    actual = turing.router.config.log_config.KafkaLogConfig(
        brokers=brokers,
        topic=topic,
        serialization_format=serialization_format
    ).to_open_api()
    assert actual == expected


@pytest.mark.parametrize(
    "brokers,topic,serialization_format,expected", [
        pytest.param(
            "1.2.3.4:5.6.7.8,9.0.1.2:3.4.5.6",
            "new_topics",
            turing.router.config.log_config.KafkaConfigSerializationFormat.JSON,
            turing.router.config.common.schemas.InvalidKafkaBrokersException
        ),
        pytest.param(
            "1.2.3.4:5.6.7.8,9.0.1.2:3.4.5.6",
            "new_topics",
            turing.router.config.log_config.KafkaConfigSerializationFormat.PROTOBUF,
            turing.router.config.common.schemas.InvalidKafkaBrokersException
        )
    ])
def test_create_kafka_log_config_with_invalid_brokers(brokers, topic, serialization_format, expected):
    with pytest.raises(expected):
        turing.router.config.log_config.KafkaLogConfig(
            brokers=brokers,
            topic=topic,
            serialization_format=serialization_format
        )


@pytest.mark.parametrize(
    "new_brokers,brokers,topic,serialization_format,expected", [
        pytest.param(
            "1.2.3.4:5678,9.0.1.2:3456",
            "9.8.7.6:5432,1.0.9.8:7654",
            "new_topics",
            turing.router.config.log_config.KafkaConfigSerializationFormat.JSON,
            turing.generated.models.RouterConfigConfigLogConfig(
                result_logger_type=turing.generated.models.ResultLoggerType('kafka'),
                kafka_config=turing.generated.models.KafkaConfig(
                    brokers="1.2.3.4:5678,9.0.1.2:3456",
                    topic="new_topics",
                    serialization_format="json"
                ),
            )
        ),
        pytest.param(
            "1.2.3.4:5678,9.0.1.2:3456",
            "9.8.7.6:5432,1.0.9.8:7654",
            "new_topics",
            turing.router.config.log_config.KafkaConfigSerializationFormat.PROTOBUF,
            turing.generated.models.RouterConfigConfigLogConfig(
                result_logger_type=turing.generated.models.ResultLoggerType('kafka'),
                kafka_config=turing.generated.models.KafkaConfig(
                    brokers="1.2.3.4:5678,9.0.1.2:3456",
                    topic="new_topics",
                    serialization_format="protobuf"
                ),
            )
        )
    ])
def test_set_kafka_log_config_with_valid_brokers(new_brokers, brokers, topic, serialization_format, expected):
    actual = turing.router.config.log_config.KafkaLogConfig(
        brokers=brokers,
        topic=topic,
        serialization_format=serialization_format
    )
    actual.brokers = new_brokers
    assert actual.to_open_api() == expected


@pytest.mark.parametrize(
    "new_brokers,brokers,topic,serialization_format,expected", [
        pytest.param(
            "1.2.3.4:5.6.7.8,9.0.1.2:3.4.5.6",
            "1.2.3.4:5678,9.0.1.2:3456",
            "new_topics",
            turing.router.config.log_config.KafkaConfigSerializationFormat.JSON,
            turing.router.config.common.schemas.InvalidKafkaBrokersException
        ),
        pytest.param(
            "1.2.3.4:5.6.7.8,9.0.1.2:3.4.5.6",
            "1.2.3.4:5678,9.0.1.2:3456",
            "new_topics",
            turing.router.config.log_config.KafkaConfigSerializationFormat.PROTOBUF,
            turing.router.config.common.schemas.InvalidKafkaBrokersException
        )
    ])
def test_set_kafka_log_config_with_invalid_brokers(new_brokers, brokers, topic, serialization_format, expected):
    actual = turing.router.config.log_config.KafkaLogConfig(
        brokers=brokers,
        topic=topic,
        serialization_format=serialization_format
    )
    with pytest.raises(expected):
        actual.brokers = new_brokers


@pytest.mark.parametrize(
    "brokers,topic,serialization_format,expected", [
        pytest.param(
            "1.2.3.4:5678,9.0.1.2:3456",
            "!@#$%^&*()",
            turing.router.config.log_config.KafkaConfigSerializationFormat.JSON,
            turing.router.config.common.schemas.InvalidKafkaTopicException
        ),
        pytest.param(
            "1.2.3.4:5678,9.0.1.2:3456",
            "!@#$%^&*()",
            turing.router.config.log_config.KafkaConfigSerializationFormat.PROTOBUF,
            turing.router.config.common.schemas.InvalidKafkaTopicException
        )
    ])
def test_create_kafka_log_config_with_invalid_topic(brokers, topic, serialization_format, expected):
    with pytest.raises(expected):
        turing.router.config.log_config.KafkaLogConfig(
            brokers=brokers,
            topic=topic,
            serialization_format=serialization_format
        )


@pytest.mark.parametrize(
    "new_topic,brokers,topic,serialization_format,expected", [
        pytest.param(
            "new_topics",
            "1.2.3.4:5678,9.0.1.2:3456",
            "not_so_new_topics",
            turing.router.config.log_config.KafkaConfigSerializationFormat.JSON,
            turing.generated.models.RouterConfigConfigLogConfig(
                result_logger_type=turing.generated.models.ResultLoggerType('kafka'),
                kafka_config=turing.generated.models.KafkaConfig(
                    brokers="1.2.3.4:5678,9.0.1.2:3456",
                    topic="new_topics",
                    serialization_format="json"
                ),
            )
        ),
        pytest.param(
            "new_topics",
            "1.2.3.4:5678,9.0.1.2:3456",
            "not_so_new_topics",
            turing.router.config.log_config.KafkaConfigSerializationFormat.PROTOBUF,
            turing.generated.models.RouterConfigConfigLogConfig(
                result_logger_type=turing.generated.models.ResultLoggerType('kafka'),
                kafka_config=turing.generated.models.KafkaConfig(
                    brokers="1.2.3.4:5678,9.0.1.2:3456",
                    topic="new_topics",
                    serialization_format="protobuf"
                ),
            )
        )
    ])
def test_set_kafka_log_config_with_valid_topic(new_topic, brokers, topic, serialization_format, expected):
    actual = turing.router.config.log_config.KafkaLogConfig(
        brokers=brokers,
        topic=topic,
        serialization_format=serialization_format
    )
    actual.topic = new_topic
    assert actual.to_open_api() == expected


@pytest.mark.parametrize(
    "new_topic,brokers,topic,serialization_format,expected", [
        pytest.param(
            "!@#$%^&*()",
            "1.2.3.4:5678,9.0.1.2:3456",
            "new_topics",
            turing.router.config.log_config.KafkaConfigSerializationFormat.JSON,
            turing.router.config.common.schemas.InvalidKafkaTopicException
        ),
        pytest.param(
            "!@#$%^&*()",
            "1.2.3.4:5678,9.0.1.2:3456",
            "new_topics",
            turing.router.config.log_config.KafkaConfigSerializationFormat.PROTOBUF,
            turing.router.config.common.schemas.InvalidKafkaTopicException
        )
    ])
def test_set_kafka_log_config_with_invalid_topic(new_topic, brokers, topic, serialization_format, expected):
    actual = turing.router.config.log_config.KafkaLogConfig(
        brokers=brokers,
        topic=topic,
        serialization_format=serialization_format
    )
    with pytest.raises(expected):
        actual.topic = new_topic
