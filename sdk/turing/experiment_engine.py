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


# class StandardExperimentConfig:
#     def __init__(self,
#                  client = None,
#                  experiments = None,
#                  variables = None):
#         self.client = client
#         self.experiments = experiments
#         self.variables = variables
#
#     @property
#     def client(self):
#         return self._client
#
#     @client.setter
#     def client(self, client):
#         self._client = client
#
#     @property
#     def experiments(self):
#         return self._experiments
#
#     @experiments.setter
#     def experiments(self, experiments):
#         self._experiments = experiments
#
#     @property
#     def variables(self):
#         return self._variables
#
#     @variables.setter
#     def variables(self, variables):
#         self._variables = variables
#
#
# class ExperimentClient:
#     def __init__(self, id: str, username: str, passkey: str = None):
#         self.id = id
#         self.username = username
#         self.passkey = passkey
#
#     @property
#     def id(self) -> str:
#         return self._id
#
#     @id.setter
#     def id(self, id: str):
#         self._id = id
#
#     @property
#     def username(self) -> str:
#         return self._username
#
#     @username.setter
#     def username(self, username: str):
#         self._username = username
#
#     @property
#     def passkey(self) -> str:
#         return self._passkey
#
#     @passkey.setter
#     def passkey(self, passkey: str):
#         self._passkey = passkey
#
