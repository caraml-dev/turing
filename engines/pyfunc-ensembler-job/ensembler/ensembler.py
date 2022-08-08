import mlflow
from pyspark.sql import SparkSession, DataFrame, types
from pyspark.sql.functions import struct
import turing.batch.config as sdk
import turing.generated.models as openapi


class Ensembler:
    _PRIMITIVE_TYPE_MAP = {
        sdk.ResultType.DOUBLE: types.DoubleType(),
        sdk.ResultType.FLOAT: types.FloatType(),
        sdk.ResultType.INTEGER: types.LongType(),
        sdk.ResultType.LONG: types.LongType(),
        sdk.ResultType.STRING: types.StringType(),
    }

    _CAST_TYPE_MAP = {
        sdk.ResultType.INTEGER: types.IntegerType(),
    }

    def __init__(
        self,
        ensembler_uri: str,
        result_column_name: str,
        result_type: types.DataType,
        cast_type: types.DataType,
    ):
        self._ensembler_uri = ensembler_uri
        self._result_column_name = result_column_name
        self._result_type = result_type
        self._cast_type = cast_type

    def ensemble(self, combined_df: DataFrame, spark: SparkSession) -> DataFrame:
        udf = mlflow.pyfunc.spark_udf(spark, self._ensembler_uri, self._result_type)

        return combined_df.withColumn(
            self._result_column_name,
            udf(struct(combined_df.columns))
            if self._cast_type is None
            else udf(struct(combined_df.columns)).cast(self._cast_type),
        )

    @classmethod
    def from_config(cls, config: openapi.EnsemblingJobEnsemblerSpec) -> "Ensembler":
        spark_result_type, spark_cast_type = None, None

        result_type = sdk.ResultType[config.result.type.value]
        if result_type == sdk.ResultType.ARRAY:
            item_type = sdk.ResultType[config.result.item_type.value]
            if item_type in cls._PRIMITIVE_TYPE_MAP:
                spark_result_type = types.ArrayType(
                    cls._PRIMITIVE_TYPE_MAP.get(item_type)
                )
                if item_type in cls._CAST_TYPE_MAP:
                    spark_cast_type = types.ArrayType(cls._CAST_TYPE_MAP.get(item_type))
            else:
                raise ValueError(f"unknown item type for array: {item_type}")
        else:
            spark_result_type = cls._PRIMITIVE_TYPE_MAP.get(result_type)
            spark_cast_type = cls._CAST_TYPE_MAP.get(result_type)

        if spark_result_type is None:
            raise ValueError(f"unknown result type: {result_type}")

        return Ensembler(
            config.uri, config.result.column_name, spark_result_type, spark_cast_type
        )
