import pytest
from turing.router.config.autoscaling_policy import (
    AutoscalingMetric,
    AutoscalingPolicy,
    InvalidAutoscalingMetric,
)


def test_set_invalid_metric():
    policy = AutoscalingPolicy(
        metric=AutoscalingMetric.CONCURRENCY,
        target="1",
    )
    with pytest.raises(InvalidAutoscalingMetric):
        policy.metric = "invalid"


@pytest.mark.parametrize(
    "metric,expected",
    [
        pytest.param(
            AutoscalingMetric.CONCURRENCY,
            AutoscalingMetric.CONCURRENCY,
        ),
        pytest.param(
            "concurrency",
            AutoscalingMetric.CONCURRENCY,
        ),
        pytest.param(
            "cpu",
            AutoscalingMetric.CPU,
        ),
        pytest.param(
            "rps",
            AutoscalingMetric.RPS,
        ),
        pytest.param(
            "memory",
            AutoscalingMetric.MEMORY,
        ),
    ],
)
def test_set_route_with_invalid_endpoint(metric, expected):
    policy = AutoscalingPolicy(metric=metric, target="1")
    assert policy.metric == expected
