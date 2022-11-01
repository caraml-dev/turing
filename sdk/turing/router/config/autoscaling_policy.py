from dataclasses import dataclass, field
from enum import Enum
from typing import Union, Optional

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
    """Class to create a new autoscaling policy. This must be set within each router, enricher, docker or pyfunc
    ensembler deployment, where only ONE of the following conditions must be true:

    - the fields metric and target are BOTH set (user-defined autoscaling policy is used)
        In this set-up, the autoscaling metric and target will be used together with the resource requests values set in
        the resource_request field of the router, enricher, docker or pyfunc ensembler to scale up or down the
        deployment automatically.

    - the field payload_size is set (default autoscaling policy is used)
        In this set-up, the autoscaling metric and target, as well as resource requests will be determined automatically
        under the hood based on the payload_size set. As such the resource_request field of the router, enricher, docker
        or pyfunc ensembler MUST be left as None.

    :param metric: type of the autoscaling metric
    :param target: target value of the autoscaling metric; measured in % if the metric is cpu or memory
    :param payload_size: estimated payload size that this component will receive in requests
    """

    metric: AutoscalingMetric = None
    target: str = None
    payload_size: str = None

    _metric: AutoscalingMetric = field(init=False, repr=False)
    _target: str = field(init=False, repr=False)
    _payload_size: str = field(init=False, repr=False)

    def __init__(
        self,
        metric: Optional[Union[str, AutoscalingMetric]] = None,
        target: Optional[str] = None,
        payload_size: Optional[str] = None,
    ):
        self.metric = metric
        self.target = target
        self.payload_size = payload_size

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

    @property
    def payload_size(self) -> str:
        return self._payload_size

    @payload_size.setter
    def payload_size(self, payload_size: str):
        self._payload_size = payload_size

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.AutoscalingPolicy(
            metric=self.metric.value, target=self.target, payload_size=self.payload_size
        )


DEFAULT_AUTOSCALING_POLICY = AutoscalingPolicy(
    metric=AutoscalingMetric.CONCURRENCY, target="1"
)
