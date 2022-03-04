# LogConfig

Logging for Turing Routers is done through BigQuery or Kafka, and its configuration is managed by the `LogConfig` 
class. Two helper classes (child classes of `LogConfig`) have been created to assist you in constructing these objects:

```python
@dataclass
class BigQueryLogConfig(LogConfig):
    """
    Class to create a new log config with a BigQuery config

    :param table: name of the BigQuery table; if the table does not exist, it will be created automatically
    :param service_account_secret: service account which has both JobUser and DataEditor privileges and write access
    :param batch_load: optional parameter to indicate if batch loading is used
    """
    def __init__(self,
                 table: str,
                 service_account_secret: str,
                 batch_load: bool = None):
        self.table = table
        self.service_account_secret = service_account_secret
        self.batch_load = batch_load
        
        super().__init__(result_logger_type=ResultLoggerType.BIGQUERY)
```

```python
@dataclass
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
```

If you are using a `KafkaLogConfig`, you would additionally have to define a `serialization_format`, which is of a 
`KafkaConfigSerializationFormat`:

```python
class KafkaConfigSerializationFormat(Enum):
    JSON = "json"
    PROTOBUF = "protobuf"
```

If you do not intend to use any logging, simply create a regular `LogConfig` object with `result_loggger_type` set 
as `ResultLoggerType.NOP`, without defining the other arguments:

```python
log_config = LogConfig(result_logger_type=ResultLoggerType.NOP)
```

While `ResultLoggerType` may take on the `enum` value of `ResultLoggerType.CONSOLE`, its behaviour is 
currently undefined and you will almost certainly experience errors while using it.