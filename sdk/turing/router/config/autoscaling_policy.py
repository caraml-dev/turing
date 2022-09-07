from dataclasses import dataclass, field
from heapq import merge

import turing.generated.models
from turing.generated.model_utils import OpenApiModel


@dataclass
class AutoscalingPolicy:
    metric: str
    target: str

    _metric: str = field(init=False, repr=False)
    _target: str = field(init=False, repr=False)

    @property
    def metric(self) -> str:
        return self._metric

    @metric.setter
    def metric(self, metric: str):
        assert metric in {"concurrency", "cpu", "memory", "rps"}
        self._metric = metric

    @property
    def target(self) -> str:
        return self._target

    @target.setter
    def target(self, target: str):
        assert target.isnumeric()
        self._target = target

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.AutoscalingPolicy(
            metric=self.metric,
            target=self.target,
        )


DEFAULT_AUTOSCALING_POLICY = AutoscalingPolicy(metric="concurrency", target="1")
