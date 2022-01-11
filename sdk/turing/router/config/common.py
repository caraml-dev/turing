import re
import turing.generated.models
from turing.generated.model_utils import OpenApiModel


class EnvVar:
    def __init__(self,
                 name: str,
                 value: str):
        self._name = name
        self._value = value

    @property
    def name(self) -> str:
        return self._name

    @property
    def value(self) -> str:
        return self._value

    @classmethod
    def verify_name(cls, name):
        matched = re.fullmatch(r"^[a-z0-9_]*$", name, re.IGNORECASE)
        if bool(matched) is False:
            raise InvalidEnvironmentVariableNameException(
                f"The name of a variable can contain only alphanumeric character or the underscore; "
                f"name passed: {name}"
            )

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.EnvVar(
            name=self.name,
            value=self.value
        )


class InvalidEnvironmentVariableNameException(Exception):
    pass
