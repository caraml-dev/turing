# merlin

---
![Version: 0.7.1](https://img.shields.io/badge/Version-0.7.1-informational?style=flat-square)
![AppVersion: v0.9.1](https://img.shields.io/badge/AppVersion-v0.9.1-informational?style=flat-square)

Kubernetes-friendly ML model management, deployment, and serving.

## Introduction

This Helm chart installs [Turing](https://github.com/caraml-dev/turing) and all its dependencies in a Kubernetes cluster.

## Prerequisites

To use the charts here, [Helm](https://helm.sh/) must be configured for your
Kubernetes cluster. Setting up Kubernetes and Helm is outside the scope of
this README. Please refer to the Kubernetes and Helm documentation.

- **Helm 3.0+** – This chart was tested with Helm v3.7.1, but it is also expected to work with earlier Helm versions
- **Kubernetes 1.22+** – This chart was tested with GKE v1.22.x and with [k3d](https://github.com/rancher/k3d) v1.22.x,
but it's possible it works with earlier k8s versions too
- **Istio 1.12.4+** – This chart was tested with Istio v1.12.4
- **Knative 1.7.4+, <1.8** – This chart was tested with Knative 1.7.4

It's recommended to use [turing/turing-init](https://github.com/caraml-dev/turing/blob/main/infra/charts/turing-init/README.md) Helm chart
to configure and install Istio and Knative into the cluster, before proceeding with installation of Turing.

Configuration and installation of [turing/turing-init](https://github.com/caraml-dev/turing/blob/main/infra/charts/turing-init/README.md)
is out of scope of this README, please refer to [turing/turing-init](https://github.com/caraml-dev/turing/blob/main/infra/charts/turing-init/README.md)
for installation instructions.

## Installation

### Add Helm repository

```shell
$ helm repo add turing https://turing-ml.github.io/charts
```

### Installing the chart

This command will install Turing release named `turing` in the `default` namespace.
Default chart values will be used for the installation:
```shell
$ helm install turing turing/turing
```

You can (and most likely, should) override the default configuration with values suitable for your installation.
Refer to [Configuration](#configuration) section for the detailed description of available configuration keys.

You can also refer to [values.minimal.yaml](./values.minimal.yaml) to check a minimal configuration that needs
to be provided for Turing installation.

### Uninstalling the chart

To uninstall `turing` release:
```shell
$ helm uninstall turing
```

The command removes all the Kubernetes components associated with the chart and deletes the release,
except for postgresql PVC, those will have to be removed manually.

## Configuration

The following table lists the configurable parameters of the Turing chart and their default values.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| alert.enabled | bool | `false` |  |
| alert.gitlab.alertBranch | string | `"master"` |  |
| alert.gitlab.alertRepository | string | `"lens/artillery/datascience"` |  |
| alert.gitlab.baseURL | string | `"https://gitlab.com/"` |  |
| alert.gitlab.dashboardBranch | string | `"master"` |  |
| alert.gitlab.dashboardRepository | string | `"data-science/slo-specs"` |  |
| alert.warden.apiHost | string | `""` |  |
| apiHost | string | `"http://merlin.mlp/v1"` |  |
| authorization.enabled | bool | `false` |  |
| authorization.serverUrl | string | `"http://mlp-authorization-keto"` |  |
| dockerRegistries | string | `"docker.io,ghcr.io/gojek"` |  |
| docsURL | string | `"https://github.com/gojek/merlin/blob/main/docs/getting-started/README.md"` |  |
| environment | string | `"dev"` |  |
| environmentConfigs[0].cluster | string | `"dev"` |  |
| environmentConfigs[0].cpu_limit | string | `"400m"` |  |
| environmentConfigs[0].cpu_request | string | `"100m"` |  |
| environmentConfigs[0].deployment_timeout | string | `"10m"` |  |
| environmentConfigs[0].gcp_project | string | `"gcp-project"` |  |
| environmentConfigs[0].is_default | bool | `true` |  |
| environmentConfigs[0].is_default_prediction_job | bool | `true` |  |
| environmentConfigs[0].is_prediction_job_enabled | bool | `true` |  |
| environmentConfigs[0].max_cpu | string | `"8"` |  |
| environmentConfigs[0].max_memory | string | `"8Gi"` |  |
| environmentConfigs[0].max_replica | int | `1` |  |
| environmentConfigs[0].memory_limit | string | `"500Mi"` |  |
| environmentConfigs[0].memory_request | string | `"200Mi"` |  |
| environmentConfigs[0].min_replica | int | `0` |  |
| environmentConfigs[0].name | string | `"dev"` |  |
| environmentConfigs[0].namespace_timeout | string | `"2m"` |  |
| environmentConfigs[0].prediction_job_config.driver_cpu_request | string | `"2"` |  |
| environmentConfigs[0].prediction_job_config.driver_memory_request | string | `"2Gi"` |  |
| environmentConfigs[0].prediction_job_config.executor_cpu_request | string | `"2"` |  |
| environmentConfigs[0].prediction_job_config.executor_memory_request | string | `"2Gi"` |  |
| environmentConfigs[0].prediction_job_config.executor_replica | int | `3` |  |
| environmentConfigs[0].queue_resource_percentage | string | `"20"` |  |
| environmentConfigs[0].region | string | `"id"` |  |
| global.merlin | object | `{}` |  |
| homepage | string | `"/merlin"` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.registry | string | `"ghcr.io"` |  |
| image.repository | string | `"gojek/merlin"` |  |
| image.tag | string | `"v0.9.1"` |  |
| imageBuilder.baseImage | string | `""` |  |
| imageBuilder.buildContextURI | string | `"git://TOKEN@github.com/gojek/merlin.git#refs/tags/v0.1"` |  |
| imageBuilder.clusterName | string | `"dev"` |  |
| imageBuilder.contextSubPath | string | `"python/pyfunc-server"` |  |
| imageBuilder.dockerRegistry | string | `"gojek"` |  |
| imageBuilder.dockerfilePath | string | `"./Dockerfile"` |  |
| imageBuilder.namespace | string | `"mlp"` |  |
| imageBuilder.predictionJobBaseImage | string | `"gojek/mlp/merlin-pyspark:v0.4.1"` |  |
| imageBuilder.predictionJobBuildContextURI | string | `"git://TOKEN@github.com/gojek/merlin.git#refs/tags/v0.1"` |  |
| imageBuilder.predictionJobContextSubPath | string | `"python/batch-predictor"` |  |
| imageBuilder.predictionJobDockerfilePath | string | `"docker/app.Dockerfile"` |  |
| imageBuilder.timeout | string | `"30m"` |  |
| ingress.enabled | bool | `false` |  |
| ingress.useV1Beta1 | bool | `false` | Whether to use networking.k8s.io/v1 (k8s version >= 1.19) or networking.k8s.io/v1beta1 (1.16 >= k8s version >= 1.22) |
| mlflow.artifactRoot | string | `"gs://bucket-name/mlflow"` |  |
| mlflow.image.registry | string | `"ghcr.io"` |  |
| mlflow.image.repository | string | `"gojek/mlflow"` |  |
| mlflow.image.tag | string | `"1.3.0"` |  |
| mlflow.postgresql.enabled | bool | `false` |  |
| mlflow.postgresql.postgresqlDatabase | string | `"mlflow"` |  |
| mlflow.postgresql.postgresqlUsername | string | `"mlflow"` |  |
| mlpApi.apiHost | string | `"http://mlp.mlp:8080/v1"` |  |
| mlpApi.encryptionKey | string | `""` |  |
| monitoring.baseURL | string | `""` |  |
| monitoring.enabled | bool | `false` |  |
| monitoring.jobBaseURL | string | `""` |  |
| newrelic.appname | string | `"merlin-api-dev"` |  |
| newrelic.enabled | bool | `false` |  |
| newrelic.licenseSecretName | string | `"newrelic-license-secret"` |  |
| postgresql.metrics.enabled | bool | `false` |  |
| postgresql.metrics.serviceMonitor.enabled | bool | `false` |  |
| postgresql.persistence.enabled | bool | `true` |  |
| postgresql.persistence.size | string | `"10Gi"` |  |
| postgresql.postgresqlDatabase | string | `"merlin"` |  |
| postgresql.postgresqlPassword | string | `"merlin"` |  |
| postgresql.postgresqlUsername | string | `"merlin"` |  |
| postgresql.replication.applicationName | string | `"merlin"` |  |
| postgresql.replication.enabled | bool | `false` |  |
| postgresql.replication.numSynchronousReplicas | int | `2` |  |
| postgresql.replication.password | string | `"repl_password"` |  |
| postgresql.replication.slaveReplicas | int | `2` |  |
| postgresql.replication.synchronousCommit | string | `"on"` |  |
| postgresql.replication.user | string | `"repl_user"` |  |
| postgresql.resources.requests.cpu | string | `"500m"` |  |
| postgresql.resources.requests.memory | string | `"256Mi"` |  |
| replicaCount | string | `"1"` |  |
| resources.limits.cpu | string | `"1"` |  |
| resources.limits.memory | string | `"1Gi"` |  |
| resources.requests.cpu | string | `"500m"` |  |
| resources.requests.memory | string | `"1Gi"` |  |
| sentry.dsn | string | `""` |  |
| sentry.enabled | bool | `false` |  |
| service.externalPort | int | `8080` |  |
| service.internalPort | int | `8080` |  |
| swagger.apiHost | string | `"merlin.dev"` |  |
| swagger.basePath | string | `"/api/merlin/v1"` |  |
| swagger.enabled | bool | `false` |  |
| swagger.image.tag | string | `"v3.23.5"` |  |
| swagger.service.externalPort | int | `8080` |  |
| swagger.service.internalPort | int | `8080` |  |
| vault.enabled | bool | `true` |  |
| vault.secretName | string | `"vault-secret"` |  |
| vault.server.dev.enabled | bool | `true` |  |
| vault.server.postStart[0] | string | `"/bin/sh"` |  |
| vault.server.postStart[1] | string | `"-c"` |  |
| vault.server.postStart[2] | string | `"sleep 5 &&\n  vault secrets disable secret/ &&\n  vault secrets enable -path=secret -version=1 kv"` |  |
