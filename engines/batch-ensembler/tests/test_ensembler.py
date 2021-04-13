from abc import ABC
from typing import Optional, Any, List
import pandas
import pytest
from chispa.dataframe_comparer import assert_df_equality
import ensembler.components.experimentation as experimentation
from ensembler.ensembler import Ensembler
from ensembler.api.proto.v1 import batch_ensembling_job_pb2 as pb2
from pyspark.sql import types as T
from tests.utils.proto_utils import from_yaml


@pytest.fixture()
def input_df(spark):
    return spark.createDataFrame(
        [
            (1, "customer_1", 0.01, 0.02),
            (2, "customer_2", 0.02, 0.04),
            (3, "customer_3", 0.03, 0.06),
        ],
        f"customer_id int,"
        f"name string,"
        f"{experimentation.PREDICTION_COLUMN_PREFIX}model_a double,"
        f"{experimentation.PREDICTION_COLUMN_PREFIX}model_b double"
    )


class TestEnsembler(experimentation.Ensembler, ABC):

    def ensemble(
            self,
            features: pandas.Series,
            predictions: pandas.Series,
            treatment_config: Optional[dict]) -> Any:
        customer_id = features['customer_id']
        if customer_id % 2 == 0:
            result = predictions['model_a']
        else:
            result = predictions['model_b']
        return str(result)


@pytest.fixture()
def config_simple():
    import os
    import mlflow
    from mlflow.pyfunc import log_model
    log_model(
        artifact_path="ensembler",
        python_model=TestEnsembler(),
        code_path=[os.path.join(os.path.dirname(__file__), "../ensembler")])

    ensembler_path = os.path.join(mlflow.get_artifact_uri(), "ensembler")

    yield from_yaml(f"""\
    uri: {ensembler_path}
    result: 
        columnName: test_results
        type: STRING
    """, pb2.Ensembler())


class ArrayEnsembler(experimentation.Ensembler, ABC):

    def ensemble(
            self,
            features: pandas.Series,
            predictions: pandas.Series,
            treatment_config: Optional[dict]) -> Any:
        return predictions.to_numpy() * 2


@pytest.fixture()
def config_array():
    import os
    import mlflow
    from mlflow.pyfunc import log_model
    log_model(
        artifact_path="ensembler_v2",
        python_model=ArrayEnsembler(),
        code_path=[os.path.join(os.path.dirname(__file__), "../ensembler")])

    ensembler_path = os.path.join(mlflow.get_artifact_uri(), "ensembler_v2")

    yield from_yaml(f"""\
    uri: {ensembler_path}
    result: 
        columnName: test_results
        type: ARRAY
        itemType: DOUBLE
    """, pb2.Ensembler())


def test_ensemble_simple(spark, input_df, config_simple: pb2.Ensembler):
    ensembler = Ensembler.from_config(config_simple)

    result_df = ensembler.ensemble(input_df, spark)
    expected_df = spark.createDataFrame(
        [
            (1, "customer_1", 0.01, 0.02, '0.02'),
            (2, "customer_2", 0.02, 0.04, '0.02'),
            (3, "customer_3", 0.03, 0.06, '0.06'),
        ],
        f"customer_id int,"
        f"name string,"
        f"{experimentation.PREDICTION_COLUMN_PREFIX}model_a double,"
        f"{experimentation.PREDICTION_COLUMN_PREFIX}model_b double,"
        f"{config_simple.result.column_name} string"
    )
    assert_df_equality(result_df, expected_df)


def test_ensemble_array(spark, input_df, config_array: pb2.Ensembler):
    ensembler = Ensembler.from_config(config_array)
    result_df = ensembler.ensemble(input_df, spark)

    expected_df = spark.createDataFrame(
        [
            (1, "customer_1", 0.01, 0.02, [0.02, 0.04]),
            (2, "customer_2", 0.02, 0.04, [0.04, 0.08]),
            (3, "customer_3", 0.03, 0.06, [0.06, 0.12]),
        ],
        T.StructType(
            [
                T.StructField('customer_id', T.IntegerType()),
                T.StructField('name', T.StringType()),
                T.StructField(f'{experimentation.PREDICTION_COLUMN_PREFIX}model_a', T.DoubleType()),
                T.StructField(f'{experimentation.PREDICTION_COLUMN_PREFIX}model_b', T.DoubleType()),
                T.StructField(config_array.result.column_name, T.ArrayType(T.DoubleType()))
            ]
        )
    )
    assert_df_equality(result_df, expected_df)
