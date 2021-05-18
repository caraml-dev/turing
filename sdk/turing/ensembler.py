import abc
import os
from abc import abstractmethod
from enum import Enum
from typing import Optional, Union, List, Any, Dict
import turing.utils
import mlflow.pyfunc
import numpy
import pandas

import turing.generated.models
from turing._base_types import ApiObject, ApiObjectSpec


class EnsemblerBase(abc.ABC):

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


class PyFunc(EnsemblerBase, mlflow.pyfunc.PythonModel):
    PREDICTION_COLUMN_PREFIX = '__predictions__'

    def load_context(self, context):
        self.initialize(context.artifacts)

    @abstractmethod
    def initialize(self, artifacts: dict):
        """
        Implementation of EnsemblerBase can specify initialization step which
        will be called one time during model initialization.
        :param artifacts: dictionary of artifacts passed to log_component method
        """

    def predict(self, context, model_input: pandas.DataFrame) -> \
            Union[numpy.ndarray, pandas.Series, pandas.DataFrame]:
        prediction_columns = {
            col: col[len(PyFunc.PREDICTION_COLUMN_PREFIX):]
            for col in model_input.columns if col.startswith(PyFunc.PREDICTION_COLUMN_PREFIX)
        }

        return model_input \
            .rename(columns=prediction_columns) \
            .apply(lambda row:
                   self.ensemble(
                       features=row.drop(prediction_columns.values()),
                       predictions=row[prediction_columns.values()],
                       treatment_config=None
                   ), axis=1, result_type='expand')


class EnsemblerType(Enum):
    PYFUNC = "pyfunc"


@turing.utils.autostr
@ApiObjectSpec(turing.generated.models.Ensembler)
class Ensembler(ApiObject):
    def __init__(self, name: str, type: EnsemblerType, **kwargs):
        super().__init__(**kwargs)
        self._name = name
        self._type = type

    @property
    def name(self) -> str:
        return self._name

    @name.setter
    def name(self, name):
        self._name = name

    @property
    def type(self) -> str:
        return self._type.value

    @classmethod
    def list(
            cls,
            type: Optional[EnsemblerType] = None,
            page: Optional[int] = None,
            page_size: Optional[int] = None):
        from turing import active_session

        active_session.create_ensembler


@ApiObjectSpec(turing.generated.models.PyFuncEnsembler)
class PyFuncEnsembler(Ensembler):
    DEFAULT_ENSEMBLER_PATH = "ensembler"

    def __init__(
            self,
            mlflow_experiment_id: int,
            mlflow_run_id: str,
            artifact_uri: str,
            **kwargs):
        kwargs.pop('type', None)
        super(PyFuncEnsembler, self).__init__(type=EnsemblerType.PYFUNC, **kwargs)
        self._mlflow_experiment_id = mlflow_experiment_id
        self._mlflow_run_id = mlflow_run_id
        self._artifact_uri = artifact_uri

    @property
    def mlflow_experiment_id(self) -> int:
        return self._mlflow_experiment_id

    @mlflow_experiment_id.setter
    def mlflow_experiment_id(self, mlflow_experiment_id: int):
        self._mlflow_experiment_id = mlflow_experiment_id

    @property
    def mlflow_run_id(self) -> str:
        return self._mlflow_run_id

    @mlflow_run_id.setter
    def mlflow_run_id(self, mlflow_run_id):
        self._mlflow_run_id = mlflow_run_id

    @property
    def artifact_uri(self) -> str:
        return self._artifact_uri

    @artifact_uri.setter
    def artifact_uri(self, artifact_uri):
        self._artifact_uri = artifact_uri

    @classmethod
    def _experiment_name(cls, name: str) -> str:
        pass

    def _save(self) -> 'PyFuncEnsembler':
        from turing import active_session

        return PyFuncEnsembler.from_open_api(
            active_session.update_ensembler(self.to_open_api()))

    def update(
            self,
            name: Optional[str] = None,
            ensembler_instance: Optional[EnsemblerBase] = None,
            conda_env: Optional[Union[str, Dict[str, Any]]] = None,
            code_dir: Optional[List[str]] = None,
            artifacts: Optional[Dict[str, str]] = None) -> 'PyFuncEnsembler':
        import mlflow
        from turing import active_session

        if name:
            self.name = name

        if ensembler_instance:
            project_name = active_session.active_project.name
            mlflow.set_experiment(experiment_name=os.path.join(project_name, "ensemblers", self.name))

            mlflow.start_run()
            mlflow.pyfunc.log_model(
                PyFuncEnsembler.DEFAULT_ENSEMBLER_PATH,
                python_model=ensembler_instance,
                conda_env=conda_env,
                code_path=code_dir,
                artifacts=artifacts,
            )

            run = mlflow.active_run()

            self.mlflow_experiment_id = int(run.info.experiment_id)
            self.mlflow_run_id = run.info.run_id
            self.artifact_uri = mlflow.get_artifact_uri()

            mlflow.end_run()

        return self._save()

    @classmethod
    def create(
            cls,
            name: str,
            ensembler_instance: EnsemblerBase,
            conda_env: Union[str, Dict[str, Any]],
            code_dir: List[str] = None,
            artifacts: Dict[str, str] = None,
            ) -> 'PyFuncEnsembler':
        import mlflow
        from turing import active_session

        project_name = active_session.active_project.name
        mlflow.set_experiment(experiment_name=os.path.join(project_name, "ensemblers", name))

        mlflow.start_run()
        mlflow.pyfunc.log_model(
            PyFuncEnsembler.DEFAULT_ENSEMBLER_PATH,
            python_model=ensembler_instance,
            conda_env=conda_env,
            code_path=code_dir,
            artifacts=artifacts,
        )

        run = mlflow.active_run()

        ensembler = PyFuncEnsembler(
            name=name,
            mlflow_experiment_id=int(run.info.experiment_id),
            mlflow_run_id=run.info.run_id,
            artifact_uri=mlflow.get_artifact_uri()
        )
        mlflow.end_run()

        return PyFuncEnsembler.from_open_api(
            active_session.create_ensembler(ensembler.to_open_api()))



