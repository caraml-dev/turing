from abc import ABC
from typing import Optional, Any
import pandas
import pytest
from chispa.dataframe_comparer import assert_df_equality
from ensembler.ensembler import Ensembler
from ensembler.api.proto.v1 import batch_ensembling_job_pb2 as pb2
from pyspark.sql import functions as F
from turing.ensembler import PyFunc

from tests.utils.proto_utils import from_yaml

NUM_CUSTOMERS = 3


@pytest.fixture
def input_df(spark):
    return spark.createDataFrame(
        [
            (
                c_id, f'customer_{c_id}', c_id, c_id * 2, c_id / 100,
                c_id / 100, c_id, c_id, str(c_id)
            ) for c_id in range(1, NUM_CUSTOMERS + 1)
        ],
        f'customer_id int,'
        f'name string,'
        f'{PyFunc.PREDICTION_COLUMN_PREFIX}model_a int,'
        f'{PyFunc.PREDICTION_COLUMN_PREFIX}model_b int,'
        f'{PyFunc.PREDICTION_COLUMN_PREFIX}model_double double,'
        f'{PyFunc.PREDICTION_COLUMN_PREFIX}model_float float,'
        f'{PyFunc.PREDICTION_COLUMN_PREFIX}model_integer int,'
        f'{PyFunc.PREDICTION_COLUMN_PREFIX}model_long long,'
        f'{PyFunc.PREDICTION_COLUMN_PREFIX}model_string string'
    )


@pytest.fixture
def expected_result_df(input_df, request):
    if request.param == pb2.Ensembler.ResultType.ARRAY:
        input_df = input_df.withColumn(
            f'{PyFunc.PREDICTION_COLUMN_PREFIX}model_array',
            F.when(F.lit(True), F.array(
                F.col(f'{PyFunc.PREDICTION_COLUMN_PREFIX}model_a') * 2,
                F.col(f'{PyFunc.PREDICTION_COLUMN_PREFIX}model_b') * 2,
            ))
        )
    return input_df.select(
        'customer_id',
        f'{PyFunc.PREDICTION_COLUMN_PREFIX}model_{pb2.Ensembler.ResultType.Name(request.param).lower()}'
    )


class TestEnsembler(PyFunc, ABC):

    def __init__(self, result_type: pb2.Ensembler.ResultType) -> None:
        super().__init__()
        self.result_type = result_type

    def ensemble(
            self,
            features: pandas.Series,
            predictions: pandas.Series,
            treatment_config: Optional[dict]):
        return predictions[f'model_{pb2.Ensembler.ResultType.Name(self.result_type).lower()}']


@pytest.fixture
def config_simple(request):
    import os
    import mlflow
    from mlflow.pyfunc import log_model
    log_model(
        artifact_path='ensembler',
        python_model=TestEnsembler(result_type=request.param),
        code_path=[os.path.join(os.path.dirname(__file__), '../ensembler')])

    ensembler_path = os.path.join(mlflow.get_artifact_uri(), 'ensembler')

    yield from_yaml(f"""\
    uri: {ensembler_path}
    result: 
        columnName: test_results
        type: {pb2.Ensembler.ResultType.Name(request.param)}
    """, pb2.Ensembler())


class ArrayEnsembler(PyFunc, ABC):

    def ensemble(
            self,
            features: pandas.Series,
            predictions: pandas.Series,
            treatment_config: Optional[dict]) -> Any:
        return predictions[['model_a', 'model_b']].to_numpy() * 2


@pytest.fixture
def config_array():
    import os
    import mlflow
    from mlflow.pyfunc import log_model
    log_model(
        artifact_path='ensembler_v2',
        python_model=ArrayEnsembler(),
        code_path=[os.path.join(os.path.dirname(__file__), '../ensembler')])

    ensembler_path = os.path.join(mlflow.get_artifact_uri(), 'ensembler_v2')

    yield from_yaml(f"""\
    uri: {ensembler_path}
    result: 
        columnName: test_results
        type: ARRAY
        itemType: INTEGER
    """, pb2.Ensembler())


@pytest.mark.parametrize(
    'config_simple,expected_result_df',
    (
            (pb2.Ensembler.ResultType.DOUBLE, pb2.Ensembler.ResultType.DOUBLE),
            (pb2.Ensembler.ResultType.FLOAT, pb2.Ensembler.ResultType.FLOAT),
            (pb2.Ensembler.ResultType.INTEGER, pb2.Ensembler.ResultType.INTEGER),
            (pb2.Ensembler.ResultType.LONG, pb2.Ensembler.ResultType.LONG),
            (pb2.Ensembler.ResultType.STRING, pb2.Ensembler.ResultType.STRING),
    ),
    indirect=True
)
def test_ensemble_simple(spark, input_df, config_simple, expected_result_df):
    ensembler = Ensembler.from_config(config_simple)

    expected_result_df.show()

    result_df = ensembler.ensemble(input_df, spark)

    expected_df = input_df.join(
        expected_result_df.toDF('customer_id', config_simple.result.column_name),
        on='customer_id')
    assert_df_equality(result_df, expected_df, ignore_row_order=True)


@pytest.mark.parametrize(
    'expected_result_df',
    [pb2.Ensembler.ResultType.ARRAY],
    indirect=True
)
def test_ensemble_array(spark, input_df, config_array, expected_result_df):
    ensembler = Ensembler.from_config(config_array)
    result_df = ensembler.ensemble(input_df, spark)

    expected_df = input_df.join(
        expected_result_df.toDF('customer_id', config_array.result.column_name),
        on='customer_id')
    assert_df_equality(result_df, expected_df, ignore_row_order=True)
