from dataclasses import dataclass, field

import turing.generated.models
from turing.generated.model_utils import OpenApiModel


@dataclass
class MountedMLPSecret:
    mlp_secret_name: str
    env_var_name: str

    _mlp_secret_name: str = field(init=False, repr=False)
    _env_var_name: str = field(init=False, repr=False)

    @property
    def mlp_secret_name(self) -> str:
        return self._mlp_secret_name

    @mlp_secret_name.setter
    def mlp_secret_name(self, mlp_secret_name):
        self._mlp_secret_name = mlp_secret_name

    @property
    def env_var_name(self) -> str:
        return self._env_var_name

    @env_var_name.setter
    def env_var_name(self, env_var_name):
        self._env_var_name = env_var_name

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.MountedMLPSecret(
            mlp_secret_name=self.mlp_secret_name, env_var_name=self.env_var_name
        )
