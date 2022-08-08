import time
import logging

from typing import List, Dict, Optional

import turing.generated.models
from turing._base_types import ApiObject, ApiObjectSpec
from turing.router.config.router_config import RouterConfig
from turing.router.config.router_version import RouterVersion, RouterStatus


logger = logging.getLogger("router_sdk_logger")
logger.setLevel(level=logging.INFO)

ch = logging.StreamHandler()
ch.setLevel(logging.DEBUG)

formatter = logging.Formatter("%(asctime)s - %(name)s - %(levelname)s - %(message)s")
ch.setFormatter(formatter)

logger.addHandler(ch)


@ApiObjectSpec(turing.generated.models.Router)
class Router(ApiObject):
    """
    API entity for Router
    """

    def __init__(
        self,
        id: int,
        name: str,
        project_id: int,
        environment_name: str,
        monitoring_url: str,
        status: str,
        config: Dict = None,
        endpoint: str = None,
        **kwargs,
    ):
        super(Router, self).__init__(**kwargs)
        self._id = id
        self._name = name
        self._project_id = project_id
        self._environment_name = environment_name
        self._endpoint = endpoint
        self._monitoring_url = monitoring_url
        self._status = RouterStatus(status)
        self._config = config

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
    def config(self) -> Optional["RouterConfig"]:
        if self._config is not None:
            return RouterConfig(
                name=self.name, environment_name=self.environment_name, **self._config
            )
        else:
            return None

    @property
    def version(self) -> int:
        return self._config.get("version") if self._config else None

    @classmethod
    def list(cls) -> List["Router"]:
        """
        List routers in the active project

        :return: list of routers
        """
        response = turing.active_session.list_routers()
        return [Router.from_open_api(item) for item in response]

    @classmethod
    def create(cls, config: RouterConfig) -> "Router":
        """
        Create router with a given configuration

        :param config: configuration of router
        :return: instance of router created
        """
        return Router.from_open_api(
            turing.active_session.create_router(router_config=config.to_open_api())
        )

    @classmethod
    def delete(cls, router_id: int) -> int:
        """
        Delete specific router given its router ID

        :param router_id: router_id of the router to be deleted
        :return: router_id of the deleted router
        """
        return turing.active_session.delete_router(router_id=router_id).id

    @classmethod
    def get(cls, router_id: int) -> "Router":
        """
        Fetch router by its router ID

        :param router_id: router_id of the router to be fetched
        :return: router with the corresponding id
        """
        return Router.from_open_api(
            turing.active_session.get_router(router_id=router_id)
        )

    def update(self, config: RouterConfig) -> "Router":
        """
        Update the current router with a new set of configs specified in the RouterConfig argument

        :param config: configuration of router
        :return: instance of router (self); this router contains details of the router and its currently deployed version
        """
        self._config = config
        updated_router = Router.from_open_api(
            turing.active_session.update_router(
                router_id=self.id, router_config=config.to_open_api()
            )
        )
        self.__dict__ = updated_router.__dict__
        return self

    def create_version(self, config: RouterConfig) -> "RouterVersion":
        """
        Creates a new router version for the router WITHOUT deploying it

        :param config: configuration of router
        :return: the new router version
        """
        return RouterVersion.create(config=config, router_id=self.id)

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

    def list_versions(self) -> List["RouterVersion"]:
        """
        List router versions for this router

        :return: list of router versions
        """
        response = turing.active_session.list_router_versions(router_id=self.id)
        return [
            RouterVersion(
                environment_name=self.environment_name, name=self.name, **ver.to_dict()
            )
            for ver in response
        ]

    def get_version(self, version: int) -> "RouterVersion":
        """
        Fetch a version of this router given a version number

        :return: list of router versions
        """
        version = turing.active_session.get_router_version(
            router_id=self.id, version=version
        )
        return RouterVersion(
            environment_name=self.environment_name, name=self.name, **version.to_dict()
        )

    def delete_version(self, version: int) -> Dict[str, int]:
        """
        Delete a version of this router given a version number


        :return: router_id and deleted version of this router
        """
        return turing.active_session.delete_router_version(
            router_id=self.id, version=version
        ).to_dict()

    def deploy_version(self, version: int) -> Dict[str, int]:
        """
        Deploy specific router version by its router ID and version

        :return: router_id and version of this router
        """
        return turing.active_session.deploy_router_version(
            router_id=self.id, version=version
        ).to_dict()

    def get_events(self) -> List[turing.generated.models.Event]:
        """
        Fetch deployment events associated with the router

        :return: list of events involving this router
        """
        response = turing.active_session.get_router_events(router_id=self.id).get(
            "events"
        )
        return [event for event in response] if response else []

    def wait_for_status(
        self, status: RouterStatus, max_tries: int = 15, duration: float = 10.0
    ):
        for i in range(1, max_tries + 1):
            logger.debug(f"Checking if router {self.id} is {status.value}...")
            current_router = Router.get(self.id)
            cur_status = current_router.status
            if cur_status == status:
                # Wait for backend components to fully resolve
                time.sleep(5)
                logger.debug(f"Router {self.id} is finally {status.value}.")
                self.__dict__ = current_router.__dict__
                return
            else:
                logger.debug(f"Router {self.id} is {cur_status.value}.")
                logger.debug(
                    f"Retrying {i}/{max_tries} time(s): waiting for {duration} seconds before retrying..."
                )
                time.sleep(duration)

        raise TimeoutError

    def wait_for_version_status(
        self,
        status: RouterStatus,
        version: int,
        max_tries: int = 15,
        duration: float = 10.0,
    ):
        for i in range(1, max_tries + 1):
            logger.debug(
                f"Checking if router {self.id} with version {version} is {status.value}..."
            )
            cur_status = self.get_version(version).status
            if cur_status == status:
                # Wait for backend components to fully resolve
                time.sleep(5)
                logger.debug(
                    f"Router {self.id} with version {version} is finally {status.value}."
                )
                return
            else:
                logger.debug(
                    f"Router {self.id} with version {version} is {cur_status.value}."
                )
                logger.debug(
                    f"Retrying {i}/{max_tries} time(s): waiting for {duration} seconds before retrying..."
                )
                time.sleep(duration)

        raise TimeoutError
