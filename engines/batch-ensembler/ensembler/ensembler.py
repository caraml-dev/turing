import mlflow
from pyspark.sql import SparkSession, DataFrame, types
from pyspark.sql.functions import struct
from .api.proto.v1 import batch_ensembling_job_pb2 as pb2


class Ensembler:
    _PRIMITIVE_TYPE_MAP = {
        pb2.Ensembler.ResultType.DOUBLE: types.DoubleType(),
        pb2.Ensembler.ResultType.FLOAT: types.FloatType(),
        pb2.Ensembler.ResultType.INTEGER: types.LongType(),
        pb2.Ensembler.ResultType.LONG: types.LongType(),
        pb2.Ensembler.ResultType.STRING: types.StringType(),
    }

    _CAST_TYPE_MAP = {
        pb2.Ensembler.ResultType.INTEGER: types.IntegerType(),
    }

    def __init__(self,
                 ensembler_uri: str,
                 result_column_name: str,
                 result_type: types.DataType,
                 cast_type: types.DataType):
        self._ensembler_uri = ensembler_uri
        self._result_column_name = result_column_name
        self._result_type = result_type
        self._cast_type = cast_type

    def ensemble(self, combined_df: DataFrame, spark: SparkSession) -> DataFrame:
        udf = mlflow.pyfunc.spark_udf(
            spark,
            self._ensembler_uri,
            self._result_type
        )

        return combined_df.withColumn(
            self._result_column_name,
            udf(struct(combined_df.columns))
            if self._cast_type is None
            else udf(struct(combined_df.columns)).cast(self._cast_type)
        )

    @classmethod
    def from_config(cls, config: pb2.Ensembler) -> 'Ensembler':
        result_type = None
        cast_type = None
        if config.result.type == pb2.Ensembler.ResultType.ARRAY:
            if config.result.item_type in cls._PRIMITIVE_TYPE_MAP:
                result_type = types.ArrayType(
                    cls._PRIMITIVE_TYPE_MAP.get(config.result.item_type)
                )
                if config.result.item_type in cls._CAST_TYPE_MAP:
                    cast_type = types.ArrayType(
                        cls._CAST_TYPE_MAP.get(config.result.item_type)
                    )
            else:
                raise ValueError(f'unknown item type for array: {config.result.item_type}')
        else:
            result_type = cls._PRIMITIVE_TYPE_MAP.get(config.result.type)
            cast_type = cls._CAST_TYPE_MAP.get(config.result.type)

        if result_type is None:
            raise ValueError(f'unknown result type: {config.result.type}')

        return Ensembler(config.uri, config.result.column_name, result_type, cast_type)
