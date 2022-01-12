from enum import Enum

from typing import List, Dict

import turing.generated.models
from turing._base_types import ApiObject, ApiObjectSpec
from turing.router.config.router_config import RouterConfig


class RouterStatus(Enum):
    """
    Status of router
    """
    DEPLOYED = "deployed"
    UNDEPLOYED = "undeployed"
    FAILED = "failed"
    PENDING = "pending"


@ApiObjectSpec(turing.generated.models.Router)
class Router(ApiObject):
    """
    API entity for Router
    """
    def __init__(self,
                 id: int,
                 name: str,
                 project_id: int,
                 environment_name: str,
                 monitoring_url: str,
                 status: str,
                 config: Dict = None,
                 endpoint: str = None,
                 **kwargs):
        super(Router, self).__init__(**kwargs)
        self._id = id
        self._name = name
        self._project_id = project_id
        self._environment_name = environment_name
        self._endpoint = endpoint
        self._monitoring_url = monitoring_url
        self._status = RouterStatus(status)
        if config is not None:
            self._config = RouterConfig(name=name, environment_name=environment_name, **config)

    @property
    def id(self) -> int:
        return self._id

    @property
    def name(self) -> str:
        return self._name

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
    def monitoring_url(self) -> str:
        return self._monitoring_url

    @property
    def status(self) -> RouterStatus:
        return self._status

    @property
    def config(self) -> 'RouterConfig':
        return self._config

    @classmethod
    def list(cls) -> List['Router']:
        """
        List routers in the active project

        :return: list of routers
        """
        response = turing.active_session.list_routers()
        return [Router.from_open_api(item) for item in response]

    @classmethod
    def create(cls, config: turing.generated.models.RouterConfig) -> 'Router':
        """
        Create router with a given configuration

        :param config: configuration of router
        :return: instance of router created
        """
        return Router.from_open_api(turing.active_session.create_router(router_config=config))
