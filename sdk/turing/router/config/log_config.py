import abc
from enum import Enum
from typing import Iterable, MutableMapping, Optional, Dict, List
import turing.generated.models
from turing._base_types import DataObject
from turing.generated.model_utils import OpenApiModel


class ResultLoggerType(Enum):
    NOP = "nop"
    CONSOLE = "console"
    BIGQUERY = "bigquery"
    KAFKA = "kafka"


class BigQueryConfig:
    def __init__(self,
                 table: str,
                 service_account_secret: str,
                 batch_load: bool):
        self._table = table
        self._service_account_secret = service_account_secret
        self._batch_load = batch_load

    @property
    def table(self) -> str:
        return self._table

    @property
    def service_account_secret(self) -> str:
        return self._service_account_secret

    @property
    def batch_load(self) -> bool:
        return self._batch_load

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.BigQueryConfig(
            table=self.table,
            service_account_secret=self.service_account_secret,
            batch_load=self.batch_load
        )


class KafkaConfig:
    def __init__(self,
                 brokers: str,
                 topic: str,
                 serialization_format: str):
        assert serialization_format in {"json", "protobuf"}
        self._brokers = brokers
        self._topic = topic
        self._serialization_format = serialization_format

    @property
    def brokers(self) -> str:
        return self._brokers

    @property
    def topic(self) -> str:
        return self._topic

    @property
    def serialization_format(self) -> str:
        return self._serialization_format

    def to_open_api(self) -> OpenApiModel:
        assert self.serialization_format in {"json", "protobuf"}
        return turing.generated.models.KafkaConfig(
            brokers=self.brokers,
            topic=self.topic,
            serialization_format=self.serialization_format
        )
