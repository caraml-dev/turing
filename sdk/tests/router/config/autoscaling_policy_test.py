import pytest
from turing.router.config.autoscaling_policy import (
    AutoscalingMetric,
    AutoscalingPolicy,
    DEFAULT_AUTOSCALING_POLICY,
    InvalidAutoscalingMetric,
)


def test_set_invalid_metric():
    policy = AutoscalingPolicy(
        metric=DEFAULT_AUTOSCALING_POLICY.metric,
        target=DEFAULT_AUTOSCALING_POLICY.target,
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
def test_set_autoscaling_policy_metric(metric, expected):
    policy = AutoscalingPolicy(metric=metric, target=DEFAULT_AUTOSCALING_POLICY.target)
    assert policy.metric == expected


@pytest.mark.parametrize(
    "target,expected,error",
    [
        pytest.param(
            "100",
            "100",
            None,
        ),
        pytest.param(
            "0.5",
            "0.5",
            None,
        ),
        pytest.param(
            "0.5.5",
            None,
            AssertionError,
        ),
    ],
)
def test_set_autoscaling_policy_target(target, expected, error):
    if error is not None:
        with pytest.raises(error):
            AutoscalingPolicy(metric=AutoscalingMetric.CONCURRENCY, target=target)
    else:
        policy = AutoscalingPolicy(metric=AutoscalingMetric.CONCURRENCY, target=target)
        assert policy.target == expected
