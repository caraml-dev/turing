import abc
from typing import Iterable, MutableMapping, Optional, Dict
import turing.generated.models
from turing._base_types import DataObject
from turing.generated.model_utils import OpenApiModel


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
