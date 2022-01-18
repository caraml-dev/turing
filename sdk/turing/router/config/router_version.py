from turing.router.config.router_config import RouterConfig
from turing.router.config.log_config import RouterVersionLogConfig
from datetime import datetime


class RouterVersion(RouterConfig):
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
        self.status = status
        self.name = name
        self.monitoring_url = monitoring_url
        self.log_config = RouterVersionLogConfig(**kwargs.get('log_config'))
        super().__init__(environment_name=environment_name, name=name, **kwargs)

    def get_config(self) -> RouterConfig:
        return RouterConfig(**self.to_dict())
