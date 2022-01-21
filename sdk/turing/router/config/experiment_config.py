from dataclasses import dataclass, field
from typing import Dict

import turing.generated.models
from turing.generated.model_utils import OpenApiModel


@dataclass
class ExperimentConfig:
    type: str = "nop"
    config: Dict = None

    def __init__(self, type: str = "nop", config: Dict = None, plugin_config: Dict = None):
        self.type = type
        self.config = config
        self.plugin_config = plugin_config

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

    @property
    def plugin_config(self) -> Dict:
        return self._plugin_config

    @plugin_config.setter
    def plugin_config(self, plugin_config: Dict):
        self._plugin_config = plugin_config

    def to_open_api(self) -> OpenApiModel:
        if self.config is None:
            config = {}
        else:
            config = self.config

        if self.plugin_config is None:
            plugin_config = {}
        else:
            plugin_config = self.plugin_config

        return turing.generated.models.ExperimentConfig(
            type=self.type,
            config=config,
            plugin_config=turing.generated.models.ExperimentConfigPluginConfig(**plugin_config)
        )

