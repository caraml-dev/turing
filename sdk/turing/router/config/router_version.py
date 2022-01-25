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
