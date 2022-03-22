from __future__ import annotations

import turing
import turing.generated.models
import dataclasses
from enum import Enum

from turing.router.config.router_config import RouterConfig
from turing.router.config.log_config import RouterVersionLogConfig
from datetime import datetime


class RouterStatus(Enum):
    """
    Status of router
    """
    DEPLOYED = "deployed"
    UNDEPLOYED = "undeployed"
    FAILED = "failed"
    PENDING = "pending"


@dataclasses.dataclass
class RouterVersion(RouterConfig):
    """
    Class to used to contain a RouterVersion. Used when returning a response containing a router's version from Turing
    API. Not to be instantiated manually.
    """
    def __init__(self,
                 id: int,
                 version: int,
                 created_at: datetime,
                 updated_at: datetime,
                 status: str,
                 environment_name: str,
                 name: str,
                 monitoring_url: str,
                 **kwargs):
        self.id = id
        self.version = version
        self.created_at = created_at
        self.updated_at = updated_at
        self.environment_name = environment_name
        self.status = RouterStatus(status)
        self.name = name
        self.monitoring_url = monitoring_url
        self.log_config = RouterVersionLogConfig(**kwargs.get('log_config'))
        super().__init__(environment_name=environment_name, name=name, **kwargs)

    def get_config(self) -> RouterConfig:
        """
        Generates a RouterConfig instance from the attributes contained in this object; NOTE that the name and
        environment_name of this version gets passed to the generated RouterConfig. This means that if you were to use
        the generated RouterConfig for another instance, you would have to change its name, and also maybe its
        environment_name

        :return: a new RouterConfig instance containing attributes of this router version
        """
        return RouterConfig(**self.to_dict())

    @classmethod
    def create(cls, config: RouterConfig, router_id: int) -> RouterVersion:
        """
        Creates a new router version for the router with the given router_id WITHOUT deploying it

        :param config: configuration of router
        :param router_id: router id of the router for which this router version will be created
        :return: the new router version
        """
        version = turing.active_session.create_router_version(
            router_id=router_id,
            router_version_config=RouterVersion._get_open_api_router_version_config(config)
        )
        return RouterVersion(
            environment_name=version.router.environment_name,
            name=version.router.name,
            **version.to_dict()
        )

    @classmethod
    def _get_open_api_router_version_config(cls, config: RouterConfig) -> turing.generated.models.RouterVersionConfig:
        """
        Temporary method to construct the OpenAPI RouterVersionConfig object as opposed to using the to_open_api()
        method. To be removed once the Turing API update router endpoint has been refactored.
        """
        kwargs = {}

        if config.rules is not None:
            kwargs['rules'] = [rule.to_open_api() for rule in config.rules]
        if config.resource_request is not None:
            kwargs['resource_request'] = config.resource_request.to_open_api()
        if config.enricher is not None:
            kwargs['enricher'] = config.enricher.to_open_api()
        if config.ensembler is not None:
            kwargs['ensembler'] = config.ensembler.to_open_api()

        return turing.generated.models.RouterVersionConfig(
            routes=[route.to_open_api() for route in config.routes],
            default_route_id=config.default_route_id,
            experiment_engine=config.experiment_engine.to_open_api(),
            timeout=config.timeout,
            log_config=config.log_config.to_open_api(),
            **kwargs
        )
