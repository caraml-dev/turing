from typing import Optional, List
import turing.generated.models
from turing._base_types import ApiObject, ApiObjectSpec


@ApiObjectSpec(turing.generated.models.Project)
class Project(ApiObject):
    """
    API entity for MLP Project
    """

    def __init__(self, name: str, mlflow_tracking_url: str, **kwargs):
        super().__init__(**kwargs)
        self._name = name
        self._mlflow_tracking_url = mlflow_tracking_url

    @property
    def name(self) -> str:
        return self._name

    @name.setter
    def name(self, name: str):
        self._name = name

    @property
    def mlflow_tracking_url(self) -> str:
        return self._mlflow_tracking_url

    @mlflow_tracking_url.setter
    def mlflow_tracking_url(self, mlflow_tracking_url: str):
        self._mlflow_tracking_url = mlflow_tracking_url

    @classmethod
    def list(cls, name: Optional[str] = None) -> List["Project"]:
        response = turing.active_session.list_projects(name=name)
        return [Project.from_open_api(item) for item in response]
