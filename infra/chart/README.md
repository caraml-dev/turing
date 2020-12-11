# turing

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square)

Turing: ML Experimentation System

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| dbMigrations.image.tag | string | `"v4.7.1"` | Docker tag for golang-migrate Docker image https://hub.docker.com/r/migrate/migrate |
| postgresql.persistence.enabled | bool | `true` | Persist Postgresql data in a Persistent Volume Claim  |
| postgresql.postgresqlDatabase | string | `"turing"` | Database name for Turing Postgresql database |
| postgresql.postgresqlPassword | string | `nil` | Password for Turing Postgresql database (required) |
| postgresql.postgresqlUsername | string | `"turing"` | Username for Turing Postgresql database |
| postgresql.resources | object | `{}` | Resources requests and limits for Turing database. This should be set  according to your cluster capacity and service level objectives. Reference: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| swaggerUi.apiHost | string | `"127.0.0.1"` | Host for the Swagger UI |
| swaggerUi.basePath | string | `"/v1"` | Base URL path to serve the Swagger UI |
| swaggerUi.image | object | `{"tag":"v3.24.3"}` | Docker tag for Swagger UI https://hub.docker.com/r/swaggerapi/swagger-ui |
| swaggerUi.service.externalPort | int | `8080` | Swagger UI Kubernetes service port number |
| swaggerUi.service.internalPort | int | `8081` | Swagger UI container port number |
| turing.config | object | `{}` | Turing API server configuration.  Refer to https://github.com/gojek/turing/blob/main/api/turing/config/example.yaml |
| turing.extraArgs | list | `[]` | List of string containing additional Turing API server arguments. For example, multiple "-config" can be specified to use multiple config files |
| turing.extraContainers | list | `[]` | List of sidecar containers to attach to the Pod. For example, you can attach sidecar container that forward logs or dynamically update some  configuration files. |
| turing.extraEnvs | list | `[]` | List of extra environment variables to add to Turing API server container |
| turing.extraInitContainers | list | `[]` | List of extra initContainers to add to the Pod. For example, you need to run some init scripts to fetch credentials from a remote server |
| turing.extraVolumeMounts | list | `[]` | Extra volume mounts to attach to Turing API server container. For example to mount the extra volume containing secrets |
| turing.extraVolumes | list | `[]` | Extra volumes to attach to the Pod. For example, you can mount  additional secrets to these volumes |
| turing.image.registry | string | `"docker.io/"` | Docker registry for Turing API image. User is required to override the registry for now as there is no publicly available Turing image |
| turing.image.repository | string | `"turing"` | Docker image repository for Turing API |
| turing.image.tag | string | `"latest"` | Docker image tag for Turing API |
| turing.ingress.class | string | `""` | Ingress class annotation to add to this Ingress rule,  useful when there are multiple ingress controllers installed |
| turing.ingress.enabled | bool | `false` | Enable ingress to provision Ingress resource for external access to Turing API |
| turing.ingress.host | string | `""` | Set host value to enable name based virtual hosting. This allows routing HTTP traffic to multiple host names at the same IP address. If no host is specified, the ingress rule applies to all inbound HTTP traffic through  the IP address specified. https://kubernetes.io/docs/concepts/services-networking/ingress/#name-based-virtual-hosting |
| turing.livenessProbe.path | string | `"/v1/internal/live"` | HTTP path for liveness check |
| turing.readinessProbe.path | string | `"/v1/internal/ready"` | HTTP path for readiness check |
| turing.resources | object | `{}` | Resources requests and limits for Turing API. This should be set  according to your cluster capacity and service level objectives. Reference: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| turing.service.externalPort | int | `8080` | Turing API Kubernetes service port number |
| turing.service.internalPort | int | `8080` | Turing API container port number |

