from abc import abstractmethod
from typing import Optional, Any, Union
import numpy
import pandas
import mlflow

PREDICTION_COLUMN_PREFIX = '__predictions__'


class Ensembler(mlflow.pyfunc.PythonModel):

    def load_context(self, context):
        self.initialize(context.artifacts)

    @abstractmethod
    def initialize(self, artifacts: dict):
        """
        Implementation of Ensembler can specify initialization step which
        will be called one time during model initialization.
        :param artifacts: dictionary of artifacts passed to log_component method
        """

    def predict(self, context, model_input: pandas.DataFrame) -> \
            Union[numpy.ndarray, pandas.Series, pandas.DataFrame]:
        prediction_columns = {
            col: col[len(PREDICTION_COLUMN_PREFIX):]
            for col in model_input.columns if col.startswith(PREDICTION_COLUMN_PREFIX)
        }

        return model_input \
            .rename(columns=prediction_columns) \
            .apply(lambda row:
                   self.ensemble(
                       features=row.drop(prediction_columns.values()),
                       predictions=row[prediction_columns.values()],
                       treatment_config=None
                   ), axis=1, result_type='expand')

    @abstractmethod
    def ensemble(
            self,
            features: pandas.Series,
            predictions: pandas.Series,
            treatment_config: Optional[dict]) -> Any:
        """
        Ensembler should have an ensemble method, that implements the logic on how to
        ensemble final prediction results from individual model predictions and a treatment
        configuration.

        :param features: pandas.Series, containing a single row with input features
        :param predictions: pandas.Series, containing a single row with all models predictions
                `predictions['model-a']` will contain prediction results from the model-a
        :param treatment_config: dictionary, representing the configuration of a treatment,
                that should be applied to a given record. If the experiment engine is not configured
                for this Batch experiment, then `treatment_config` will be `None`

        :returns ensembling result (one of str, int, float, double or array)
        """
