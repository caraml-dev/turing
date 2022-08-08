from typing import Optional, Any
import pandas
import pytest
from chispa.dataframe_comparer import assert_df_equality
from ensembler.ensembler import Ensembler
from pyspark.sql import functions as F
from turing.ensembler import PyFunc
import turing.batch.config as sdk
import turing.generated.models as openapi
from tests.utils.openapi_utils import from_yaml

NUM_CUSTOMERS = 3


@pytest.fixture
def input_df(spark):
    return spark.createDataFrame(
        [
            (
                c_id,
                f"customer_{c_id}",
                c_id,
                c_id * 2,
                c_id / 100,
                c_id / 100,
                c_id,
                c_id,
                str(c_id),
            )
            for c_id in range(1, NUM_CUSTOMERS + 1)
        ],
        f"customer_id int,"
        f"name string,"
        f"{PyFunc.PREDICTION_COLUMN_PREFIX}model_a int,"
        f"{PyFunc.PREDICTION_COLUMN_PREFIX}model_b int,"
        f"{PyFunc.PREDICTION_COLUMN_PREFIX}model_double double,"
        f"{PyFunc.PREDICTION_COLUMN_PREFIX}model_float float,"
        f"{PyFunc.PREDICTION_COLUMN_PREFIX}model_integer int,"
        f"{PyFunc.PREDICTION_COLUMN_PREFIX}model_long long,"
        f"{PyFunc.PREDICTION_COLUMN_PREFIX}model_string string",
    )


@pytest.fixture
def expected_result_df(input_df, request):
    if request.param == sdk.ResultType.ARRAY:
        input_df = input_df.withColumn(
            f"{PyFunc.PREDICTION_COLUMN_PREFIX}model_array",
            F.when(
                F.lit(True),
                F.array(
                    F.col(f"{PyFunc.PREDICTION_COLUMN_PREFIX}model_a") * 2,
                    F.col(f"{PyFunc.PREDICTION_COLUMN_PREFIX}model_b") * 2,
                ),
            ),
        )
    return input_df.select(
        "customer_id",
        f"{PyFunc.PREDICTION_COLUMN_PREFIX}model_{request.param.name.lower()}",
    )


class TestEnsembler(PyFunc):
    def __init__(self, result_type: sdk.ResultType) -> None:
        super().__init__()
        self.result_type = result_type

    def initialize(self, artifacts: dict):
        pass

    def ensemble(
        self,
        input: pandas.Series,
        predictions: pandas.Series,
        treatment_config: Optional[dict],
    ):
        return predictions[f"model_{self.result_type.name.lower()}"]


@pytest.fixture
def config_simple(request):
    import os
    import mlflow
    from mlflow.pyfunc import log_model

    log_model(
        artifact_path="ensembler",
        python_model=TestEnsembler(result_type=request.param),
        code_path=[os.path.join(os.path.dirname(__file__), "../ensembler")],
    )

    ensembler_path = os.path.join(mlflow.get_artifact_uri(), "ensembler")

    yield from_yaml(
        f"""\
    uri: {ensembler_path}
    result:
        column_name: test_results
        type: {request.param.name}
    """,
        openapi.EnsemblingJobEnsemblerSpec,
    )


class ArrayEnsembler(PyFunc):
    def initialize(self, artifacts: dict):
        pass

    def ensemble(
        self,
        input: pandas.Series,
        predictions: pandas.Series,
        treatment_config: Optional[dict],
    ) -> Any:
        return predictions[["model_a", "model_b"]].to_numpy() * 2


@pytest.fixture
def config_array():
    import os
    import mlflow
    from mlflow.pyfunc import log_model

    log_model(
        artifact_path="ensembler_v2",
        python_model=ArrayEnsembler(),
        code_path=[os.path.join(os.path.dirname(__file__), "../ensembler")],
    )

    ensembler_path = os.path.join(mlflow.get_artifact_uri(), "ensembler_v2")

    yield from_yaml(
        f"""\
    uri: {ensembler_path}
    result:
        column_name: test_results
        type: ARRAY
        item_type: INTEGER
    """,
        openapi.EnsemblingJobEnsemblerSpec,
    )


@pytest.mark.parametrize(
    "config_simple,expected_result_df",
    (
        (sdk.ResultType.DOUBLE, sdk.ResultType.DOUBLE),
        (sdk.ResultType.FLOAT, sdk.ResultType.FLOAT),
        (sdk.ResultType.INTEGER, sdk.ResultType.INTEGER),
        (sdk.ResultType.LONG, sdk.ResultType.LONG),
        (sdk.ResultType.STRING, sdk.ResultType.STRING),
    ),
    indirect=True,
)
def test_ensemble_simple(spark, input_df, config_simple, expected_result_df):
    ensembler = Ensembler.from_config(config_simple)
    result_df = ensembler.ensemble(input_df, spark)

    expected_df = input_df.join(
        expected_result_df.toDF("customer_id", config_simple.result.column_name),
        on="customer_id",
    )
    assert_df_equality(result_df, expected_df, ignore_row_order=True)


@pytest.mark.parametrize("expected_result_df", [sdk.ResultType.ARRAY], indirect=True)
def test_ensemble_array(spark, input_df, config_array, expected_result_df):
    ensembler = Ensembler.from_config(config_array)
    result_df = ensembler.ensemble(input_df, spark)

    expected_df = input_df.join(
        expected_result_df.toDF("customer_id", config_array.result.column_name),
        on="customer_id",
    )
    assert_df_equality(result_df, expected_df, ignore_row_order=True)
