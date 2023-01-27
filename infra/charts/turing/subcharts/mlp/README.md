# mlp

---
![Version: 1.0.1](https://img.shields.io/badge/Version-1.0.1-informational?style=flat-square)
![AppVersion: v1.14.6](https://img.shields.io/badge/AppVersion-v1.14.6-informational?style=flat-square)

MLP API

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
| apiHost | string | `"http://mlp/v1"` |  |
| authorization.enabled | bool | `false` |  |
| authorization.serverUrl | string | `"http://mlp-authorization-keto"` |  |
| dbMigrations.image.tag | string | `"v4.7.1"` |  |
| docs[0].href | string | `"https://github.com/gojek/merlin/blob/main/docs/getting-started/README.md"` |  |
| docs[0].label | string | `"Merlin User Guide"` |  |
| docs[1].href | string | `"https://github.com/caraml-dev/turing"` |  |
| docs[1].label | string | `"Turing User Guide"` |  |
| docs[2].href | string | `"https://docs.feast.dev/user-guide/overview"` |  |
| docs[2].label | string | `"Feast User Guide"` |  |
| encryption.key | string | `""` |  |
| environment | string | `"production"` |  |
| extraEnvs | list | `[]` | List of extra environment variables to add to MLP API server container |
| gitlab.clientId | string | `nil` |  |
| gitlab.clientSecret | string | `nil` |  |
| gitlab.enabled | bool | `false` |  |
| gitlab.host | string | `"https://gitlab.com"` |  |
| gitlab.oauthScopes | string | `"read_user"` |  |
| gitlab.redirectUrl | string | `"http://mlp/settings/connected-accounts"` |  |
| global.mlp | object | `{}` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.registry | string | `"ghcr.io"` |  |
| image.repository | string | `"gojek/mlp"` |  |
| image.tag | string | `"v1.4.16"` |  |
| ingress.enabled | bool | `false` |  |
| ingress.useV1Beta1 | bool | `false` | Whether to use networking.k8s.io/v1 (k8s version >= 1.19) or networking.k8s.io/v1beta1 (1.16 >= k8s version >= 1.22) |
| livenessProbe.path | string | `"/v1/internal/live"` |  |
| mlflowTrackingUrl | string | `"http://mlflow.mlp"` |  |
| oauthClientId | string | `""` |  |
| postgresql.auth.database | string | `"mlp"` |  |
| postgresql.auth.password | string | `"mlp"` |  |
| postgresql.auth.username | string | `"mlp"` |  |
| postgresql.image.tag | string | `"12.13.0"` |  |
| postgresql.metrics.enabled | bool | `false` |  |
| postgresql.metrics.serviceMonitor.enabled | bool | `false` |  |
| postgresql.persistence.enabled | bool | `true` | Persist Postgresql data in a Persistent Volume Claim |
| postgresql.persistence.size | string | `"10Gi"` |  |
| postgresql.replication.applicationName | string | `"mlp"` | Replication Cluster application name. Useful for defining multiple replication policies |
| postgresql.replication.enabled | bool | `false` |  |
| postgresql.replication.numSynchronousReplicas | int | `2` | From the number of `slaveReplicas` defined above, set the number of those that will have synchronous replication NOTE: It cannot be > slaveReplicas |
| postgresql.replication.password | string | `"repl_password"` |  |
| postgresql.replication.slaveReplicas | int | `2` |  |
| postgresql.replication.synchronousCommit | string | `"on"` | Set synchronous commit mode: on, off, remote_apply, remote_write and local ref: https://www.postgresql.org/docs/9.6/runtime-config-wal.html#GUC-WAL-LEVEL |
| postgresql.replication.user | string | `"repl_user"` |  |
| postgresql.resources | object | `{"requests":{"cpu":"500m","memory":"256Mi"}}` | Configure resource requests and limits ref: http://kubernetes.io/docs/user-guide/compute-resources/ |
| readinessProbe.path | string | `"/v1/internal/ready"` |  |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| service.externalPort | int | `8080` |  |
| service.internalPort | int | `8080` |  |
| streams[0] | string | `"promotion-marketing"` |  |
| streams[1] | string | `"operation-strategy"` |  |
| streams[2] | string | `"business-analyst"` |  |
| teams[0] | string | `"marketing"` |  |
| teams[1] | string | `"operation"` |  |
| teams[2] | string | `"business"` |  |
