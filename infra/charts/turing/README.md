# turing

---
![Version: 0.2.1](https://img.shields.io/badge/Version-0.2.1-informational?style=flat-square)
![AppVersion: v1.0.0](https://img.shields.io/badge/AppVersion-v1.0.0-informational?style=flat-square)

Turing: ML Experimentation System

## Introduction

This Helm chart installs [Turing](https://github.com/gojek/turing) and all its dependencies in a Kubernetes cluster.

## Prerequisites

To use the charts here, [Helm](https://helm.sh/) must be configured for your
Kubernetes cluster. Setting up Kubernetes and Helm is outside the scope of
this README. Please refer to the Kubernetes and Helm documentation.

- **Helm 3.0+** – This chart was tested with Helm v3.7.1, but it is also expected to work with earlier Helm versions
- **Kubernetes 1.18+** – This chart was tested with GKE v1.20.x and with [k3d](https://github.com/rancher/k3d) v1.21.x,
but it's possible it works with earlier k8s versions too
- **Istio 1.9.9+** – This chart was tested with Istio v1.9.9
- **Knative 0.18.3+, <1.x** – This chart was tested with Knative 0.18.3

It's recommended to use [turing/turing-init](https://github.com/gojek/turing/blob/main/infra/charts/turing-init/README.md) Helm chart
to configure and install Istio and Knative into the cluster, before proceeding with installation of Turing.

Configuration and installation of [turing/turing-init](https://github.com/gojek/turing/blob/main/infra/charts/turing-init/README.md)
is out of scope of this README, please refer to [turing/turing-init](https://github.com/gojek/turing/blob/main/infra/charts/turing-init/README.md)
for installation instructions.

## Installation

### Add Helm repository

```shell
$ helm repo add turing https://turing-ml.github.io/charts
```

### Installing the chart

This command will install Turing in the `default` namespace and default
chart values will be used for the installation:
```shell
$ helm install <release-name> turing/turing
```

You can (and most likely, should) override the default configuration with your own values.
Refer to [Configuration](#-configuration) section for the detailed description of available configuration keys.

### Uninstalling the chart

To uninstall my-release:
```shell
$ helm uninstall my-release
```

This will create a release of `spark-operator` in the default namespace. To install in a different one:

```shell
$ helm install -n spark my-release spark-operator/spark-operator
```

Note that `helm` will fail to install if the namespace doesn't exist. Either create the namespace beforehand or pass the `--create-namespace` flag to the `helm install` command.

## Uninstalling the chart

To uninstall `my-release`:

```shell
$ helm uninstall my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release, except for the `crds`, those will have to be removed manually.

## Configuration
---
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| dbMigrations.image.tag | string | `"v4.7.1"` | Docker tag for golang-migrate Docker image https://hub.docker.com/r/migrate/migrate |
| global.mlp.encryption | object | `{}` |  |
| global.sentry | object | `{}` |  |
| merlin.apiHost | string | `"/api/merlin/v1"` |  |
| merlin.environmentConfigs[0].cluster | string | `"dev"` |  |
| merlin.environmentConfigs[0].cpu_limit | string | `"400m"` |  |
| merlin.environmentConfigs[0].cpu_request | string | `"100m"` |  |
| merlin.environmentConfigs[0].deployment_timeout | string | `"10m"` |  |
| merlin.environmentConfigs[0].gcp_project | string | `""` |  |
| merlin.environmentConfigs[0].is_default | bool | `true` |  |
| merlin.environmentConfigs[0].is_default_prediction_job | bool | `true` |  |
| merlin.environmentConfigs[0].is_prediction_job_enabled | bool | `false` |  |
| merlin.environmentConfigs[0].max_cpu | string | `"8"` |  |
| merlin.environmentConfigs[0].max_memory | string | `"8Gi"` |  |
| merlin.environmentConfigs[0].max_replica | int | `1` |  |
| merlin.environmentConfigs[0].memory_limit | string | `"500Mi"` |  |
| merlin.environmentConfigs[0].memory_request | string | `"200Mi"` |  |
| merlin.environmentConfigs[0].min_replica | int | `0` |  |
| merlin.environmentConfigs[0].name | string | `"dev"` |  |
| merlin.environmentConfigs[0].namespace_timeout | string | `"2m"` |  |
| merlin.environmentConfigs[0].prediction_job_config.driver_cpu_request | string | `"2"` |  |
| merlin.environmentConfigs[0].prediction_job_config.driver_memory_request | string | `"2Gi"` |  |
| merlin.environmentConfigs[0].prediction_job_config.executor_cpu_request | string | `"2"` |  |
| merlin.environmentConfigs[0].prediction_job_config.executor_memory_request | string | `"2Gi"` |  |
| merlin.environmentConfigs[0].prediction_job_config.executor_replica | int | `3` |  |
| merlin.environmentConfigs[0].queue_resource_percentage | string | `"20"` |  |
| merlin.environmentConfigs[0].region | string | `""` |  |
| merlin.mlpApi.apiHost | string | `"http://{{ .Release.Name }}-mlp:8080/v1"` |  |
| merlin.postgresql.nameOverride | string | `"postgresql-merlin"` |  |
| mlp.apiHost | string | `"/api/v1"` |  |
| mlp.extraEnvs[0].name | string | `"REACT_APP_MERLIN_UI_HOMEPAGE"` |  |
| mlp.extraEnvs[0].value | string | `"/merlin"` |  |
| mlp.extraEnvs[1].name | string | `"REACT_APP_MERLIN_API"` |  |
| mlp.extraEnvs[1].value | string | `"/api/merlin/v1"` |  |
| mlp.extraEnvs[2].name | string | `"REACT_APP_TURING_UI_HOMEPAGE"` |  |
| mlp.extraEnvs[2].value | string | `"/turing"` |  |
| mlp.extraEnvs[3].name | string | `"REACT_APP_TURING_API"` |  |
| mlp.extraEnvs[3].value | string | `"/api/turing/v1"` |  |
| mlp.extraEnvs[4].name | string | `"REACT_APP_FEAST_CORE_API"` |  |
| mlp.extraEnvs[4].value | string | `"http://feast.dev/v1"` |  |
| mlp.postgresql.nameOverride | string | `"postgresql-mlp"` |  |
| postgresql.metrics.enabled | bool | `false` |  |
| postgresql.metrics.replication.applicationName | string | `"merlin"` | Replication Cluster application name. Useful for defining multiple replication policies |
| postgresql.metrics.replication.enabled | bool | `false` |  |
| postgresql.metrics.replication.numSynchronousReplicas | int | `2` | From the number of `slaveReplicas` defined above, set the number of those that will have synchronous replication NOTE: It cannot be > slaveReplicas |
| postgresql.metrics.replication.password | string | `"repl_password"` |  |
| postgresql.metrics.replication.slaveReplicas | int | `2` |  |
| postgresql.metrics.replication.synchronousCommit | string | `"on"` | Set synchronous commit mode: on, off, remote_apply, remote_write and local ref: https://www.postgresql.org/docs/9.6/runtime-config-wal.html#GUC-WAL-LEVEL |
| postgresql.metrics.replication.user | string | `"repl_user"` |  |
| postgresql.metrics.serviceMonitor.enabled | bool | `false` |  |
| postgresql.persistence.enabled | bool | `true` | Persist Postgresql data in a Persistent Volume Claim |
| postgresql.persistence.size | string | `"10Gi"` |  |
| postgresql.postgresqlDatabase | string | `"turing"` | Database name for Turing Postgresql database |
| postgresql.postgresqlPassword | string | `"turing"` | Password for Turing Postgresql database |
| postgresql.postgresqlUsername | string | `"turing"` | Username for Turing Postgresql database |
| postgresql.resources | object | `{"requests":{"cpu":"500m","memory":"256Mi"}}` | Resources requests and limits for Turing database. This should be set according to your cluster capacity and service level objectives. Reference: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| sentry.dsn | string | `""` |  |
| tags.mlp | bool | `true` |  |
| turing.clusterConfig.useInClusterConfig | bool | `false` | Configuration to use in cluster configuration or authenticate kubernetes with vault configuration |
| turing.config | object | `{"AlertConfig":{"Enabled":false},"BatchEnsemblingConfig":{"Enabled":false},"DeployConfig":{},"KubernetesLabelConfigs":{},"MLPConfig":{},"RouterDefaults":{"Image":"ghcr.io/gojek/turing/turing-router:v1.0.0-rc1"},"Sentry":{"Enabled":false}}` | Turing API server configuration.  Refer to https://github.com/gojek/turing/blob/main/api/turing/config/example.yaml |
| turing.extraArgs | list | `[]` | List of string containing additional Turing API server arguments. For example, multiple "-config" can be specified to use multiple config files |
| turing.extraContainers | list | `[]` | List of sidecar containers to attach to the Pod. For example, you can attach sidecar container that forward logs or dynamically update some  configuration files. |
| turing.extraEnvs | list | `[]` | List of extra environment variables to add to Turing API server container |
| turing.extraInitContainers | list | `[]` | List of extra initContainers to add to the Pod. For example, you need to run some init scripts to fetch credentials from a remote server |
| turing.extraVolumeMounts | list | `[]` | Extra volume mounts to attach to Turing API server container. For example to mount the extra volume containing secrets |
| turing.extraVolumes | list | `[]` | Extra volumes to attach to the Pod. For example, you can mount  additional secrets to these volumes |
| turing.image.registry | string | `"ghcr.io"` | Docker registry for Turing API image. User is required to override the registry for now as there is no publicly available Turing image |
| turing.image.repository | string | `"gojek/turing"` | Docker image repository for Turing API |
| turing.image.tag | string | `"v1.0.0-rc1"` | Docker image tag for Turing API |
| turing.ingress.class | string | `""` | Ingress class annotation to add to this Ingress rule,  useful when there are multiple ingress controllers installed |
| turing.ingress.enabled | bool | `false` | Enable ingress to provision Ingress resource for external access to Turing API |
| turing.ingress.host | string | `""` | Set host value to enable name based virtual hosting. This allows routing HTTP traffic to multiple host names at the same IP address. If no host is specified, the ingress rule applies to all inbound HTTP traffic through  the IP address specified. https://kubernetes.io/docs/concepts/services-networking/ingress/#name-based-virtual-hosting |
| turing.labels | object | `{}` |  |
| turing.livenessProbe.path | string | `"/v1/internal/live"` | HTTP path for liveness check |
| turing.readinessProbe.path | string | `"/v1/internal/ready"` | HTTP path for readiness check |
| turing.resources | object | `{}` | Resources requests and limits for Turing API. This should be set according to your cluster capacity and service level objectives. Reference: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| turing.service.externalPort | int | `8080` | Turing API Kubernetes service port number |
| turing.service.internalPort | int | `8080` | Turing API container port number |
| turing.uiConfig | object | `{"alertConfig":{},"apiConfig":{"merlinApiUrl":"/api/merlin/v1","mlpApiUrl":"/api/v1","turingApiUrl":"/api/turing/v1"},"appConfig":{"docsUrl":[{"href":"https://github.com/gojek/turing/tree/main/docs","label":"Turing User Guide"}],"scaling":{"maxAllowedReplica":20}},"authConfig":{"oauthClientId":""},"sentryConfig":{}}` | Turing UI configuration. Refer to https://github.com/gojek/turing/blob/main/ui/public/app.config.js |
