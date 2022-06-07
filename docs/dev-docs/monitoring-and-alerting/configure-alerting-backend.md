# Introduction

The Turing UI exposes alerting configurations for various Prometheus metrics, derived from the Kube state metrics as well as the Knative default metrics. When an alert is configured by the user, the API publishes the [Prometheus alerting rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) using GitOps as an inventory for alerts. Appropriate CI/CD jobs may be configured on the git repo to apply the changes as desired (eg: publishing alerts to a Slack channel).

## GitOps Configuration

Currently, only Gitlab repositories may be configured for publishing alerts. The required client configurations (such as the Gitlab token) may be set at deploy time, under `AlertConfig.GitLab` (please refer to the [sample Helm values file](https://github.com/gojek/turing/blob/main/api/turing/config/example.yaml) for an example).

## Available Metrics

| Name | Prometheus Metric | Source |
| ---- | ----------------- | ------ |
| throughput | revision_request_count | Knative |
| latency95p | revision_request_latencies_bucket | Knative |
| error_rate | revision_request_count | Knative |
| cpu_util | container_cpu_usage_seconds_total, kube_pod_container_resource_requests_cpu_cores | Kube state |
| memory_util | container_memory_usage_bytes, kube_pod_container_resource_requests_memory_bytes | Kube state |

## Sample Alert Configuration

```yaml
groups:
    - name: development_foo_turing-hello-turing-router_cpu_util
      rules:
        - alert: turing-hello-turing-router_cpu_util_violation_development
          expr: |-
            sum by(cluster) (rate(container_cpu_usage_seconds_total{environment="staging",pod=~"turing-hello-turing-router-[0-9]*.*"}[1m])) / sum by(cluster) (kube_pod_container_resource_requests_cpu_cores{environment="staging",pod=~"turing-hello-turing-router-[0-9]*.*"}) * 100 > 90
          for: 5m
          labels:
            owner: foo
            service_name: turing-hello-turing-router
            severity: warning
          annotations:
            dashboard: http://monitoring.com/turing-dashboard?var-cluster=test-kube-cluster&var-project=test-project&var-experiment=turing-hello
            description: 'cpu_util for the past 5m: {{ $value }}%'
            playbook: http://docs.com/Alert+Troubleshooting+Playbook
            summary: 'cpu_util is higher than the threshold: 90%'
        - alert: turing-hello-turing-router_cpu_util_violation_development
          expr: |-
            sum by(cluster) (rate(container_cpu_usage_seconds_total{environment="staging",pod=~"turing-hello-turing-router-[0-9]*.*"}[1m])) / sum by(cluster) (kube_pod_container_resource_requests_cpu_cores{environment="staging",pod=~"turing-hello-turing-router-[0-9]*.*"}) * 100 > 95
          for: 5m
          labels:
            owner: foo
            service_name: turing-hello-turing-router
            severity: critical
          annotations:
            dashboard: http://monitoring.com/turing-dashboard?var-cluster=test-kube-cluster&var-project=test-project&var-experiment=turing-hello
            description: 'cpu_util for the past 5m: {{ $value }}%'
            playbook: http://docs.com/Alert+Troubleshooting+Playbook
            summary: 'cpu_util is higher than the threshold: 95%'
```