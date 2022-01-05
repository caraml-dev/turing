import abc
from enum import Enum
from typing import Iterable, MutableMapping, Optional, Dict, List
import turing.generated.models
from turing._base_types import DataObject
from turing.generated.model_utils import OpenApiModel


class FieldSource(Enum):
    HEADER = "header"
    PAYLOAD = "payload"


class TrafficRuleCondition:
    def __init__(self,
                 field_source: str,
                 field: str,
                 operator: str,
                 values: List[str]):
        assert operator == "in"
        self._field_source = field_source
        self._field = field
        self._operator = operator
        self._values = values

    @property
    def field_source(self) -> str:
        return self._field_source

    @property
    def field(self) -> str:
        return self._field

    @property
    def operator(self) -> str:
        return self._operator

    @property
    def values(self) -> List[str]:
        return self._values


class Route:
    def __init__(self,
                 id: str,
                 type: str,
                 endpoint: str,
                 timeout: str,
                 annotations: Dict = None):
        self._id = id
        self._type = type
        self._endpoint = endpoint
        self._timeout = timeout
        self._annotations = annotations

    @property
    def id(self) -> str:
        return self._id

    @property
    def type(self) -> str:
        return self._type

    @property
    def endpoint(self) -> str:
        return self._endpoint

    @property
    def timeout(self) -> str:
        return self._timeout

    @property
    def annotations(self) -> Dict:
        return self._annotations

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.Route(
            id=self.id,
            type=self.type,
            endpoint=self.endpoint,
            timeout=self.timeout,
            annotations=self.annotations
        )
