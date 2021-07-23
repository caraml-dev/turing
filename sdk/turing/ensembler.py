import abc
from typing import Optional, Union, List, Any, Dict
import mlflow
import numpy
import pandas
import turing.generated.models
from turing.generated.models import EnsemblerType
from turing._base_types import ApiObject, ApiObjectSpec
from turing.batch import EnsemblingJob
from turing.batch.config import EnsemblingJobConfig


class EnsemblerBase(abc.ABC):

    @abc.abstractmethod
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
        pass


class PyFunc(EnsemblerBase, mlflow.pyfunc.PythonModel, abc.ABC):
    """
    Abstract implementation of PyFunc Ensembler.
    It leverages the contract of mlflow's PythonModel and implements its `predict` method.
    """

    PREDICTION_COLUMN_PREFIX = '__predictions__'

    def load_context(self, context):
        self.initialize(context.artifacts)

    @abc.abstractmethod
    def initialize(self, artifacts: dict):
        """
        Implementation of EnsemblerBase can specify initialization step which
        will be called one time during model initialization.
        :param artifacts: dictionary of artifacts passed to log_component method
        """
        pass

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


@ApiObjectSpec(turing.generated.models.Ensembler)
class Ensembler(ApiObject):
    """
    API entity for Ensembler
    """

    def __init__(self, name: str, type: EnsemblerType, project_id: int = None, **kwargs):
        super(Ensembler, self).__init__(**kwargs)
        self._project_id = project_id
        self._name = name
        self._type = type

    @property
    def name(self) -> str:
        return self._name

    @property
    def type(self) -> 'EnsemblerType':
        return self._type

    @property
    def project_id(self) -> int:
        return self._project_id

    @classmethod
    def list(
            cls,
            ensembler_type: Optional[EnsemblerType] = None,
            page: Optional[int] = None,
            page_size: Optional[int] = None) -> List['Ensembler']:
        """
        List ensemblers saved in the active project

        :param ensembler_type: (optional) filter ensemblers by type
        :param page: (optional) pagination parameters – page number
        :param page_size: (optional) pagination parameters - page size

        :return: list of ensemblers
        """

        response = turing.active_session.list_ensemblers(
            ensembler_type=ensembler_type,
            page=page,
            page_size=page_size
        )
        return [Ensembler.from_open_api(item) for item in response.results]


@ApiObjectSpec(turing.generated.models.PyFuncEnsembler)
class PyFuncEnsembler(Ensembler):
    """
    API entity for PyFuncEnsembler
    """

    TYPE = EnsemblerType("pyfunc")
    DEFAULT_ENSEMBLER_PATH = "ensembler"

    def __init__(
            self,
            mlflow_experiment_id: int = None,
            mlflow_run_id: str = None,
            artifact_uri: str = None,
            **kwargs):
        kwargs.pop('type', None)
        super(PyFuncEnsembler, self).__init__(type=PyFuncEnsembler.TYPE, **kwargs)
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
    def _experiment_name(cls, project_name: str, ensembler_name: str) -> str:
        return f"{project_name}/ensemblers/{ensembler_name}"

    def _save(self):
        self.__dict__.update(
            PyFuncEnsembler.from_open_api(
                turing.active_session.update_ensembler(self.to_open_api())
            ).__dict__
        )

    def update(
            self,
            name: Optional[str] = None,
            ensembler_instance: Optional[EnsemblerBase] = None,
            conda_env: Optional[Union[str, Dict[str, Any]]] = None,
            code_dir: Optional[List[str]] = None,
            artifacts: Optional[Dict[str, str]] = None):
        """
        Update existing pyfunc ensembler. Ensembler's data will be updated in-place

        :param name: (optional) new name
        :param ensembler_instance: (optional) updated implementation of the ensembler
        :param conda_env: (optional) either a dictionary representation of a Conda
            environment or the path to a Conda environment yaml file
        :param code_dir: (optional) a list of local filesystem paths to Python file dependencies
            (or directories containing file dependencies). These files are prepended to the
            system path before the ensembler is loaded
        :param artifacts: (optional) dictionary of artifact that will be stored together
            with the model. This will be passed to turing.ensembler.PyFunc.initialize().
            Example: {"config" : "config/staging.yaml"}
        """
        if name:
            self._name = name

        if ensembler_instance:
            project_name = turing.active_session.active_project.name
            mlflow.set_experiment(experiment_name=self._experiment_name(project_name, self.name))

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

        self._save()

    def submit_job(self, job_config: EnsemblingJobConfig) -> 'EnsemblingJob':
        """
        Submit a batch ensembling jobs

        :param job_config: ensembling job configuration

        :return: instance of a submitted ensembling job
        """
        return EnsemblingJob.submit(self.id, job_config)

    @classmethod
    def get_by_id(cls, ensembler_id: int) -> 'PyFuncEnsembler':
        """
        Get the instance of a pyfunc ensembler with given ID

        :param ensembler_id:
        :return: instance of pyfunc ensembler
        """
        return PyFuncEnsembler.from_open_api(
            turing.active_session.get_ensembler(ensembler_id))

    @classmethod
    def list(
            cls,
            page: Optional[int] = None,
            page_size: Optional[int] = None,
            **kwargs) -> List['PyFuncEnsembler']:
        """
        List pyfunc ensemblers saved in the active project

        :param page: (optional) pagination parameters – page number
        :param page_size: (optional) pagination parameters - page size

        :return: list of pyfunc ensemblers
        """
        response = turing.active_session.list_ensemblers(
            ensembler_type=PyFuncEnsembler.TYPE,
            page=page,
            page_size=page_size
        )
        return [PyFuncEnsembler.from_open_api(item) for item in response.results]

    @classmethod
    def create(
            cls,
            name: str,
            ensembler_instance: PyFunc,
            conda_env: Union[str, Dict[str, Any]],
            code_dir: Optional[List[str]] = None,
            artifacts: Dict[str, str] = None,
            ) -> 'PyFuncEnsembler':
        """
        Save new pyfunc ensembler in the active project

        :param name: ensembler's name. Must be unique among all ensemblers within the project
        :param ensembler_instance: implementation of PyFunc ensembler
        :param conda_env: either a dictionary representation of a Conda environment or
            the path to a Conda environment yaml file
        :param code_dir: (optional) a list of local filesystem paths to Python file dependencies
            (or directories containing file dependencies). These files are prepended to the
            system path before the ensembler is loaded
        :param artifacts: dictionary of artifact that will be stored together with the model.
            This will be passed to turing.ensembler.PyFunc.initialize().
            Example: {"config" : "config/staging.yaml"}

        :return: saved instance of PyFuncEnsembler
        """
        project_name = turing.active_session.active_project.name
        mlflow.set_experiment(experiment_name=cls._experiment_name(project_name, name))

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
            turing.active_session.create_ensembler(ensembler.to_open_api()))
