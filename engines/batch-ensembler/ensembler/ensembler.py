import mlflow
import ensembler.api.proto.v1.batch_ensembling_job_pb2 as pb2
from pyspark.sql import SparkSession, DataFrame, types
from pyspark.sql.functions import struct


class Ensembler(object):
    _PRIMITIVE_TYPE_MAP = {
        pb2.Ensembler.ResultType.DOUBLE: types.DoubleType(),
        pb2.Ensembler.ResultType.FLOAT: types.FloatType(),
        pb2.Ensembler.ResultType.INTEGER: types.IntegerType(),
        pb2.Ensembler.ResultType.LONG: types.LongType(),
        pb2.Ensembler.ResultType.STRING: types.StringType(),
    }

    def __init__(self, ensembler_uri: str, result_column_name: str, result_type: types.DataType):
        self._ensembler_uri = ensembler_uri
        self._result_column_name = result_column_name
        self._result_type = result_type

    def ensemble(self, combined_df: DataFrame, spark: SparkSession) -> DataFrame:
        udf = mlflow.pyfunc.spark_udf(
            spark,
            self._ensembler_uri,
            self._result_type
        )

        return combined_df.withColumn(
            self._result_column_name,
            udf(struct(combined_df.columns))
        )

    @classmethod
    def from_config(cls, config: pb2.Ensembler) -> 'Ensembler':
        if config.result.type == pb2.Ensembler.ResultType.ARRAY:
            it = config.result.item_type
            if it in cls._PRIMITIVE_TYPE_MAP:
                rt = types.ArrayType(cls._PRIMITIVE_TYPE_MAP.get(it))
            else:
                raise ValueError(f'unknown item type for array: {it}')
        elif config.result.type in cls._PRIMITIVE_TYPE_MAP:
            rt = cls._PRIMITIVE_TYPE_MAP.get(config.result.type)
        else:
            raise ValueError(f'unknown result type: {config.result.type}')

        return Ensembler(config.uri, config.result.column_name, rt)
