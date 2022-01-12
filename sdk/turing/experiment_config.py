import turing.generated.models
from typing import List, Dict
from turing.generated.model_utils import OpenApiModel


class ExperimentConfig:
    def __init__(self,
                 type: str = "nop",
                 config: Dict = None):
        self.type = type
        self.config = config

    @property
    def type(self) -> str:
        return self._type
    
    @type.setter
    def type(self, type: str):
        self._type = type

    @property
    def config(self):
        return self._config

    @config.setter
    def config(self, config):
        if config is not None and 'project_id' in config:
            config['project_id'] = int(config['project_id'])
        self._config = config

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.ExperimentConfig(
            type=self.type,
            config=self.config
        )