import turing.generated.models
from .config import EnsemblingJobConfig, ResourceRequest, ResultConfig
from .source import *
from .sink import *


class ResultType:
    DOUBLE = turing.generated.models.EnsemblingJobResultType("DOUBLE")
    FLOAT = turing.generated.models.EnsemblingJobResultType("FLOAT")
    INTEGER = turing.generated.models.EnsemblingJobResultType("INTEGER")
    LONG = turing.generated.models.EnsemblingJobResultType("LONG")
    STRING = turing.generated.models.EnsemblingJobResultType("STRING")
    ARRAY = turing.generated.models.EnsemblingJobResultType("ARRAY")


__all__ = [
    "EnsemblingJobConfig", "ResourceRequest",
    "ResultConfig", "ResultType",
]
