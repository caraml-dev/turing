import abc
from enum import Enum
from typing import Iterable, MutableMapping, Optional, Dict, List
import turing.generated.models
from turing._base_types import DataObject
from turing.generated.model_utils import OpenApiModel


class ResourceRequest:
    def __init__(self, min_replica: int, max_replica: int, cpu_request: str, memory_request: str):
        self._min_replica = min_replica
        self._max_replica = max_replica
        self._cpu_request = cpu_request
        self._memory_request = memory_request

    @property
    def min_replica(self) -> int:
        return self._min_replica

    @property
    def max_replica(self) -> int:
        return self._max_replica

    @property
    def cpu_request(self) -> str:
        return self._cpu_request

    @property
    def memory_request(self) -> str:
        return self._memory_request

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.ResourceRequest(
            min_replica=self.min_replica,
            max_replica=self.max_replica,
            cpu_request=self.cpu_request,
            memory_request=self.memory_request
        )

