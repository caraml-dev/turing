# Introduction

## Monitoring of Real-Time Routers

Turing routers (as well as enrichers and ensemblers) are deployed using the Knative framework, on top of the Istio Service mesh. Both tools provide some out-of-the box Prometheus metrics to track common stats such as error rate, request latency, etc. and can be configured by following the official guides:
* [Knative](https://knative.dev/docs/serving/observability/metrics/serving-metrics/#autoscaler)
* [Istio](https://istio.io/v1.9/docs/tasks/observability/metrics/)

The following custom Prometheus metrics are published by the Turing routers too.

| Metric Name | Description | Type | Tags | Unit |
| ----------- | ----------- | ---- | ---- | ---- |
| mlp_turing_exp_engine_request_duration_ms | The duration for fetching a treatment from the experiment engine | Histogram | `status`, `engine` | Milliseconds |
| mlp_route_request_duration_ms | The duration for the call to a route | Histogram | `status`, `route` | Milliseconds |
| mlp_turing_comp_request_duration_ms | The duration for a custom operation in the code, useful for debugging | Histogram | `status`, `component` | | Milliseconds |

Users are also free to publish their own custom metrics from the Enricher / Ensembler. All custom metrics (from the router, enricher or ensembler) should be scraped from the `user-container` pods for use.

### Configuring the Monitoring URL on Turing

Once the required dashboard has been created using the Prometheus metrics and other data, the template string `RouterDefaults.MonitoringURLFormat` can be used to configure the monitoring URL, when deploying the Turing application (please refer to the [sample Helm values file](https://github.com/gojek/turing/blob/main/api/turing/config/example.yaml) for an example). This URL will be used by Turing when creating / editing a router and subsequently, in the Turing UI as a navigation link to the dashboard.

## Monitoring Batch Ensemblers

TBD