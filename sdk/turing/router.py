from typing import Optional, List
import turing.generated.models
from turing._base_types import ApiObject, ApiObjectSpec


@ApiObjectSpec(turing.generated.models.Router)
class Router(ApiObject):
    """
    API entity for Router
    """
    def __init__(self,
                 id: int,
                 project_id: int,
                 environment_name: str,
                 endpoint: str,
                 status: str,
                 **kwargs):
        super(Router, self).__init__(**kwargs)
        self._id = int(id)
        self._project_id = int(project_id)
        self._environment_name = environment_name
        self._endpoint = endpoint
        self._status = status

    @property
    def id(self) -> int:
        return self._id

    @property
    def project_id(self) -> int:
        return self._project_id

    @property
    def environment_name(self) -> str:
        return self._environment_name

    @property
    def endpoint(self) -> str:
        return self._endpoint

    @property
    def status(self) -> str:
        return self._status

    @classmethod
    def list(cls) -> List['Router']:
        """
        List routers in the active project

        :return: list of routers
        """
        response = turing.active_session.list_routers()
        return [Router.from_open_api(item) for item in response]
