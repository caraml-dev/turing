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
        else:
            self._config = None

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
    def create(cls, config: RouterConfig) -> 'Router':
        """
        Create router with a given configuration

        :param config: configuration of router
        :return: instance of router created
        """
        return Router.from_open_api(turing.active_session.create_router(router_config=config.to_open_api()))

    @classmethod
    def delete(cls, router_id: int) -> int:
        """
        Delete specific router given its router ID

        :param router_id: router_id of the router to be deleted
        :return: router_id of the deleted router
        """
        return turing.active_session.delete_router(router_id=router_id).id

    @classmethod
    def get(cls, router_id: int) -> 'Router':
        """
        Fetch router by its router ID

        :param router_id: router_id of the router to be fetched
        :return: router with the corresponding id
        """
        return Router.from_open_api(
            turing.active_session.get_router(router_id=router_id))

    def update(self, config: RouterConfig) -> 'Router':
        """
        Update the current router with a new set of configs specified in the RouterConfig argument

        :param config: configuration of router
        :return: instance of router updated (self)
        """
        self._config = config
        return Router.from_open_api(turing.active_session.update_router(router_id=self.id,
                                                                        router_config=config.to_open_api()))

    def deploy(self) -> Dict[str, int]:
        """
        Deploy this router

        :return: router_id and version of this router
        """
        return turing.active_session.deploy_router(router_id=self.id).to_dict()

    def undeploy(self) -> Dict[str, int]:
        """
        Undeploy this router

        :return: router_id of this router
        """
        return turing.active_session.undeploy_router(router_id=self.id).to_dict()
