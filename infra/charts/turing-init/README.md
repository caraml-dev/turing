# turing-init

---
![Version: 0.0.2](https://img.shields.io/badge/Version-0.0.2-informational?style=flat-square)

A Helm chart for Kubernetes

## Introduction

This Helm chart installs the infrastructure components [Turing](https://github.com/gojek/turing) requires.

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://googlecloudplatform.github.io/spark-on-k8s-operator | spark-operator | 1.1.7 |

## Installation

### Add Helm repository

```shell
$ helm repo add turing https://turing-ml.github.io/charts
```

### Installing the chart

This command will install turing-init named `turing-init` in the `default` namespace.
Default chart values will be used for the installation:
```shell
$ helm install turing-init turing/turing-init
```

It is unlikely that you would need to change the values of the default configuration.

### Uninstalling the chart

To uninstall `turing-init` release:
```shell
$ helm uninstall turing-init
```

The command removes all the Kubernetes components associated with the chart.

## Configuration

The following table lists the configurable parameters of the Turing chart and their default values.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.registry | string | `"ghcr.io/"` | Docker registry for Turing cluster init |
| image.repository | string | `"gojek/turing/cluster-init"` | Docker image repository for Turing cluster init |
| image.tag | string | `"latest"` | Docker image tag for Turing cluster init |
| istio.operatorConfig | object | `{"apiVersion":"install.istio.io/v1alpha1","kind":"IstioOperator","spec":{"addonComponents":{"pilot":{"enabled":true}},"components":{"ingressGateways":[{"enabled":true,"name":"istio-ingressgateway"},{"enabled":true,"k8s":{"service":{"ports":[{"name":"status-port","port":15020},{"name":"http2","port":80},{"name":"https","port":443}],"type":"ClusterIP"}},"label":{"app":"cluster-local-gateway","istio":"cluster-local-gateway"},"name":"cluster-local-gateway"}]},"values":{"gateways":{"istio-ingressgateway":{"runAsRoot":true}},"global":{"proxy":{"autoInject":"enabled"},"useMCP":false}}}}` | istio operator config, defaults are the minimum to run turing, see https://istio.io/v1.9/docs/reference/config/istio.operator.v1alpha1/ |
| istio.version | string | `"1.9.9"` | Istio version to use |
| knative.domains | string | `""` | Knative domains, comma seperated values, i.e. www.example.com,www.gojek.com |
| knative.istioVersion | string | `"0.18.1"` | Knative Istio Version to use |
| knative.registriesSkippingTagResolving | string | `""` | Knative registries skipping tag resolving, comma seperated values, i.e. www.example.com,www.gojek.com |
| knative.version | string | `"0.18.3"` | Knative Version to use |
| spark-operator | object | `{"image":{"repository":"gcr.io/spark-operator/spark-operator","tag":"v1beta2-1.2.3-3.1.1"},"replicas":1,"resources":{},"webhook":{"enable":true}}` | Override any spark-operator values here: https://github.com/GoogleCloudPlatform/spark-on-k8s-operator/blob/master/charts/spark-operator-chart/README.md |
| spark-operator.image.repository | string | `"gcr.io/spark-operator/spark-operator"` | repository of the spark operator |
| spark-operator.image.tag | string | `"v1beta2-1.2.3-3.1.1"` | image tag of the spark operator |
| spark-operator.replicas | int | `1` | number of replicas |
| spark-operator.resources | object | `{}` | Resources requests and limits for spark operator. This should be set  according to your cluster capacity and service level objectives. Reference: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| spark-operator.webhook.enable | bool | `true` | this is needed to be set to true, if not the configmaps will not load |
