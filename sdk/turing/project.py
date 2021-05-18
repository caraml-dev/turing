import turing.utils
from turing._base_types import ApiObject, ApiObjectSpec
import turing.generated.models


@turing.utils.autostr
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
