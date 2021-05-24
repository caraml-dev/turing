from typing import Optional, List
import turing.generated.models
from turing._base_types import ApiObject, ApiObjectSpec


@ApiObjectSpec(turing.generated.models.Project)
class Project(ApiObject):

    def __init__(self, name: str, **kwargs):
        super().__init__(**kwargs)
        self._name = name

    @property
    def name(self) -> str:
        return self._name

    @name.setter
    def name(self, name: str):
        self._name = name

    @classmethod
    def list(cls, name: Optional[str] = None) -> List['Project']:
        from turing.session import active_session

        response = active_session.list_projects(name=name)
        return [Project.from_open_api(item) for item in response]
