from dataclasses import dataclass, field
from enum import Enum
from typing import Union

import turing.generated.models
from turing.generated.model_utils import OpenApiModel


class AutoscalingMetric(Enum):
    """
    Autoscaling metrics supported by Turing
    """

    CONCURRENCY = "concurrency"
    CPU = "cpu"
    MEMORY = "memory"
    RPS = "rps"

    @staticmethod
    def from_string(metric_name: str):
        try:
            return {
                "concurrency": AutoscalingMetric.CONCURRENCY,
                "cpu": AutoscalingMetric.CPU,
                "memory": AutoscalingMetric.MEMORY,
                "rps": AutoscalingMetric.RPS,
            }[metric_name]
        except:
            raise InvalidAutoscalingMetric(f"Unknown metric {metric_name}")


class InvalidAutoscalingMetric(Exception):
    pass


@dataclass
class AutoscalingPolicy:
    """Class to create a new autoscaling policy.

    :param metric: type of the autoscaling metric
    :param target: target value of the autoscaling metric; measured in % if the metric is cpu or memory.
    """

    metric: AutoscalingMetric
    target: str

    _metric: AutoscalingMetric = field(init=False, repr=False)
    _target: str = field(init=False, repr=False)

    def __init__(self, metric: Union[str, AutoscalingMetric], target: str):
        self.metric = metric
        self.target = target

    @property
    def metric(self) -> AutoscalingMetric:
        return self._metric

    @metric.setter
    def metric(self, metric: Union[str, AutoscalingMetric]):
        if isinstance(metric, str):
            self._metric = AutoscalingMetric.from_string(metric)
        else:
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
            metric=self.metric.value,
            target=self.target,
        )


DEFAULT_AUTOSCALING_POLICY = AutoscalingPolicy(
    metric=AutoscalingMetric.CONCURRENCY, target="1"
)
