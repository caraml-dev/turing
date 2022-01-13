import pytest
import turing.generated.models
from turing.router.config.log_config import (ResultLoggerType, LogConfig, BigQueryLogConfig, KafkaLogConfig,
                                             KafkaConfigSerializationFormat, InvalidResultLoggerTypeAndConfigCombination)
from turing.router.config.common.schemas import (InvalidKafkaBrokersException, InvalidKafkaTopicException,
                                                 InvalidBigQueryTableException)


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
    actual = ResultLoggerType(result_logger_type).to_open_api()
    assert actual == expected


@pytest.mark.parametrize(
    "result_logger_type,bigquery_config,kafka_config,expected", [
        pytest.param(
            ResultLoggerType.NOP,
            None,
            None,
            turing.generated.models.RouterConfigConfigLogConfig(
                result_logger_type=turing.generated.models.ResultLoggerType('nop'),
            )
        ),
        pytest.param(
            ResultLoggerType.BIGQUERY,
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
            ResultLoggerType.KAFKA,
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
    actual = LogConfig(
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
    actual = BigQueryLogConfig(
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
            InvalidBigQueryTableException
        )
    ])
def test_create_bigquery_log_config_with_invalid_table(table, service_account_secret, batch_load, expected):
    with pytest.raises(expected):
        BigQueryLogConfig(
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
    actual = BigQueryLogConfig(
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
            InvalidBigQueryTableException
        )
    ])
def test_set_bigquery_log_config_with_invalid_table(new_table, table, service_account_secret, batch_load, expected):
    actual = BigQueryLogConfig(
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
            KafkaConfigSerializationFormat.JSON,
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
            KafkaConfigSerializationFormat.PROTOBUF,
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
    actual = KafkaLogConfig(
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
            KafkaConfigSerializationFormat.JSON,
            InvalidKafkaBrokersException
        ),
        pytest.param(
            "1.2.3.4:5.6.7.8,9.0.1.2:3.4.5.6",
            "new_topics",
            KafkaConfigSerializationFormat.PROTOBUF,
            InvalidKafkaBrokersException
        )
    ])
def test_create_kafka_log_config_with_invalid_brokers(brokers, topic, serialization_format, expected):
    with pytest.raises(expected):
        KafkaLogConfig(
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
            KafkaConfigSerializationFormat.JSON,
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
            KafkaConfigSerializationFormat.PROTOBUF,
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
    actual = KafkaLogConfig(
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
            KafkaConfigSerializationFormat.JSON,
            InvalidKafkaBrokersException
        ),
        pytest.param(
            "1.2.3.4:5.6.7.8,9.0.1.2:3.4.5.6",
            "1.2.3.4:5678,9.0.1.2:3456",
            "new_topics",
            KafkaConfigSerializationFormat.PROTOBUF,
            InvalidKafkaBrokersException
        )
    ])
def test_set_kafka_log_config_with_invalid_brokers(new_brokers, brokers, topic, serialization_format, expected):
    actual = KafkaLogConfig(
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
            KafkaConfigSerializationFormat.JSON,
            InvalidKafkaTopicException
        ),
        pytest.param(
            "1.2.3.4:5678,9.0.1.2:3456",
            "!@#$%^&*()",
            KafkaConfigSerializationFormat.PROTOBUF,
            InvalidKafkaTopicException
        )
    ])
def test_create_kafka_log_config_with_invalid_topic(brokers, topic, serialization_format, expected):
    with pytest.raises(expected):
        KafkaLogConfig(
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
            KafkaConfigSerializationFormat.JSON,
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
            KafkaConfigSerializationFormat.PROTOBUF,
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
    actual = KafkaLogConfig(
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
            KafkaConfigSerializationFormat.JSON,
            InvalidKafkaTopicException
        ),
        pytest.param(
            "!@#$%^&*()",
            "1.2.3.4:5678,9.0.1.2:3456",
            "new_topics",
            KafkaConfigSerializationFormat.PROTOBUF,
            InvalidKafkaTopicException
        )
    ])
def test_set_kafka_log_config_with_invalid_topic(new_topic, brokers, topic, serialization_format, expected):
    actual = KafkaLogConfig(
        brokers=brokers,
        topic=topic,
        serialization_format=serialization_format
    )
    with pytest.raises(expected):
        actual.topic = new_topic


@pytest.mark.parametrize(
    "result_logger_type,bigquery_config,kafka_config,expected", [
        pytest.param(
            ResultLoggerType.BIGQUERY,
            None,
            turing.generated.models.KafkaConfig(
                brokers="1.2.3.4:5678,9.0.1.2:3456",
                topic="new_topics",
                serialization_format="json"
            ),
            InvalidResultLoggerTypeAndConfigCombination
        ),
        pytest.param(
            ResultLoggerType.KAFKA,
            turing.generated.models.BigQueryConfig(
                table="bigqueryproject.bigquerydataset.bigquerytable",
                service_account_secret="my-little-secret"
            ),
            None,
            InvalidResultLoggerTypeAndConfigCombination
        )
    ])
def test_create_log_config_with_conflicting_logger_type_and_config(
        result_logger_type,
        bigquery_config,
        kafka_config,
        expected
):
    with pytest.raises(expected):
        LogConfig(
            result_logger_type,
            bigquery_config,
            kafka_config,
        ).to_open_api()
