from turing.router.config.router_config import RouterConfig
from turing.router.config.log_config import RouterVersionLogConfig
from datetime import datetime


class RouterVersion(RouterConfig):
    def __init__(self,
                 created_at: datetime,
                 updated_at: datetime,
                 environment_name: str,
                 name: str,
                 **kwargs):
        self.created_at = created_at
        self.updated_at = updated_at
        self.environment_name = environment_name
        self.name = name
        self.log_config = RouterVersionLogConfig(**kwargs.get('log_config'))
        super().__init__(environment_name=environment_name, name=name, **kwargs)

    def get_config(self):
        a = self.to_dict()
        return RouterConfig(**a)
