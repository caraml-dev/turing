import re
import turing.generated.models
from enum import Enum
from typing import Optional, Dict, Union
from turing.generated.model_utils import OpenApiModel
from turing.router.config.common.schemas import BigQueryTableSchema, KafkaBrokersSchema, KafkaTopicSchema


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
                 kafka_config: turing.generated.models.KafkaConfig = None,
                 **kwargs):
        """
        Method to create a new LogConfig instance

        :param result_logger_type: logging type
        :param bigquery_config: config file for logging using BigQuery
        :param kafka_config: config file for logging using Kafka
        """
        self.result_logger_type = result_logger_type
        self.bigquery_config = bigquery_config
        self.kafka_config = kafka_config

    @property
    def result_logger_type(self) -> ResultLoggerType:
        return self._result_logger_type

    @result_logger_type.setter
    def result_logger_type(self, result_logger_type: Union[ResultLoggerType, str]):
        if isinstance(result_logger_type, ResultLoggerType):
            self._result_logger_type = result_logger_type
        elif isinstance(result_logger_type, str):
            self._result_logger_type = ResultLoggerType(result_logger_type)
        else:
            self._result_logger_type = result_logger_type

    @property
    def bigquery_config(self) -> turing.generated.models.BigQueryConfig:
        return self._bigquery_config

    @bigquery_config.setter
    def bigquery_config(self, bigquery_config: Union[turing.generated.models.BigQueryConfig, Dict]):
        if isinstance(bigquery_config, turing.generated.models.BigQueryConfig):
            self._bigquery_config = bigquery_config
        elif isinstance(bigquery_config, dict):
            self._bigquery_config = turing.generated.models.BigQueryConfig(**bigquery_config)
        else:
            self._bigquery_config = bigquery_config

    @property
    def kafka_config(self) -> turing.generated.models.KafkaConfig:
        return self._kafka_config

    @kafka_config.setter
    def kafka_config(self, kafka_config: Union[turing.generated.models.KafkaConfig, Dict]):
        if isinstance(kafka_config, turing.generated.models.KafkaConfig):
            self._kafka_config = kafka_config
        elif isinstance(kafka_config, dict):
            self._kafka_config = turing.generated.models.KafkaConfig(**kafka_config)
        else:
            self._kafka_config = kafka_config

    def to_open_api(self) -> OpenApiModel:
        self.verify_result_logger_type_and_config_combination()

        kwargs = {}
        if self.bigquery_config is not None:
            kwargs["bigquery_config"] = self.bigquery_config
        if self.kafka_config is not None:
            kwargs["kafka_config"] = self.kafka_config

        return turing.generated.models.RouterConfigConfigLogConfig(
            result_logger_type=self.result_logger_type.to_open_api(),
            **kwargs
        )

    def verify_result_logger_type_and_config_combination(self):
        if self.result_logger_type == ResultLoggerType.BIGQUERY and self.kafka_config is not None:
            raise InvalidResultLoggerTypeAndConfigCombination(
                f"kafka_config must be set to None when result_logger_type is: {self.result_logger_type}"
            )
        if self.result_logger_type == ResultLoggerType.KAFKA and self.bigquery_config is not None:
            raise InvalidResultLoggerTypeAndConfigCombination(
                f"bigquery_config must be set to None when result_logger_type is: {self.result_logger_type}"
            )


class InvalidResultLoggerTypeAndConfigCombination(Exception):
    pass


class BigQueryLogConfig(LogConfig):
    def __init__(self,
                 table: str,
                 service_account_secret: str,
                 batch_load: bool = None):
        """
        Method to create a new log config with a BigQuery config

        :param table: name of the BigQuery table; if the table does not exist, it will be created automatically
        :param service_account_secret: service account which has both JobUser and DataEditor privileges and write access
        :param batch_load: optional parameter to indicate if batch loading is used
        """
        self.table = table
        self.service_account_secret = service_account_secret
        self.batch_load = batch_load

        super().__init__(result_logger_type=ResultLoggerType.BIGQUERY)

    @property
    def table(self) -> str:
        return self._table

    @table.setter
    def table(self, table: str):
        BigQueryTableSchema.verify_regex(table)
        self._table = table

    @property
    def service_account_secret(self) -> str:
        return self._service_account_secret

    @service_account_secret.setter
    def service_account_secret(self, service_account_secret: str):
        self._service_account_secret = service_account_secret

    @property
    def batch_load(self) -> Optional[bool]:
        return self._batch_load

    @batch_load.setter
    def batch_load(self, batch_load: bool):
        self._batch_load = batch_load

    def to_open_api(self) -> OpenApiModel:
        self.bigquery_config = turing.generated.models.BigQueryConfig(
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
        """
        Method to create a new log config with a Kafka config

        :param brokers: comma-separated list of one or more Kafka brokers
        :param topic: valid Kafka topic name on the server; data will be written to this topic
        :param serialization_format: message serialization format to be used
        """
        self.brokers = brokers
        self.topic = topic
        self.serialization_format = serialization_format

        super().__init__(result_logger_type=ResultLoggerType.KAFKA)

    @property
    def brokers(self) -> str:
        return self._brokers

    @brokers.setter
    def brokers(self, brokers):
        KafkaBrokersSchema.verify_regex(brokers)
        self._brokers = brokers

    @property
    def topic(self) -> str:
        return self._topic

    @topic.setter
    def topic(self, topic):
        KafkaTopicSchema.verify_regex(topic)
        self._topic = topic

    @property
    def serialization_format(self) -> KafkaConfigSerializationFormat:
        return self._serialization_format

    @serialization_format.setter
    def serialization_format(self, serialization_format: KafkaConfigSerializationFormat):
        self._serialization_format = serialization_format

    def to_open_api(self) -> OpenApiModel:
        self.kafka_config = turing.generated.models.KafkaConfig(
            brokers=self.brokers,
            topic=self.topic,
            serialization_format=self.serialization_format.value
        )
        return super().to_open_api()
