from dataclasses import dataclass, field
from typing import Dict

import turing.generated.models
from turing.generated.model_utils import OpenApiModel


@dataclass
class ExperimentConfig:
    type: str = "nop"
    config: Dict = None

    def __init__(self, type: str = "nop", config: Dict = None, **kwargs):
        self.type = type
        self.config = config
        self.__dict__.update(kwargs)

    @property
    def type(self) -> str:
        return self._type
    
    @type.setter
    def type(self, type: str):
        self._type = type

    @property
    def config(self) -> Dict:
        return self._config

    @config.setter
    def config(self, config: Dict):
        if config is not None and 'project_id' in config:
            config['project_id'] = int(config['project_id'])
        self._config = config

    def to_open_api(self) -> OpenApiModel:
        if self.config is None:
            config = {}
        else:
            config = self.config

        return turing.generated.models.ExperimentConfig(
            type=self.type,
            config=config
        )
