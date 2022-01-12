import turing.generated.models
from turing.generated.model_utils import OpenApiModel
from turing.router.config.common.schemas import EnvVarNameSchema


class EnvVar:
    def __init__(self,
                 name: str,
                 value: str):
        EnvVarNameSchema.verify_regex(name)
        self.name = name
        self.value = value

    @property
    def name(self) -> str:
        return self._name

    @name.setter
    def name(self, name):
        EnvVarNameSchema.verify_regex(name)
        self._name = name

    @property
    def value(self) -> str:
        return self._value

    @value.setter
    def value(self, value:str):
        self._value = value

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.EnvVar(
            name=self.name,
            value=self.value
        )