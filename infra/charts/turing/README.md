# turing

---
![Version: 0.2.16](https://img.shields.io/badge/Version-0.2.16-informational?style=flat-square)
![AppVersion: v1.0.0](https://img.shields.io/badge/AppVersion-v1.0.0-informational?style=flat-square)

Turing: ML Experimentation System

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
| global.mlp.encryption.key | string | `nil` | Global MLP Encryption Key to be used by all MLP components |
| global.sentry.dsn | string | `nil` | Global Sentry DSN value |
| merlin.environmentConfigs | list | computed value | List of Merlin environment configs, available to Turing for deploying routers By default, a new dev environment will automatically be created |
| merlin.mlpApi.apiHost | string | computed value | API endpoint to be used by Merlin to talk to MLP API |
| merlin.postgresql | object | `{"containerPorts":{"postgresql":5432},"nameOverride":"postgresql-merlin"}` | Postgresql configuration to be applied to Merlin's's postgresql database deployment Reference: https://artifacthub.io/packages/helm/bitnami/postgresql/12.1.9#parameters |
| merlin.postgresql.nameOverride | string | `"postgresql-merlin"` | Name of Merlin's Postgresql deployment |
| mlp.apiHost | string | `"/api/v1"` | MLP API endpoint, used by the MLP UI for fetching data |
| mlp.environmentConfigSecret.envKey | string | `"environment.yaml"` |  |
| mlp.environmentConfigSecret.name | string | `""` |  |
| mlp.extraEnvs | list | computed value | List of extra environment variables to add to MLP API container |
| mlp.postgresql | object | `{"containerPorts":{"postgresql":5432},"nameOverride":"postgresql-mlp"}` | Postgresql configuration to be applied to MLP's postgresql database deployment Reference: https://artifacthub.io/packages/helm/bitnami/postgresql/12.1.9#parameters |
| mlp.postgresql.nameOverride | string | `"postgresql-mlp"` | Name of MLP's Postgresql deployment |
| postgresql | object | `{"auth":{"database":"turing","password":"turing","username":"turing"},"containerPorts":{"postgresql":5432},"image":{"tag":"12.13.0"},"metrics":{"enabled":false,"serviceMonitor":{"enabled":false}},"persistence":{"enabled":true,"size":"10Gi"},"replication":{"applicationName":"turing","enabled":false,"numSynchronousReplicas":2,"password":"repl_password","slaveReplicas":2,"synchronousCommit":"on","user":"repl_user"},"resources":{"requests":{"cpu":"500m","memory":"256Mi"}}}` | Postgresql configuration to be applied to Turing's postgresql database deployment Reference: https://artifacthub.io/packages/helm/bitnami/postgresql/12.1.9#parameters |
| postgresql.auth.password | string | `"turing"` | Password for Turing Postgresql database |
| postgresql.persistence.enabled | bool | `true` | Persist Postgresql data in a Persistent Volume Claim |
| postgresql.resources | object | `{"requests":{"cpu":"500m","memory":"256Mi"}}` | Resources requests and limits for Turing database. This should be set according to your cluster capacity and service level objectives. Reference: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| sentry.dsn | string | `""` | Sentry DSN value used by both Turing API and Turing UI |
| tags.db | bool | `true` | Specifies if Postgresql database needs to be installed together with Turing |
| tags.mlp | bool | `true` | Specifies if the necessary MLP components needs to be installed together with Turing |
| turing.clusterConfig.ensemblingServiceK8sConfig.cluster | object | `{}` |  |
| turing.clusterConfig.ensemblingServiceK8sConfig.name | string | `"dev-cluster"` |  |
| turing.clusterConfig.ensemblingServiceK8sConfig.user | object | `{}` |  |
| turing.clusterConfig.environmentConfigPath | string | `"environments.yaml"` |  |
| turing.clusterConfig.useInClusterConfig | bool | `false` | Configuration to tell Turing API how it should authenticate with deployment k8s cluster By default, Turing API expects to use a remote k8s cluster for deployment and to do so, it requires cluster credentials to be stored in Vault's KV Secrets store. TODO: Change comments, no more vault |
| turing.config | object | computed value | Turing API server configuration. Please refer to https://github.com/caraml-dev/turing/blob/main/api/turing/config/example.yaml for the detailed explanation on Turing API config options |
| turing.environmentConfigs[0].k8s_config.cluster | object | `{}` |  |
| turing.environmentConfigs[0].k8s_config.name | string | `"dev-cluster"` |  |
| turing.environmentConfigs[0].k8s_config.user | object | `{}` |  |
| turing.environmentConfigs[0].name | string | `"dev"` |  |
| turing.experimentEngines | list | `[]` | Turing Experiment Engines configuration |
| turing.extraArgs | list | `[]` | List of string containing additional Turing API server arguments. For example, multiple "-config" can be specified to use multiple config files |
| turing.extraContainers | list | `[]` | List of sidecar containers to attach to the Pod. For example, you can attach sidecar container that forward logs or dynamically update some  configuration files. |
| turing.extraEnvs | list | `[]` | List of extra environment variables to add to Turing API server container |
| turing.extraInitContainers | list | `[]` | List of extra initContainers to add to the Pod. For example, you need to run some init scripts to fetch credentials from a remote server |
| turing.extraVolumeMounts | list | `[]` | Extra volume mounts to attach to Turing API server container. For example to mount the extra volume containing secrets |
| turing.extraVolumes | list | `[]` | Extra volumes to attach to the Pod. For example, you can mount  additional secrets to these volumes |
| turing.image.registry | string | `"ghcr.io"` | Docker registry for Turing API image. User is required to override the registry for now as there is no publicly available Turing image |
| turing.image.repository | string | `"caraml-dev/turing"` | Docker image repository for Turing API |
| turing.image.tag | string | `"v1.7.1"` | Docker image tag for Turing API |
| turing.ingress.class | string | `""` | Ingress class annotation to add to this Ingress rule,  useful when there are multiple ingress controllers installed |
| turing.ingress.enabled | bool | `false` | Enable ingress to provision Ingress resource for external access to Turing API |
| turing.ingress.host | string | `""` | Set host value to enable name based virtual hosting. This allows routing HTTP traffic to multiple host names at the same IP address. If no host is specified, the ingress rule applies to all inbound HTTP traffic through  the IP address specified. https://kubernetes.io/docs/concepts/services-networking/ingress/#name-based-virtual-hosting |
| turing.ingress.useV1Beta1 | bool | `false` | Whether to use networking.k8s.io/v1 (k8s version >= 1.19) or networking.k8s.io/v1beta1 (1.16 >= k8s version >= 1.22) |
| turing.labels | object | `{}` |  |
| turing.livenessProbe.path | string | `"/v1/internal/live"` | HTTP path for liveness check |
| turing.openApiSpecOverrides | object | `{}` | Override OpenAPI spec as long as it follows the OAS3 specifications. A common use for this is to set the enums of the ExperimentEngineType. See api/api/override-sample.yaml for an example. |
| turing.readinessProbe.path | string | `"/v1/internal/ready"` | HTTP path for readiness check |
| turing.resources | object | `{}` | Resources requests and limits for Turing API. This should be set according to your cluster capacity and service level objectives. Reference: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| turing.service.externalPort | int | `8080` | Turing API Kubernetes service port number |
| turing.service.internalPort | int | `8080` | Turing API container port number |
| turing.uiConfig | object | computed value | Turing UI configuration. Please Refer to https://github.com/caraml-dev/turing/blob/main/ui/public/app.config.js for the detailed explanation on Turing UI config options |
