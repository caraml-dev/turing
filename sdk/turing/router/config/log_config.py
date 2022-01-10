import re
import turing.generated.models
from enum import Enum
from typing import Optional
from turing.generated.model_utils import OpenApiModel


class ResultLoggerType(Enum):
    NOP = "nop"
    CONSOLE = "console"
    BIGQUERY = "bigquery"
    KAFKA = "kafka"

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.ResultLoggerType(self.value)


class LogConfig:
    def __init__(self,
                 result_logger_type: ResultLoggerType,
                 bigquery_config: turing.generated.models.BigQueryConfig = None,
                 kafka_config: turing.generated.models.KafkaConfig = None):
        self._result_logger_type = result_logger_type
        self._bigquery_config = bigquery_config
        self._kafka_config = kafka_config

    @property
    def result_logger_type(self) -> ResultLoggerType:
        return self._result_logger_type

    @property
    def bigquery_config(self) -> turing.generated.models.BigQueryConfig:
        return self._bigquery_config

    @property
    def kafka_config(self) -> turing.generated.models.KafkaConfig:
        return self._kafka_config

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.RouterConfigConfigLogConfig(
            result_logger_type=self.result_logger_type.to_open_api(),
            bigquery_config=self.bigquery_config,
            kafka_config=self.kafka_config
        )


class BigQueryLogConfig(LogConfig):
    def __init__(self,
                 table: str,
                 service_account_secret: str,
                 batch_load: bool = None):
        BigQueryLogConfig._verify_table(table)
        self._table = table
        self._service_account_secret = service_account_secret
        self._batch_load = batch_load

        super().__init__(result_logger_type=ResultLoggerType.BIGQUERY)

    @property
    def table(self) -> str:
        return self._table

    @table.setter
    def table(self, table):
        BigQueryLogConfig._verify_table(table)
        self._table = table

    @property
    def service_account_secret(self) -> str:
        return self._service_account_secret

    @property
    def batch_load(self) -> Optional[bool]:
        return self._batch_load

    @classmethod
    def _verify_table(cls, table):
        matched = re.fullmatch(r"^[a-z][a-z0-9-]+\.\w+([_]?\w)+\.\w+([_]?\w)+$", table, re.IGNORECASE)
        if bool(matched) is False:
            raise InvalidBigQueryTableException(
                f"Valid BQ table name is required, e.g. project_name.dataset.table; "
                f"table passed: {table}"
            )

    def to_open_api(self) -> OpenApiModel:
        self._bigquery_config = turing.generated.models.BigQueryConfig(
            table=self.table,
            service_account_secret=self.service_account_secret,
            batch_load=self.batch_load
        )
        return super().to_open_api()


class KafkaConfigSerializationFormat(Enum):
    JSON = "json"
    PROTOBUF = "protobuf"


class KafkaLogConfig(LogConfig):
    def __init__(self,
                 brokers: str,
                 topic: str,
                 serialization_format: KafkaConfigSerializationFormat):
        KafkaLogConfig._verify_brokers(brokers)
        KafkaLogConfig._verify_topic(topic)
        self._brokers = brokers
        self._topic = topic
        self._serialization_format = serialization_format

        super().__init__(result_logger_type=ResultLoggerType.KAFKA)

    @property
    def brokers(self) -> str:
        return self._brokers

    @brokers.setter
    def brokers(self, brokers):
        KafkaLogConfig._verify_brokers(brokers)
        self._brokers = brokers

    @property
    def topic(self) -> str:
        return self._topic

    @topic.setter
    def topic(self, topic):
        KafkaLogConfig._verify_topic(topic)
        self._topic = topic

    @property
    def serialization_format(self) -> KafkaConfigSerializationFormat:
        return self._serialization_format

    @classmethod
    def _verify_brokers(cls, brokers):
        matched = re.fullmatch(
            r"^([a-z]+:\/\/)?\[?([0-9a-zA-Z\-%._:]*)\]?:([0-9]+)(,([a-z]+:\/\/)?\[?([0-9a-zA-Z\-%._:]*)\]?:([0-9]+))*$",
            brokers,
            re.IGNORECASE
        )
        if bool(matched) is False:
            raise InvalidKafkaBrokersException(
                f"One or more valid Kafka brokers is required, e.g. host1:port1,host2:port2; "
                f"brokers passed: {brokers}"
            )

    @classmethod
    def _verify_topic(cls, topic):
        matched = re.fullmatch(r"^[A-Za-z0-9_.-]{1,249}", topic, re.IGNORECASE)
        if bool(matched) is False:
            raise InvalidKafkaTopicException(
                f"A valid Kafka topic name may only contain letters, numbers, dot, hyphen or underscore; "
                f"topic passed: {topic}"
            )

    def to_open_api(self) -> OpenApiModel:
        KafkaLogConfig._verify_brokers(self._brokers)
        KafkaLogConfig._verify_topic(self._topic)

        self._kafka_config = turing.generated.models.KafkaConfig(
            brokers=self.brokers,
            topic=self.topic,
            serialization_format=self.serialization_format.value
        )
        return super().to_open_api()


class InvalidBigQueryTableException(Exception):
    pass


class InvalidKafkaBrokersException(Exception):
    pass


class InvalidKafkaTopicException(Exception):
    pass
